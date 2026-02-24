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

func TestSecretsConfig_ValidateGCP(t *testing.T) {
	tests := []struct {
		name    string
		config  SecretsConfig
		wantErr bool
	}{
		{
			name: "valid gcp config",
			config: SecretsConfig{
				Providers: ProvidersConfig{
					GCP: &GCPProviderConfig{
						Enabled: true,
						Project: "my-project",
					},
				},
			},
		},
		{
			name: "missing project",
			config: SecretsConfig{
				Providers: ProvidersConfig{
					GCP: &GCPProviderConfig{
						Enabled: true,
					},
				},
			},
			wantErr: true,
		},
		{
			name: "disabled skips validation",
			config: SecretsConfig{
				Providers: ProvidersConfig{
					GCP: &GCPProviderConfig{
						Enabled: false,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "project")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestSecretsConfig_ValidateAWS(t *testing.T) {
	tests := []struct {
		name   string
		config SecretsConfig
	}{
		{
			name: "valid aws config enabled",
			config: SecretsConfig{
				Providers: ProvidersConfig{
					AWS: &AWSProviderConfig{
						Enabled: true,
					},
				},
			},
		},
		{
			name: "disabled skips validation",
			config: SecretsConfig{
				Providers: ProvidersConfig{
					AWS: &AWSProviderConfig{
						Enabled: false,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			require.NoError(t, err)
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
			name: "gcp provider enabled",
			config: SecretsConfig{
				Providers: ProvidersConfig{
					GCP: &GCPProviderConfig{
						Enabled: true,
						Project: "my-project",
					},
				},
			},
			expected: []string{"gcp"},
		},
		{
			name: "aws provider enabled",
			config: SecretsConfig{
				Providers: ProvidersConfig{
					AWS: &AWSProviderConfig{
						Enabled: true,
					},
				},
			},
			expected: []string{"aws"},
		},
		{
			name: "all providers enabled",
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
					GCP: &GCPProviderConfig{
						Enabled: true,
						Project: "my-project",
					},
					AWS: &AWSProviderConfig{
						Enabled: true,
					},
				},
			},
			expected: []string{"file", "vault", "gcp", "aws"},
		},
		{
			name: "gcp provider configured but not enabled",
			config: SecretsConfig{
				Providers: ProvidersConfig{
					GCP: &GCPProviderConfig{
						Enabled: false,
						Project: "my-project",
					},
				},
			},
			expected: nil,
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
