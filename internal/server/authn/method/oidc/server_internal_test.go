package oidc

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/rpc/flipt/auth"
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

func Test_Server_SkipsAuthentication(t *testing.T) {
	server := &Server{}
	assert.True(t, server.SkipsAuthentication(t.Context()))
}

func TestEndSessionURI(t *testing.T) {
	ctx := t.Context()

	t.Run("provider not configured", func(t *testing.T) {
		authConfig := config.AuthenticationConfig{}
		auth := &auth.Authentication{
			Metadata: map[string]string{
				storageMetadataOIDCProvider: "nonexistent",
			},
		}
		_, err := EndSessionURI(ctx, NewRegistry(authConfig), auth, "id-token")
		require.ErrorContains(t, err, "no oidc provider")
	})

	t.Run("UseEndSessionEndpoint disabled", func(t *testing.T) {
		authConfig := config.AuthenticationConfig{
			Methods: config.AuthenticationMethodsConfig{
				OIDC: config.AuthenticationMethod[config.AuthenticationMethodOIDCConfig]{
					Enabled: true,
					Method: config.AuthenticationMethodOIDCConfig{
						Providers: map[string]config.AuthenticationMethodOIDCProvider{
							"google": {
								IssuerURL:             "http://example.com",
								ClientID:              "id",
								ClientSecret:          "secret",
								RedirectAddress:       "http://localhost",
								UseEndSessionEndpoint: false,
							},
						},
					},
				},
			},
		}
		auth := &auth.Authentication{
			Metadata: map[string]string{
				storageMetadataOIDCProvider: "google",
			},
		}
		uri, err := EndSessionURI(ctx, NewRegistry(authConfig), auth, "id-token")
		require.NoError(t, err)
		assert.Empty(t, uri)
	})

	t.Run("provider has no end_session_endpoint", func(t *testing.T) {
		// OIDC provider that returns a discovery doc with no end_session_endpoint
		var srv *httptest.Server
		srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := strings.TrimPrefix(srv.URL, "http://")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"issuer":                 "http://" + host,
				"authorization_endpoint": "http://" + host + "/auth",
				"token_endpoint":         "http://" + host + "/token",
				"jwks_uri":               "http://" + host + "/certs",
			})
		}))
		t.Cleanup(srv.Close)

		authConfig := config.AuthenticationConfig{
			Methods: config.AuthenticationMethodsConfig{
				OIDC: config.AuthenticationMethod[config.AuthenticationMethodOIDCConfig]{
					Enabled: true,
					Method: config.AuthenticationMethodOIDCConfig{
						Providers: map[string]config.AuthenticationMethodOIDCProvider{
							"google": {
								IssuerURL:             srv.URL,
								ClientID:              "id",
								ClientSecret:          "secret",
								RedirectAddress:       "http://localhost",
								UseEndSessionEndpoint: true,
							},
						},
					},
				},
			},
		}
		auth := &auth.Authentication{
			Metadata: map[string]string{
				storageMetadataOIDCProvider: "google",
			},
		}
		_, err := EndSessionURI(ctx, NewRegistry(authConfig), auth, "id-token")
		require.ErrorIs(t, err, errProviderWithNoEndSessionEndpoint)
	})

	t.Run("success with end_session_endpoint", func(t *testing.T) {
		var srv *httptest.Server
		srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			host := strings.TrimPrefix(srv.URL, "http://")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{
				"issuer":                 "http://" + host,
				"authorization_endpoint": "http://" + host + "/auth",
				"token_endpoint":         "http://" + host + "/token",
				"jwks_uri":               "http://" + host + "/certs",
				"end_session_endpoint":   "http://" + host + "/logout",
			})
		}))
		t.Cleanup(srv.Close)

		authConfig := config.AuthenticationConfig{
			Methods: config.AuthenticationMethodsConfig{
				OIDC: config.AuthenticationMethod[config.AuthenticationMethodOIDCConfig]{
					Enabled: true,
					Method: config.AuthenticationMethodOIDCConfig{
						Providers: map[string]config.AuthenticationMethodOIDCProvider{
							"google": {
								IssuerURL:             srv.URL,
								ClientID:              "id",
								ClientSecret:          "secret",
								RedirectAddress:       "http://localhost",
								UseEndSessionEndpoint: true,
							},
						},
					},
				},
			},
		}
		auth := &auth.Authentication{
			Metadata: map[string]string{
				storageMetadataOIDCProvider: "google",
			},
		}
		uri, err := EndSessionURI(ctx, NewRegistry(authConfig), auth, "my-id-token")
		require.NoError(t, err)
		require.NotEmpty(t, uri)

		parsed, err := url.Parse(uri)
		require.NoError(t, err)
		assert.Equal(t, "my-id-token", parsed.Query().Get("id_token_hint"))
		assert.Equal(t, "http://localhost", parsed.Query().Get("post_logout_redirect_uri"))
	})
}

