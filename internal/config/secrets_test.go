package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretReference_Validate(t *testing.T) {
	tests := []struct {
		name    string
		ref     SecretReference
		wantErr string
	}{
		{
			name: "valid reference",
			ref: SecretReference{
				Provider: "vault",
				Path:     "secret/data/gpg",
				Key:      "private_key",
			},
			wantErr: "",
		},
		{
			name: "missing provider",
			ref: SecretReference{
				Path: "secret/data/gpg",
				Key:  "private_key",
			},
			wantErr: "secret_reference: provider non-empty value is required",
		},
		{
			name: "missing path",
			ref: SecretReference{
				Provider: "vault",
				Key:      "private_key",
			},
			wantErr: "secret_reference: path non-empty value is required",
		},
		{
			name: "missing key",
			ref: SecretReference{
				Provider: "vault",
				Path:     "secret/data/gpg",
			},
			wantErr: "secret_reference: key non-empty value is required",
		},
		{
			name:    "all fields missing",
			ref:     SecretReference{},
			wantErr: "secret_reference: provider non-empty value is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.ref.Validate()
			if tt.wantErr == "" {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Equal(t, tt.wantErr, err.Error())
			}
		})
	}
}

func TestSecretsConfig_EnabledProviders(t *testing.T) {
	tests := []struct {
		name     string
		config   SecretsConfig
		expected []string
	}{
		{
			name:     "no providers enabled",
			config:   SecretsConfig{},
			expected: nil,
		},
		{
			name: "file provider enabled",
			config: SecretsConfig{
				Providers: ProvidersConfig{
					File: &FileProviderConfig{
						Enabled:  true,
						BasePath: "/etc/secrets",
					},
				},
			},
			expected: []string{"file"},
		},
		{
			name: "vault provider enabled",
			config: SecretsConfig{
				Providers: ProvidersConfig{
					Vault: &VaultProviderConfig{
						Enabled: true,
						Address: "https://vault.example.com",
					},
				},
			},
			expected: []string{"vault"},
		},
		{
			name: "both providers enabled",
			config: SecretsConfig{
				Providers: ProvidersConfig{
					File: &FileProviderConfig{
						Enabled:  true,
						BasePath: "/etc/secrets",
					},
					Vault: &VaultProviderConfig{
						Enabled: true,
						Address: "https://vault.example.com",
					},
				},
			},
			expected: []string{"file", "vault"},
		},
		{
			name: "providers configured but not enabled",
			config: SecretsConfig{
				Providers: ProvidersConfig{
					File: &FileProviderConfig{
						Enabled:  false,
						BasePath: "/etc/secrets",
					},
					Vault: &VaultProviderConfig{
						Enabled: false,
						Address: "https://vault.example.com",
					},
				},
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.EnabledProviders()
			assert.Equal(t, tt.expected, result)
		})
	}
}
