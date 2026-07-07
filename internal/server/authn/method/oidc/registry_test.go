package oidc

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"golang.org/x/oauth2"
)

func TestCallbackURL(t *testing.T) {
	tests := []struct {
		name string
		host string
		want string
	}{
		{
			name: "plain",
			host: "localhost",
			want: "localhost/auth/v1/method/oidc/foo/callback",
		},
		{
			name: "no trailing slash",
			host: "localhost:8080",
			want: "localhost:8080/auth/v1/method/oidc/foo/callback",
		},
		{
			name: "with trailing slash",
			host: "localhost:8080/",
			want: "localhost:8080/auth/v1/method/oidc/foo/callback",
		},
		{
			name: "with protocol",
			host: "http://localhost:8080",
			want: "http://localhost:8080/auth/v1/method/oidc/foo/callback",
		},
		{
			name: "with protocol and trailing slash",
			host: "http://localhost:8080/",
			want: "http://localhost:8080/auth/v1/method/oidc/foo/callback",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := callbackURL(tt.host, "foo")
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestTk_String(t *testing.T) {
	t.Run("nil receiver", func(t *testing.T) {
		var tk *Tk
		assert.Equal(t, "<nil>", tk.String())
	})

	t.Run("populated token", func(t *testing.T) {
		tk := &Tk{
			idToken:     &oidc.IDToken{},
			rawIDToken:  "raw",
			oauth2Token: &oauth2.Token{AccessToken: "secret"},
		}
		assert.Equal(t, "Tk{idToken:<redacted>, rawIDToken:<redacted>, oauth2Token:<redacted>}", tk.String())
	})
}

func TestRegistry_GetProvider_NotFound(t *testing.T) {
	reg := NewRegistry(config.AuthenticationConfig{})
	_, err := reg.getProvider(t.Context(), "nonexistent")
	require.Error(t, err)
	assert.Equal(t, `no oidc provider "nonexistent" not found`, err.Error())
}

func TestRegistry_GetProvider_ClientCreationFailure(t *testing.T) {
	reg := NewRegistry(config.AuthenticationConfig{
		Methods: config.AuthenticationMethodsConfig{
			OIDC: config.AuthenticationMethod[config.AuthenticationMethodOIDCConfig]{
				Enabled: true,
				Method: config.AuthenticationMethodOIDCConfig{
					Providers: map[string]config.AuthenticationMethodOIDCProvider{
						"bad": {
							IssuerURL:       "http://127.0.0.1:1",
							ClientID:        "id",
							ClientSecret:    "secret",
							RedirectAddress: "http://localhost",
						},
					},
				},
			},
		},
	})
	_, err := reg.getProvider(t.Context(), "bad")
	require.ErrorContains(t, err, "creating OIDC provider")
}

func TestClient_UsePKCE(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		c := &client{
			cfg: config.AuthenticationMethodOIDCProvider{
				UsePKCE: true,
			},
		}
		assert.True(t, c.UsePKCE())
	})

	t.Run("disabled", func(t *testing.T) {
		c := &client{
			cfg: config.AuthenticationMethodOIDCProvider{
				UsePKCE: false,
			},
		}
		assert.False(t, c.UsePKCE())
	})
}

func TestClient_AuthURL(t *testing.T) {
	tests := []struct {
		name      string
		client    *client
		state     string
		challenge string
		nonce     string
		wantErr   string
		check     func(t *testing.T, u *url.URL)
	}{
		{
			name: "basic",
			client: &client{
				oauth2Cfg: &oauth2.Config{
					ClientID: "my-client",
					Endpoint: oauth2.Endpoint{
						AuthURL: "http://example.com/auth",
					},
				},
			},
			state: "my-state",
			nonce: "my-nonce",
			check: func(t *testing.T, u *url.URL) {
				assert.Equal(t, "my-state", u.Query().Get("state"))
				assert.Equal(t, "my-nonce", u.Query().Get("nonce"))
				assert.Equal(t, "my-client", u.Query().Get("client_id"))
				assert.Equal(t, "code", u.Query().Get("response_type"))
				assert.Empty(t, u.Query().Get("code_challenge"))
			},
		},
		{
			name: "with PKCE",
			client: &client{
				oauth2Cfg: &oauth2.Config{
					ClientID: "my-client",
					Endpoint: oauth2.Endpoint{
						AuthURL: "http://example.com/auth",
					},
				},
				cfg: config.AuthenticationMethodOIDCProvider{
					UsePKCE: true,
				},
			},
			state:     "my-state",
			challenge: "my-challenge",
			nonce:     "my-nonce",
			check: func(t *testing.T, u *url.URL) {
				assert.Equal(t, "my-state", u.Query().Get("state"))
				assert.Equal(t, "my-nonce", u.Query().Get("nonce"))
				assert.NotEmpty(t, u.Query().Get("code_challenge"))
				assert.Equal(t, "S256", u.Query().Get("code_challenge_method"))
			},
		},
		{
			name: "with authorize parameters",
			client: &client{
				oauth2Cfg: &oauth2.Config{
					ClientID: "my-client",
					Endpoint: oauth2.Endpoint{
						AuthURL: "http://example.com/auth",
					},
				},
				cfg: config.AuthenticationMethodOIDCProvider{
					AuthorizeParameters: map[string]string{
						"audience": "my-api",
						"prompt":   "login",
					},
				},
			},
			state: "my-state",
			nonce: "my-nonce",
			check: func(t *testing.T, u *url.URL) {
				assert.Equal(t, "my-api", u.Query().Get("audience"))
				assert.Equal(t, "login", u.Query().Get("prompt"))
				assert.Equal(t, "my-state", u.Query().Get("state"))
			},
		},
		{
			name: "url parse error",
			client: &client{
				oauth2Cfg: &oauth2.Config{
					ClientID: "my-client",
					Endpoint: oauth2.Endpoint{
						AuthURL: "http://host:-1/auth",
					},
				},
				cfg: config.AuthenticationMethodOIDCProvider{
					AuthorizeParameters: map[string]string{"foo": "bar"},
				},
			},
			state:   "my-state",
			nonce:   "my-nonce",
			wantErr: "invalid port",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.client.AuthURL(tt.state, tt.challenge, tt.nonce)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			u, err := url.Parse(got)
			require.NoError(t, err)
			tt.check(t, u)
		})
	}
}

