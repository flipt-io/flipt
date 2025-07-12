package config

import (
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
			return errFieldWrap("storage", name, err)
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
		return errFieldWrap("storage", "ca_cert", fmt.Errorf("please provide only one of ca_cert_path or ca_cert_bytes"))
	}

	if c.PollInterval < 1*time.Second {
		return errFieldWrap("storage", "poll_interval", fmt.Errorf("poll interval must be at least 1 second"))
	}

	if c.Backend.Type == LocalStorageBackendType && c.Backend.Path == "" {
		return errFieldRequired("storage", "local path")
	}

	// Validate signature config
	if c.Signature.Enabled {
		if err := c.Signature.validate(); err != nil {
			return errFieldWrap("storage", "signature", err)
		}
	}

	return nil
}

type SignatureType string

const (
	GPGSignatureType = SignatureType("gpg")
)

// SignatureConfig contains details for producing git author and committer metadata,
// as well as optional commit signing configuration.
type SignatureConfig struct {
	Name    string           `json:"name" mapstructure:"name" yaml:"name"`
	Email   string           `json:"email" mapstructure:"email" yaml:"email"`
	Enabled bool             `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
	Type    SignatureType    `json:"type,omitempty" mapstructure:"type" yaml:"type,omitempty"`
	KeyRef  *SecretReference `json:"key_ref,omitempty" mapstructure:"key_ref" yaml:"key_ref,omitempty"`
	KeyID   string           `json:"key_id,omitempty" mapstructure:"key_id" yaml:"key_id,omitempty"`
}

func (c *SignatureConfig) validate() error {
	if c.Type == "" {
		return errFieldRequired("", "type")
	}

	if c.Type != GPGSignatureType {
		return errFieldWrap("", "type", fmt.Errorf("unsupported signing type: %s (only 'gpg' is supported)", c.Type))
	}

	if c.KeyRef == nil {
		return errFieldRequired("", "key_ref")
	}

	if err := c.KeyRef.Validate(); err != nil {
		return errFieldWrap("", "key_ref", err)
	}

	return nil
}
