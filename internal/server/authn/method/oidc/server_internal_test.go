package oidc

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/internal/storage/authn/memory"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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
			serve(t, http.StatusOK, map[string]string{
				"issuer":                 srv.URL,
				"authorization_endpoint": srv.URL + "/auth",
				"token_endpoint":         srv.URL + "/token",
				"jwks_uri":               srv.URL + "/certs",
			})(w, r)
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
			serve(t, http.StatusOK, map[string]string{
				"issuer":                 srv.URL,
				"authorization_endpoint": srv.URL + "/auth",
				"token_endpoint":         srv.URL + "/token",
				"jwks_uri":               srv.URL + "/certs",
				"end_session_endpoint":   srv.URL + "/logout",
			})(w, r)
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

type oauthChallengeErrorStore struct {
	authn.Store
}

func (s *oauthChallengeErrorStore) PutOAuthChallenge(_ context.Context, _, _ string, _ *timestamppb.Timestamp) error {
	return fmt.Errorf("store unavailable")
}

func (s *oauthChallengeErrorStore) PopOAuthChallenge(_ context.Context, _ string) (string, error) {
	return "", fmt.Errorf("pop failed")
}

func TestServer_AuthorizeURL(t *testing.T) {
	ctx := t.Context()
	logger := zaptest.NewLogger(t)

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			serve(t, http.StatusOK, map[string]string{
				"issuer":                 srv.URL,
				"authorization_endpoint": srv.URL + "/auth",
				"token_endpoint":         srv.URL + "/token",
				"jwks_uri":               srv.URL + "/certs",
			})(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	makeCfg := func(extra ...func(*config.AuthenticationMethodOIDCProvider)) config.AuthenticationConfig {
		cfg := config.AuthenticationConfig{
			Methods: config.AuthenticationMethodsConfig{
				OIDC: config.AuthenticationMethod[config.AuthenticationMethodOIDCConfig]{
					Enabled: true,
					Method: config.AuthenticationMethodOIDCConfig{
						Providers: map[string]config.AuthenticationMethodOIDCProvider{
							"test": {
								IssuerURL:       srv.URL,
								ClientID:        "test-client",
								ClientSecret:    "test-secret",
								RedirectAddress: "http://localhost",
							},
						},
					},
				},
			},
		}
		for _, fn := range extra {
			p := cfg.Methods.OIDC.Method.Providers["test"]
			fn(&p)
			cfg.Methods.OIDC.Method.Providers["test"] = p
		}
		return cfg
	}

	tests := []struct {
		name     string
		server   *Server
		req      *auth.AuthorizeURLRequest
		wantErr  string
		wantCode codes.Code
	}{
		{
			name: "provider not found",
			server: NewServer(logger, memory.NewStore(logger),
				NewRegistry(config.AuthenticationConfig{}),
				config.AuthenticationConfig{}),
			req: &auth.AuthorizeURLRequest{
				Provider: "nonexistent",
				State:    "test-state",
			},
			wantCode: codes.InvalidArgument,
			wantErr:  "authorize: unknown provider",
		},
		{
			name: "empty challenge",
			server: NewServer(
				logger, memory.NewStore(logger),
				NewRegistry(makeCfg()),
				makeCfg(),
				WithNonceGenerator(func() string { return "" }),
			),
			req: &auth.AuthorizeURLRequest{
				Provider: "test",
				State:    "test-state",
			},
			wantCode: codes.InvalidArgument,
			wantErr:  "authorize: failed to generate challenge",
		},
		{
			name: "store error",
			server: NewServer(logger, &oauthChallengeErrorStore{Store: memory.NewStore(logger)},
				NewRegistry(makeCfg()),
				makeCfg()),
			req: &auth.AuthorizeURLRequest{
				Provider: "test",
				State:    "test-state",
			},
			wantCode: codes.Internal,
			wantErr:  "authorize: failed to persist challenge",
		},
		{
			name: "success",
			server: NewServer(
				logger, memory.NewStore(logger),
				NewRegistry(makeCfg()),
				makeCfg(),
				WithNonceGenerator(func() string { return "static-value" }),
			),
			req: &auth.AuthorizeURLRequest{
				Provider: "test",
				State:    "test-state",
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.server.AuthorizeURL(ctx, tt.req)
			if tt.wantErr != "" {
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
				assert.Equal(t, tt.wantErr, st.Message())
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Contains(t, resp.AuthorizeUrl, "/auth?")
			assert.Contains(t, resp.AuthorizeUrl, "state=test-state")
		})
	}
}

func TestServer_Callback(t *testing.T) {
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name     string
		ctx      context.Context
		req      *auth.CallbackRequest
		server   *Server
		wantErr  string
		wantCode codes.Code
	}{
		{
			name: "missing state",
			ctx:  t.Context(),
			req:  &auth.CallbackRequest{State: ""},
			server: NewServer(logger, memory.NewStore(logger),
				NewRegistry(config.AuthenticationConfig{}),
				config.AuthenticationConfig{}),
			wantCode: codes.Unauthenticated,
			wantErr:  "callback: failed to get state",
		},
		{
			name: "missing metadata",
			ctx:  t.Context(),
			req:  &auth.CallbackRequest{State: "test-state"},
			server: NewServer(logger, memory.NewStore(logger),
				NewRegistry(config.AuthenticationConfig{}),
				config.AuthenticationConfig{}),
			wantCode: codes.Unauthenticated,
			wantErr:  "callback: failed to get state",
		},
		{
			name: "missing client state",
			ctx:  metadata.NewIncomingContext(t.Context(), metadata.Pairs("some-other-key", "value")),
			req:  &auth.CallbackRequest{State: "test-state"},
			server: NewServer(logger, memory.NewStore(logger),
				NewRegistry(config.AuthenticationConfig{}),
				config.AuthenticationConfig{}),
			wantCode: codes.Unauthenticated,
			wantErr:  "callback: failed to get state",
		},
		{
			name: "state mismatch",
			ctx:  metadata.NewIncomingContext(t.Context(), metadata.Pairs("flipt_client_state", "wrong-state")),
			req:  &auth.CallbackRequest{State: "test-state"},
			server: NewServer(logger, memory.NewStore(logger),
				NewRegistry(config.AuthenticationConfig{}),
				config.AuthenticationConfig{}),
			wantCode: codes.Unauthenticated,
			wantErr:  "callback: failed to get state",
		},
		{
			name: "provider not found",
			ctx:  metadata.NewIncomingContext(t.Context(), metadata.Pairs("flipt_client_state", "test-state")),
			req:  &auth.CallbackRequest{Provider: "nonexistent", State: "test-state"},
			server: NewServer(logger, memory.NewStore(logger),
				NewRegistry(config.AuthenticationConfig{}),
				config.AuthenticationConfig{}),
			wantCode: codes.InvalidArgument,
			wantErr:  "callback: unknown provider",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.server.Callback(tt.ctx, tt.req)
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tt.wantCode, st.Code())
			assert.Equal(t, tt.wantErr, st.Message())
		})
	}
}

