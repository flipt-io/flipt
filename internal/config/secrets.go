package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// SecretsConfig contains configuration for the secrets management system.
type SecretsConfig struct {
	Providers ProvidersConfig `json:"providers,omitempty" mapstructure:"providers" yaml:"providers,omitempty"`
}

// ProvidersConfig contains configuration for all secret providers.
type ProvidersConfig struct {
	File  *FileProviderConfig  `json:"file,omitempty" mapstructure:"file" yaml:"file,omitempty"`
	Vault *VaultProviderConfig `json:"vault,omitempty" mapstructure:"vault" yaml:"vault,omitempty"`
}

// FileProviderConfig contains configuration for the file-based secret provider.
type FileProviderConfig struct {
	Enabled  bool   `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
	BasePath string `json:"base_path,omitempty" mapstructure:"base_path" yaml:"base_path,omitempty"`
}

// VaultProviderConfig contains configuration for HashiCorp Vault provider.
type VaultProviderConfig struct {
	Enabled    bool   `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
	Address    string `json:"address,omitempty" mapstructure:"address" yaml:"address,omitempty"`
	AuthMethod string `json:"auth_method,omitempty" mapstructure:"auth_method" yaml:"auth_method,omitempty"`
	Role       string `json:"role,omitempty" mapstructure:"role" yaml:"role,omitempty"`
	Mount      string `json:"mount,omitempty" mapstructure:"mount" yaml:"mount,omitempty"`
	Token      string `json:"token,omitempty" mapstructure:"token" yaml:"token,omitempty"`
	Namespace  string `json:"namespace,omitempty" mapstructure:"namespace" yaml:"namespace,omitempty"`
}

// SecretReference represents a reference to a secret value.
type SecretReference struct {
	Provider string `json:"provider" mapstructure:"provider" yaml:"provider"`
	Path     string `json:"path" mapstructure:"path" yaml:"path"`
	Key      string `json:"key" mapstructure:"key" yaml:"key"`
}

func (s *SecretsConfig) setDefaults(v *viper.Viper) { //nolint:unused // used by config system via reflection
	// File provider defaults
	v.SetDefault("secrets.providers.file.enabled", false)
	v.SetDefault("secrets.providers.file.base_path", "/etc/flipt/secrets")

	// Vault provider defaults
	v.SetDefault("secrets.providers.vault.enabled", false)
	v.SetDefault("secrets.providers.vault.mount", "secret")
	v.SetDefault("secrets.providers.vault.auth_method", "token")
}

func (s *SecretsConfig) validate() error {
	// Validate file provider
	if s.Providers.File != nil && s.Providers.File.Enabled {
		if s.Providers.File.BasePath == "" {
			return errFieldRequired("secrets.providers.file", "base_path")
		}
	}

	// Validate vault provider
	if s.Providers.Vault != nil && s.Providers.Vault.Enabled {
		if s.Providers.Vault.Address == "" {
			return errFieldRequired("secrets.providers.vault", "address")
		}

		switch s.Providers.Vault.AuthMethod {
		case "token":
			if s.Providers.Vault.Token == "" {
				return errFieldRequired("secrets.providers.vault", "token")
			}
		case "kubernetes":
			if s.Providers.Vault.Role == "" {
				return errFieldRequired("secrets.providers.vault", "role")
			}
		case "approle":
			if s.Providers.Vault.Role == "" {
				return errFieldRequired("secrets.providers.vault", "role")
			}
		default:
			return errFieldWrap("secrets.providers.vault", "auth_method", fmt.Errorf("unsupported auth method: %s", s.Providers.Vault.AuthMethod))
		}

		// Note: License validation is done at runtime when the provider is instantiated
		// This allows the config to be valid but the feature to fail if no license is present
	}

	return nil
}

// EnabledProviders returns a list of enabled provider names.
func (s *SecretsConfig) EnabledProviders() []string {
	var providers []string

	if s.Providers.File != nil && s.Providers.File.Enabled {
		providers = append(providers, "file")
	}

	if s.Providers.Vault != nil && s.Providers.Vault.Enabled {
		providers = append(providers, "vault")
	}

	return providers
}

// Validate checks if a secret reference is valid.
func (r SecretReference) Validate() error {
	if r.Provider == "" {
		return errFieldRequired("secret_reference", "provider")
	}
	if r.Path == "" {
		return errFieldRequired("secret_reference", "path")
	}
	if r.Key == "" {
		return errFieldRequired("secret_reference", "key")
	}
	return nil
}
