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

func (s StoragesConfig) validate() error {
	for name, storage := range s {
		if err := storage.validate(); err != nil {
			return fmt.Errorf("storage %s: %w", name, err)
		}
	}

	return nil
}

func (s StoragesConfig) setDefaults(v *viper.Viper) error {
	storages, ok := v.AllSettings()["storage"].(map[string]any)
	if !ok {
		return nil
	}

	for name := range storages {
		getString := func(s string) string {
			return v.GetString("storage." + name + "." + s)
		}

		setDefault := func(k string, s any) {
			v.SetDefault("storage."+name+"."+k, s)
		}

		// force name to map key
		v.Set("storage."+name+".name", name)

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
	Remote          string                         `json:"remote,omitempty" mapstructure:"remote" yaml:"remote,omitempty"`
	Backend         StorageBackendConfig           `json:"backend,omitempty" mapstructure:"backend" yaml:"backend,omitempty"`
	Branch          string                         `json:"branch,omitempty" mapstructure:"branch" yaml:"branch,omitempty"`
	CaCertBytes     string                         `json:"-" mapstructure:"ca_cert_bytes" yaml:"-"`
	CaCertPath      string                         `json:"-" mapstructure:"ca_cert_path" yaml:"-"`
	PollInterval    time.Duration                  `json:"pollInterval,omitempty" mapstructure:"poll_interval" yaml:"poll_interval,omitempty"`
	InsecureSkipTLS bool                           `json:"-" mapstructure:"insecure_skip_tls" yaml:"-"`
	Authentication  StorageGitAuthenticationConfig `json:"-" mapstructure:"authentication,omitempty" yaml:"-"`
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

	return c.Authentication.validate()
}

// StorageGitAuthenticationConfig holds structures for various types of auth we support.
// Token auth will take priority over Basic auth if both are provided.
//
// To make things easier, if there are multiple inputs that a particular auth method needs, and
// not all inputs are given but only partially, we will return a validation error.
// (e.g. if username for basic auth is given, and token is also given a validation error will be returned)
type StorageGitAuthenticationConfig struct {
	BasicAuth *GitBasicAuth `json:"-" mapstructure:"basic,omitempty" yaml:"-"`
	TokenAuth *GitTokenAuth `json:"-" mapstructure:"token,omitempty" yaml:"-"`
	SSHAuth   *GitSSHAuth   `json:"-" mapstructure:"ssh,omitempty" yaml:"-"`
}

func (a *StorageGitAuthenticationConfig) validate() error {
	if a.BasicAuth != nil {
		if err := a.BasicAuth.validate(); err != nil {
			return err
		}
	}
	if a.TokenAuth != nil {
		if err := a.TokenAuth.validate(); err != nil {
			return err
		}
	}
	if a.SSHAuth != nil {
		if err := a.SSHAuth.validate(); err != nil {
			return err
		}
	}

	return nil
}

// GitBasicAuth has configuration for authenticating with private git repositories
// with basic auth.
type GitBasicAuth struct {
	Username string `json:"-" mapstructure:"username" yaml:"-"`
	Password string `json:"-" mapstructure:"password" yaml:"-"`
}

func (b GitBasicAuth) validate() error {
	if (b.Username != "" && b.Password == "") || (b.Username == "" && b.Password != "") {
		return errors.New("both username and password need to be provided for basic auth")
	}

	return nil
}

// GitTokenAuth has configuration for authenticating with private git repositories
// with token auth.
type GitTokenAuth struct {
	AccessToken string `json:"-" mapstructure:"access_token" yaml:"-"`
}

func (t GitTokenAuth) validate() error { return nil }

// GitSSHAuth provides configuration support for SSH private key credentials when
// authenticating with private git repositories
type GitSSHAuth struct {
	User                  string `json:"-" mapstructure:"user" yaml:"-" `
	Password              string `json:"-" mapstructure:"password" yaml:"-" `
	PrivateKeyBytes       string `json:"-" mapstructure:"private_key_bytes" yaml:"-" `
	PrivateKeyPath        string `json:"-" mapstructure:"private_key_path" yaml:"-" `
	InsecureIgnoreHostKey bool   `json:"-" mapstructure:"insecure_ignore_host_key" yaml:"-"`
}

func (a GitSSHAuth) validate() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("ssh authentication %w", err)
		}
	}()

	if a.Password == "" {
		return errors.New("password required")
	}

	if (a.PrivateKeyBytes == "" && a.PrivateKeyPath == "") || (a.PrivateKeyBytes != "" && a.PrivateKeyPath != "") {
		return errors.New("please provide exclusively one of private_key_bytes or private_key_path")
	}

	return nil
}