func TestEncodeOAuthChallenge(t *testing.T) {
	tests := []struct {
		name      string
		challenge string
		nonce     string
		want      string
		wantErr   string
	}{
		{
			name:      "encodes challenge and nonce separated by dot",
			challenge: "challenge-val",
			nonce:     "nonce-val",
			want:      "challenge-val.nonce-val",
		},
		{
			name:      "error on empty challenge",
			challenge: "",
			nonce:     "nonce-val",
			wantErr:   "encodeOAuthChallenge: challenge and nonce must not be empty",
		},
		{
			name:      "error on empty nonce",
			challenge: "challenge-val",
			nonce:     "",
			wantErr:   "encodeOAuthChallenge: challenge and nonce must not be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := encodeOAuthChallenge(tt.challenge, tt.nonce)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDecodeOAuthChallenge(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantChallenge string
		wantNonce     string
		wantErr       string
	}{
		{
			name:          "new format: challenge and nonce",
			input:         "abc123.def456",
			wantChallenge: "abc123",
			wantNonce:     "def456",
		},
		{
			name:    "single value without separator returns error",
			input:   "singlevalue",
			wantErr: `decodeOAuthChallenge: invalid challenge data: "singlevalue"`,
		},
		{
			name:          "challenge with dots inside is split correctly",
			input:         "abc.def.ghi",
			wantChallenge: "abc",
			wantNonce:     "def.ghi",
		},
		{
			name:    "empty string returns error",
			input:   "",
			wantErr: `decodeOAuthChallenge: invalid challenge data: ""`,
		},
		{
			name:    "challenge only with trailing dot returns error",
			input:   "abc.",
			wantErr: `decodeOAuthChallenge: invalid challenge data: "abc."`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			challenge, nonce, err := decodeOAuthChallenge(tt.input)
			if tt.wantErr != "" {
				require.EqualError(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantChallenge, challenge)
			assert.Equal(t, tt.wantNonce, nonce)
		})
	}
}

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Run("encode then decode yields original values", func(t *testing.T) {
		challenge, nonce := "my-challenge", "my-nonce"
		encoded, err := encodeOAuthChallenge(challenge, nonce)
		require.NoError(t, err)
		gotChallenge, gotNonce, err := decodeOAuthChallenge(encoded)
		require.NoError(t, err)
		assert.Equal(t, challenge, gotChallenge)
		assert.Equal(t, nonce, gotNonce)
	})
}

func TestClaims_FallbackFrom(t *testing.T) {
	tests := []struct {
		name    string
		initial claims
		extra   map[string]any
		want    claims
	}{
		{
			name:    "no extra claims - unchanged",
			initial: claims{Name: new("John"), Email: new("john@example.com"), Picture: new("http://img.jpg")},
			extra:   nil,
			want:    claims{Name: new("John"), Email: new("john@example.com"), Picture: new("http://img.jpg")},
		},
		{
			name:    "extra claims but all fields already populated - unchanged",
			initial: claims{Name: new("John"), Email: new("john@example.com"), Picture: new("http://img.jpg")},
			extra:   map[string]any{"name": "Jane", "email": "jane@example.com", "picture": "http://jane.jpg"},
			want:    claims{Name: new("John"), Email: new("john@example.com"), Picture: new("http://img.jpg")},
		},
		{
			name:    "all fields nil - fallback from extra",
			initial: claims{},
			extra:   map[string]any{"name": "Jane", "email": "jane@example.com", "picture": "http://jane.jpg"},
			want:    claims{Name: new("Jane"), Email: new("jane@example.com"), Picture: new("http://jane.jpg")},
		},
		{
			name:    "some fields nil - partial fallback",
			initial: claims{Name: new("John")},
			extra:   map[string]any{"name": "Jane", "email": "jane@example.com", "picture": "http://jane.jpg"},
			want:    claims{Name: new("John"), Email: new("jane@example.com"), Picture: new("http://jane.jpg")},
		},
		{
			name:    "empty string fields - fallback from extra",
			initial: claims{Name: new(""), Email: new(""), Picture: new("")},
			extra:   map[string]any{"name": "Jane", "email": "jane@example.com", "picture": "http://jane.jpg"},
			want:    claims{Name: new("Jane"), Email: new("jane@example.com"), Picture: new("http://jane.jpg")},
		},
		{
			name:    "partial empty string - fallback only empty",
			initial: claims{Name: new("John"), Email: new(""), Picture: new("http://existing.jpg")},
			extra:   map[string]any{"name": "Jane", "email": "jane@example.com", "picture": "http://jane.jpg"},
			want:    claims{Name: new("John"), Email: new("jane@example.com"), Picture: new("http://existing.jpg")},
		},
		{
			name:    "non-string values in extra - ignored",
			initial: claims{},
			extra:   map[string]any{"name": 123, "email": true, "picture": nil},
			want:    claims{},
		},
		{
			name:    "empty string in extra - ignored",
			initial: claims{},
			extra:   map[string]any{"name": "", "email": "", "picture": ""},
			want:    claims{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.initial.fallbackFrom(tt.extra)
			assert.Equal(t, tt.want, tt.initial)
		})
	}
}
