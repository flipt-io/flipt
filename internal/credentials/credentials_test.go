package credentials

import (
	"context"
	"testing"

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
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want.config, got.config)
		})
	}
}

func TestCredential_HTTPClient(t *testing.T) {
	logger := zap.NewNop()
	ctx := context.Background()

	tests := []struct {
		name    string
		config  *config.CredentialConfig
		wantErr bool
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
		},
		{
			name: "ssh not supported for http",
			config: &config.CredentialConfig{
				Type: config.CredentialTypeSSH,
				SSH: &config.SSHAuthConfig{
					User:     "git",
					Password: "pass",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Credential{
				logger: logger,
				config: tt.config,
			}

			client, err := c.HTTPClient(ctx)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, client)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, client)
		})
	}
}

func TestCredential_GitAuthentication(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name    string
		config  *config.CredentialConfig
		wantErr bool
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
				assert.Error(t, err)
				assert.Nil(t, auth)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, auth)
		})
	}
}
