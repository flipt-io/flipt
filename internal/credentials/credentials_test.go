package credentials

import (
	"testing"

	"github.com/go-git/go-git/v6/plumbing/transport"
	githttp "github.com/go-git/go-git/v6/plumbing/transport/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap"
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
				assert.Equal(t, "oauth2", basicAuth.Username)
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

func TestCredential_APIAuthentication(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name      string
		config    *config.CredentialConfig
		checkAuth func(*testing.T, *APIAuth)
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
		})
	}
}
