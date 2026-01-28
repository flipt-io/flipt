package credentials

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-git/go-git/v6/plumbing/transport"
	githttp "github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"golang.org/x/oauth2"
)

func TestNew(t *testing.T) {
	logger := zap.NewNop()
	configs := config.CredentialsConfig{
		"test": &config.CredentialConfig{},
	}

	cs := New(logger, configs)
	assert.NotNil(t, cs)
	assert.Equal(t, logger, cs.logger)
	assert.Equal(t, configs, cs.configs)
}

func TestCredentialSource_Get(t *testing.T) {
	logger := zap.NewNop()
	configs := config.CredentialsConfig{
		"existing": &config.CredentialConfig{
			Type: config.CredentialTypeBasic,
			Basic: &config.BasicAuthConfig{
				Username: "user",
				Password: "pass",
			},
		},
	}

	cs := New(logger, configs)

	tests := []struct {
		name    string
		credKey string
		want    *Credential
		wantErr error
	}{
		{
			name:    "existing credential",
			credKey: "existing",
			want: &Credential{
				logger: logger,
				config: configs["existing"],
			},
			wantErr: nil,
		},
		{
			name:    "non-existing credential",
			credKey: "nonexisting",
			want:    nil,
			wantErr: errors.ErrNotFoundf("credential %q", "nonexisting"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cs.Get(tt.credKey)
			if tt.wantErr != nil {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.config, got.config)
		})
	}
}

func TestCredential_GitAuthentication(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name      string
		config    *config.CredentialConfig
		wantErr   bool
		checkAuth func(*testing.T, transport.AuthMethod)
	}{
		{
			name: "basic auth",
			config: &config.CredentialConfig{
				Type: config.CredentialTypeBasic,
				Basic: &config.BasicAuthConfig{
					Username: "user",
					Password: "pass",
				},
			},
			wantErr: false,
			checkAuth: func(t *testing.T, auth transport.AuthMethod) {
				basicAuth, ok := auth.(*githttp.BasicAuth)
				require.True(t, ok, "expected BasicAuth")
				assert.Equal(t, "user", basicAuth.Username)
				assert.Equal(t, "pass", basicAuth.Password)
			},
		},
		{
			name: "access token",
			config: &config.CredentialConfig{
				Type: config.CredentialTypeAccessToken,
				AccessToken: func() *string {
					s := "token123"
					return &s
				}(),
			},
			wantErr: false,
			checkAuth: func(t *testing.T, auth transport.AuthMethod) {
				// Access tokens should be converted to BasicAuth for Git operations
				basicAuth, ok := auth.(*githttp.BasicAuth)
				require.True(t, ok, "expected BasicAuth for access token")
				assert.Equal(t, "x-token-auth", basicAuth.Username)
				assert.Equal(t, "token123", basicAuth.Password)
			},
		},
		{
			name: "ssh with private key bytes",
			config: &config.CredentialConfig{
				Type: config.CredentialTypeSSH,
				SSH: &config.SSHAuthConfig{
					User:            "git",
					Password:        "pass",
					PrivateKeyBytes: "invalid-key-for-testing",
				},
			},
			wantErr: true, // Will error due to invalid key
			checkAuth: func(t *testing.T, auth transport.AuthMethod) {
				// This test expects an error, so auth should be nil
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Credential{
				logger: logger,
				config: tt.config,
			}

			auth, err := c.GitAuthentication()
			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, auth)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, auth)

			if tt.checkAuth != nil {
				tt.checkAuth(t, auth)
			}
		})
	}
}

func TestCredential_Type(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name         string
		config       *config.CredentialConfig
		expectedType config.CredentialType
	}{
		{
			name: "access token",
			config: &config.CredentialConfig{
				Type: config.CredentialTypeAccessToken,
				AccessToken: func() *string {
					s := "token123"
					return &s
				}(),
			},
			expectedType: config.CredentialTypeAccessToken,
		},
		{
			name: "basic auth",
			config: &config.CredentialConfig{
				Type: config.CredentialTypeBasic,
				Basic: &config.BasicAuthConfig{
					Username: "user",
					Password: "pass",
				},
			},
			expectedType: config.CredentialTypeBasic,
		},
		{
			name: "ssh",
			config: &config.CredentialConfig{
				Type: config.CredentialTypeSSH,
				SSH: &config.SSHAuthConfig{
					User:           "git",
					Password:       "pass",
					PrivateKeyPath: "/path/to/key",
				},
			},
			expectedType: config.CredentialTypeSSH,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Credential{
				logger: logger,
				config: tt.config,
			}

			assert.Equal(t, tt.expectedType, c.Type())
		})
	}
}

func TestCredential_SSHUser(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name     string
		config   *config.CredentialConfig
		expected string
	}{
		{
			name: "ssh credential with user",
			config: &config.CredentialConfig{
				Type: config.CredentialTypeSSH,
				SSH: &config.SSHAuthConfig{
					User:           "git",
					Password:       "pass",
					PrivateKeyPath: "/path/to/key",
				},
			},
			expected: "git",
		},
		{
			name: "ssh credential with custom user",
			config: &config.CredentialConfig{
				Type: config.CredentialTypeSSH,
				SSH: &config.SSHAuthConfig{
					User:           "deploy",
					Password:       "pass",
					PrivateKeyPath: "/path/to/key",
				},
			},
			expected: "deploy",
		},
		{
			name: "basic auth credential returns empty",
			config: &config.CredentialConfig{
				Type: config.CredentialTypeBasic,
				Basic: &config.BasicAuthConfig{
					Username: "user",
					Password: "pass",
				},
			},
			expected: "",
		},
		{
			name: "access token credential returns empty",
			config: &config.CredentialConfig{
				Type: config.CredentialTypeAccessToken,
				AccessToken: func() *string {
					s := "token123"
					return &s
				}(),
			},
			expected: "",
		},
		{
			name: "ssh credential with nil SSH config returns empty",
			config: &config.CredentialConfig{
				Type: config.CredentialTypeSSH,
				SSH:  nil,
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Credential{
				logger: logger,
				config: tt.config,
			}

			assert.Equal(t, tt.expected, c.SSHUser())
		})
	}
}

