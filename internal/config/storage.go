package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

var (
	_ defaulter = (*StoragesConfig)(nil)
	_ validator = (*StoragesConfig)(nil)
)

type StoragesConfig map[string]*StorageConfig

func (s *StoragesConfig) validate() error {
	for name, storage := range *s {
		if err := storage.validate(); err != nil {
			return fmt.Errorf("storage %s: %w", name, err)
		}
	}

	return nil
}

func (s *StoragesConfig) setDefaults(v *viper.Viper) error {
	storages, ok := v.AllSettings()["storage"].(map[string]any)
	if !ok || len(storages) == 0 {
		// Create default storage if no storages are configured
		v.SetDefault("storage.default.name", "default")
		v.SetDefault("storage.default.backend.type", "memory")
		v.SetDefault("storage.default.branch", "main")
		v.SetDefault("storage.default.poll_interval", "30s")
		return nil
	}

	for name := range storages {
		getString := func(s string) string {
			return v.GetString("storage." + name + "." + s)
		}

		setDefault := func(k string, s any) {
			v.SetDefault("storage."+name+"."+k, s)
		}

		setDefault("name", name)

		switch getString("backend.type") {
		case string(LocalStorageBackendType):
			setDefault("backend.path", ".")
		default:
			setDefault("backend.type", "memory")
		}

		if getString("branch") == "" {
			setDefault("branch", "main")
		}

		if getString("poll_interval") == "" {
			setDefault("poll_interval", "30s")
		}
	}

	return nil
}

type StorageBackendType string

const (
	LocalStorageBackendType  = StorageBackendType("local")
	MemoryStorageBackendType = StorageBackendType("memory")
)

type StorageBackendConfig struct {
	Type StorageBackendType `json:"type,omitempty" mapstructure:"type" yaml:"type,omitempty"`
	Path string             `json:"path,omitempty" mapstructure:"path" yaml:"path,omitempty"`
}

// StorageConfig contains fields which will configure the type of backend in which Flipt will serve
// flag state.
type StorageConfig struct {
	Remote          string               `json:"remote,omitempty" mapstructure:"remote" yaml:"remote,omitempty"`
	Backend         StorageBackendConfig `json:"backend,omitempty" mapstructure:"backend" yaml:"backend,omitempty"`
	Branch          string               `json:"branch,omitempty" mapstructure:"branch" yaml:"branch,omitempty"`
	CaCertBytes     string               `json:"-" mapstructure:"ca_cert_bytes" yaml:"-"`
	CaCertPath      string               `json:"-" mapstructure:"ca_cert_path" yaml:"-"`
	PollInterval    time.Duration        `json:"pollInterval,omitempty" mapstructure:"poll_interval" yaml:"poll_interval,omitempty"`
	InsecureSkipTLS bool                 `json:"-" mapstructure:"insecure_skip_tls" yaml:"-"`
	Credentials     string               `json:"-" mapstructure:"credentials" yaml:"-"`
	Signature       SignatureConfig      `json:"signature,omitempty" mapstructure:"signature,omitempty" yaml:"signature,omitempty"`
}

func (c *StorageConfig) validate() error {
	if c.CaCertPath != "" && c.CaCertBytes != "" {
		return errors.New("please provide only one of ca_cert_path or ca_cert_bytes")
	}

	if c.PollInterval < 1*time.Second {
		return errors.New("poll interval must be at least 1 second")
	}

	if c.Backend.Type == LocalStorageBackendType && c.Backend.Path == "" {
		return errors.New("local path must be specified")
	}

	// Check for common misconfiguration where remote is placed under backend
	v := viper.GetViper()
	for storageName, storageConfig := range v.GetStringMap("storage") {
		if backend, ok := storageConfig.(map[string]interface{})["backend"].(map[string]interface{}); ok {
			if _, exists := backend["remote"]; exists {
				return fmt.Errorf("storage %s: 'remote' should not be placed under 'backend' - move it to the storage level instead", storageName)
			}
		}
	}

	return nil
}

// SignatureConfig contains details for producing git author and committer metadata.
type SignatureConfig struct {
	Name  string `json:"name" mapstructure:"name" yaml:"name"`
	Email string `json:"email" mapstructure:"email" yaml:"email"`
}