func TestServer_Callback_ChallengeErrors(t *testing.T) {
	ctx := metadata.NewIncomingContext(t.Context(), metadata.Pairs("flipt_client_state", "test-state"))
	logger := zaptest.NewLogger(t)

	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			serve(t, http.StatusOK, map[string]string{
				"issuer":                 srv.URL,
				"authorization_endpoint": srv.URL + "/auth",
				"token_endpoint":         srv.URL + "/token",
				"jwks_uri":               srv.URL + "/certs",
			})(w, r)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	makeCfg := func() config.AuthenticationConfig {
		return config.AuthenticationConfig{
			Methods: config.AuthenticationMethodsConfig{
				OIDC: config.AuthenticationMethod[config.AuthenticationMethodOIDCConfig]{
					Enabled: true,
					Method: config.AuthenticationMethodOIDCConfig{
						Providers: map[string]config.AuthenticationMethodOIDCProvider{
							"test": {
								IssuerURL:       srv.URL,
								ClientID:        "test-client",
								ClientSecret:    "test-secret",
								RedirectAddress: "http://localhost",
							},
						},
					},
				},
			},
		}
	}

	tests := []struct {
		name     string
		server   *Server
		wantCode codes.Code
		wantErr  string
	}{
		{
			name: "oauth challenge not found",
			server: NewServer(logger, memory.NewStore(logger),
				NewRegistry(makeCfg()),
				makeCfg()),
			wantCode: codes.InvalidArgument,
			wantErr:  "callback: missing challenge",
		},
		{
			name: "oauth challenge pop error",
			server: NewServer(logger, &oauthChallengeErrorStore{Store: memory.NewStore(logger)},
				NewRegistry(makeCfg()),
				makeCfg()),
			wantCode: codes.Unknown,
			wantErr:  "callback: getting challenge",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.server.Callback(ctx, &auth.CallbackRequest{
				Provider: "test",
				State:    "test-state",
				Code:     "auth-code",
			})
			st, ok := status.FromError(err)
			require.True(t, ok)
			assert.Equal(t, tt.wantCode, st.Code())
			assert.Equal(t, tt.wantErr, st.Message())
		})
	}
}