func TestClient_VerifyLogout(t *testing.T) {
	ctx := t.Context()

	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			serve(t, http.StatusOK, map[string]string{
				"issuer":                 srv.URL,
				"authorization_endpoint": srv.URL + "/auth",
				"token_endpoint":         srv.URL + "/token",
				"jwks_uri":               srv.URL + "/certs",
				"end_session_endpoint":   srv.URL + "/logout",
			})(w, r)
		case "/certs":
			pubKey := &priv.PublicKey
			jwks := jose.JSONWebKeySet{
				Keys: []jose.JSONWebKey{
					{Key: pubKey, Algorithm: "RS256", KeyID: "test-key"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(jwks)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	provider, err := oidc.NewProvider(ctx, srv.URL)
	require.NoError(t, err)

	c := &client{
		provider:  provider,
		oidcCfg:   &oidc.Config{ClientID: "test-client"},
		createdAt: time.Now(),
	}

	now := time.Now()
	validToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": srv.URL,
		"aud": "test-client",
		"exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(),
		"jti": "logout-test",
		"sub": "test-user",
		"events": map[string]any{
			"http://schemas.openid.net/event/backchannel-logout": struct{}{},
		},
	}).SignedString(priv)
	require.NoError(t, err)

	tests := []struct {
		name    string
		token   string
		wantErr string
	}{
		{
			name:    "invalid JWT",
			token:   "not-a-jwt",
			wantErr: "failed verify raw logout token",
		},
		{
			name:  "valid logout token",
			token: validToken,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := c.VerifyLogout(ctx, tt.token)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)
			assert.Equal(t, "test-user", got.Subject)
		})
	}
}

func TestClient_UserInfo(t *testing.T) {
	ctx := t.Context()

	var (
		srv             *httptest.Server
		userInfoHandler func(w http.ResponseWriter, r *http.Request)
	)

	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			serve(t, http.StatusOK, map[string]string{
				"issuer":                 srv.URL,
				"authorization_endpoint": srv.URL + "/auth",
				"token_endpoint":         srv.URL + "/token",
				"jwks_uri":               srv.URL + "/certs",
				"userinfo_endpoint":      srv.URL + "/userinfo",
			})(w, r)
		case "/userinfo":
			if userInfoHandler != nil {
				userInfoHandler(w, r)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	provider, err := oidc.NewProvider(ctx, srv.URL)
	require.NoError(t, err)

	c := &client{
		provider:  provider,
		oidcCfg:   &oidc.Config{ClientID: "test-client"},
		createdAt: time.Now(),
	}

	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: "test-token"})

	tests := []struct {
		name         string
		tokenSource  oauth2.TokenSource
		validSubject string
		claims       any
		setUp        func()
		wantErr      string
	}{
		{
			name:    "nil token source",
			claims:  &claims{},
			wantErr: "token source is nil",
		},
		{
			name:        "nil claims",
			tokenSource: tokenSource,
			wantErr:     "claims interface is nil",
		},
		{
			name:        "claims not a pointer",
			tokenSource: tokenSource,
			claims:      "not-a-pointer",
			wantErr:     "claims must to be a pointer",
		},
		{
			name:         "user info failure",
			tokenSource:  tokenSource,
			validSubject: "test-subject",
			claims:       &claims{},
			setUp:        func() { userInfoHandler = serve(t, http.StatusInternalServerError, nil) },
			wantErr:      "request to get user info failed",
		},
		{
			name:         "invalid subject",
			tokenSource:  tokenSource,
			validSubject: "expected-subject",
			claims:       &claims{},
			setUp:        func() { userInfoHandler = serve(t, http.StatusOK, map[string]string{"sub": "different-subject"}) },
			wantErr:      "invalid subject",
		},
		{
			name:         "failed to get claims",
			tokenSource:  tokenSource,
			validSubject: "test-subject",
			claims:       new(int),
			setUp:        func() { userInfoHandler = serve(t, http.StatusOK, map[string]string{"sub": "test-subject"}) },
			wantErr:      "failed to get claims",
		},
		{
			name:         "success",
			tokenSource:  tokenSource,
			validSubject: "test-subject",
			claims:       &claims{},
			setUp:        func() { userInfoHandler = serve(t, http.StatusOK, map[string]string{"sub": "test-subject"}) },
			wantErr:      "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userInfoHandler = serve(t, http.StatusInternalServerError, nil)

			if tt.setUp != nil {
				tt.setUp()
			}

			err := c.UserInfo(ctx, tt.tokenSource, tt.validSubject, tt.claims)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)

				cl, ok := tt.claims.(*claims)
				require.True(t, ok)
				require.NotNil(t, cl.Sub)
				assert.Equal(t, "test-subject", *cl.Sub)
			}
		})
	}
}

func serve(tb testing.TB, status int, data map[string]string) func(http.ResponseWriter, *http.Request) {
	tb.Helper()
	return func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
		w.Header().Set("Content-Type", "application/json")
		if data != nil {
			assert.NoError(tb, json.NewEncoder(w).Encode(data))
		}
	}
}
