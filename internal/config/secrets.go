package config

import (
	"errors"
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

func (s *SecretsConfig) setDefaults(v *viper.Viper) []string {
	var warnings []string

	// File provider defaults
	v.SetDefault("secrets.providers.file.enabled", false)
	v.SetDefault("secrets.providers.file.base_path", "/etc/flipt/secrets")

	// Vault provider defaults
	v.SetDefault("secrets.providers.vault.enabled", false)
	v.SetDefault("secrets.providers.vault.mount", "secret")
	v.SetDefault("secrets.providers.vault.auth_method", "token")

	return warnings
}

func (s *SecretsConfig) validate() error {
	// Validate file provider
	if s.Providers.File != nil && s.Providers.File.Enabled {
		if s.Providers.File.BasePath == "" {
			return errors.New("secrets.providers.file.base_path is required when file provider is enabled")
		}
	}

	// Validate vault provider
	if s.Providers.Vault != nil && s.Providers.Vault.Enabled {
		if s.Providers.Vault.Address == "" {
			return errors.New("secrets.providers.vault.address is required when vault provider is enabled")
		}

		switch s.Providers.Vault.AuthMethod {
		case "token":
			if s.Providers.Vault.Token == "" {
				return errors.New("secrets.providers.vault.token is required when using token auth method")
			}
		case "kubernetes":
			if s.Providers.Vault.Role == "" {
				return errors.New("secrets.providers.vault.role is required when using kubernetes auth method")
			}
		case "approle":
			if s.Providers.Vault.Role == "" {
				return errors.New("secrets.providers.vault.role is required when using approle auth method")
			}
		default:
			return fmt.Errorf("unsupported vault auth method: %s", s.Providers.Vault.AuthMethod)
		}
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
		return errors.New("provider is required")
	}
	if r.Path == "" {
		return errors.New("path is required")
	}
	if r.Key == "" {
		return errors.New("key is required")
	}
	return nil
}