func TestServer_Revoke(t *testing.T) {
	ctx := t.Context()
	logger := zaptest.NewLogger(t)

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
			jwks := jose.JSONWebKeySet{
				Keys: []jose.JSONWebKey{
					{Key: &priv.PublicKey, Algorithm: "RS256", KeyID: "test-key"},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(jwks)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)

	makeCfg := func() config.AuthenticationConfig {
		return config.AuthenticationConfig{
			Methods: config.AuthenticationMethodsConfig{
				OIDC: config.AuthenticationMethod[config.AuthenticationMethodOIDCConfig]{
					Enabled: true,
					Method: config.AuthenticationMethodOIDCConfig{
						Providers: map[string]config.AuthenticationMethodOIDCProvider{
							"test": {
								IssuerURL:       srv.URL,
								ClientID:        "test-client",
								ClientSecret:    "test-secret",
								RedirectAddress: "http://localhost",
							},
						},
					},
				},
			},
		}
	}

	now := time.Now()

	noSubNoSIDToken, err := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
		"iss": srv.URL,
		"aud": "test-client",
		"exp": now.Add(time.Hour).Unix(),
		"iat": now.Unix(),
		"jti": "logout-test-no-sid",
		"events": map[string]any{
			"http://schemas.openid.net/event/backchannel-logout": struct{}{},
		},
	}).SignedString(priv)
	require.NoError(t, err)

	tests := []struct {
		name   string
		server *Server
		req    *auth.RevokeOIDCRequest
		check  func(t *testing.T, resp *auth.RevokeOIDCResponse, err error)
	}{
		{
			name: "unknown provider",
			server: NewServer(logger, memory.NewStore(logger),
				NewRegistry(config.AuthenticationConfig{}),
				config.AuthenticationConfig{}),
			req: &auth.RevokeOIDCRequest{Provider: "nonexistent"},
			check: func(t *testing.T, _ *auth.RevokeOIDCResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, codes.InvalidArgument, st.Code())
				assert.Equal(t, "revoke: unknown provider", st.Message())
			},
		},
		{
			name: "get provider fails",
			server: NewServer(logger, memory.NewStore(logger),
				NewRegistry(config.AuthenticationConfig{
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
				}),
				config.AuthenticationConfig{}),
			req: &auth.RevokeOIDCRequest{Provider: "bad"},
			check: func(t *testing.T, _ *auth.RevokeOIDCResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, codes.Internal, st.Code())
				assert.Equal(t, "revoke: failed get provider", st.Message())
			},
		},
		{
			name: "invalid logout token",
			server: NewServer(logger, memory.NewStore(logger),
				NewRegistry(makeCfg()),
				makeCfg()),
			req: &auth.RevokeOIDCRequest{
				Provider:    "test",
				LogoutToken: "not-a-jwt",
			},
			check: func(t *testing.T, _ *auth.RevokeOIDCResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, codes.InvalidArgument, st.Code())
				assert.Equal(t, "revoke: invalid logout token", st.Message())
			},
		},
		{
			name: "logout token missing sid and sub",
			server: NewServer(logger, memory.NewStore(logger),
				NewRegistry(makeCfg()),
				makeCfg()),
			req: &auth.RevokeOIDCRequest{
				Provider:    "test",
				LogoutToken: noSubNoSIDToken,
			},
			check: func(t *testing.T, _ *auth.RevokeOIDCResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				assert.Equal(t, codes.InvalidArgument, st.Code())
				assert.Equal(t, "revoke: invalid logout token", st.Message())
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := tt.server.Revoke(ctx, tt.req)
			tt.check(t, resp, err)
		})
	}
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