func TestCredential_APIAuthentication(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name       string
		config     *config.CredentialConfig
		checkAuth  func(*testing.T, *APIAuth)
		checkToken func(*testing.T, oauth2.TokenSource, error)
	}{
		{
			name: "access token credential",
			config: &config.CredentialConfig{
				Type: config.CredentialTypeAccessToken,
				AccessToken: func() *string {
					s := "glpat-abc123"
					return &s
				}(),
			},
			checkAuth: func(t *testing.T, auth *APIAuth) {
				assert.Equal(t, config.CredentialTypeAccessToken, auth.Type())
				assert.Equal(t, "glpat-abc123", auth.Token)
				assert.Empty(t, auth.Username)
				assert.Empty(t, auth.Password)
			},
			checkToken: func(t *testing.T, ts oauth2.TokenSource, err error) {
				require.NoError(t, err)
				require.NotNil(t, ts)
				token, err := ts.Token()
				require.NoError(t, err)
				assert.Equal(t, "Bearer", token.Type())
				assert.Equal(t, "glpat-abc123", token.AccessToken)
			},
		},
		{
			name: "basic auth credential",
			config: &config.CredentialConfig{
				Type: config.CredentialTypeBasic,
				Basic: &config.BasicAuthConfig{
					Username: "user",
					Password: "pass",
				},
			},
			checkAuth: func(t *testing.T, auth *APIAuth) {
				assert.Equal(t, config.CredentialTypeBasic, auth.Type())
				assert.Empty(t, auth.Token)
				assert.Equal(t, "user", auth.Username)
				assert.Equal(t, "pass", auth.Password)
			},
			checkToken: func(t *testing.T, ts oauth2.TokenSource, err error) {
				require.NoError(t, err)
				require.NotNil(t, ts)
				token, err := ts.Token()
				require.NoError(t, err)
				assert.Equal(t, "Basic", token.Type())
				assert.Equal(t, "dXNlcjpwYXNz", token.AccessToken)
			},
		},
		{
			name: "ssh credential",
			config: &config.CredentialConfig{
				Type: config.CredentialTypeSSH,
				SSH: &config.SSHAuthConfig{
					User:           "git",
					Password:       "pass",
					PrivateKeyPath: "/path/to/key",
				},
			},
			checkAuth: func(t *testing.T, auth *APIAuth) {
				// SSH doesn't provide API authentication
				assert.Equal(t, config.CredentialTypeSSH, auth.Type())
				assert.Empty(t, auth.Token)
				assert.Empty(t, auth.Username)
				assert.Empty(t, auth.Password)
			},
			checkToken: func(t *testing.T, ts oauth2.TokenSource, err error) {
				require.Error(t, err)
				assert.Nil(t, ts)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Credential{
				logger: logger,
				config: tt.config,
			}

			auth := c.APIAuthentication()
			assert.NotNil(t, auth)

			if tt.checkAuth != nil {
				tt.checkAuth(t, auth)
			}

			token, err := c.APIAuthTokenSource()
			if tt.checkToken != nil {
				tt.checkToken(t, token, err)
			}
		})
	}
}

func TestGithubAppCredentials(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /app/installations/10/access_tokens", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, err := w.Write([]byte(`
	{
  "token": "ghs_MockAccessToken",
  "expires_at": "2046-01-29T14:12:34Z",
  "permissions": {
    "contents": "read",
    "metadata": "read"
  }}`))
		assert.NoError(t, err)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := &Credential{
		config: &config.CredentialConfig{
			Type: config.CredentialTypeGithubApp,
			GitHubApp: &config.GitHubAppConfig{
				ClientID:       "1234",
				InstallationID: 10,
				ApiURL:         srv.URL,
				PrivateKeyBytes: func(t *testing.T) string {
					key, err := rsa.GenerateKey(rand.Reader, 2048)
					require.NoError(t, err)
					der := x509.MarshalPKCS1PrivateKey(key)
					block := &pem.Block{
						Type:  "RSA PRIVATE KEY",
						Bytes: der,
					}
					var buf bytes.Buffer
					err = pem.Encode(&buf, block)
					require.NoError(t, err)
					return buf.String()
				}(t),
			},
		},
	}
	t.Run("auth token for SCM", func(t *testing.T) {
		ts, err := c.APIAuthTokenSource()
		require.NoError(t, err)
		require.NotNil(t, ts)
		token, err := ts.Token()
		require.NoError(t, err)
		assert.Equal(t, "Bearer", token.Type())
		assert.Equal(t, "ghs_MockAccessToken", token.AccessToken)
	})

	t.Run("auth token for go-git http client", func(t *testing.T) {
		ghau, err := newGitHubAppAuth(zaptest.NewLogger(t), c.config.GitHubApp)
		require.NoError(t, err)
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		ghau.SetAuth(r)
		assert.Equal(t, "Basic eC10b2tlbi1hdXRoOmdoc19Nb2NrQWNjZXNzVG9rZW4=", r.Header.Get("Authorization"))
	})
}
