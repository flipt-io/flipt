package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

var (
	_ defaulter = (*AuthorizationConfig)(nil)
)

// AuthorizationConfig configures Flipts authorization mechanisms
type AuthorizationConfig struct {
	// Required designates whether authorization credentials are validated.
	// If required == true, then authorization is required for all API endpoints.
	Required     bool                       `json:"required,omitempty" mapstructure:"required" yaml:"required,omitempty"`
	Policy       *AuthorizationSourceConfig `json:"policy,omitempty" mapstructure:"policy,omitempty" yaml:"policy,omitempty"`
	Data         *AuthorizationSourceConfig `json:"data,omitempty" mapstructure:"data,omitempty" yaml:"data,omitempty"`
	PollDuration time.Duration              `json:"pollDuration,omitempty" mapstructure:"poll_duration" yaml:"poll_duration,omitempty"`
}

func (c *AuthorizationConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("authorization", map[string]interface{}{
		"required":      false,
		"poll_duration": "30s",
	})

	return nil
}

func (c *AuthorizationConfig) validate() error {
	if c.Required {
		if c.Policy == nil {
			return errors.New("authorization: policy source must be configured")
		}

		if err := c.Policy.validate(); err != nil {
			return fmt.Errorf("authorization: policy: %w", err)
		}

		if err := c.Data.validate(); err != nil {
			return fmt.Errorf("authorization: data: %w", err)
		}
	}

	return nil
}

type AuthorizationSourceConfig struct {
	Backend    AuthorizationBackend          `json:"backend,omitempty" mapstructure:"backend" yaml:"backend,omitempty"`
	Filesystem AuthorizationFilesystemConfig `json:"filesystem,omitempty" mapstructure:"filesystem" yaml:"filesystem,omitempty"`
}

func (a *AuthorizationSourceConfig) validate() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("source: %w", err)
		}
	}()

	if a == nil {
		return nil
	}

	if a.Backend != AuthorizationBackendFilesystem {
		return errors.New("backend must be one of [filesystem]")
	}

	if a.Filesystem.Path == "" {
		return errors.New("filesystem: path must be non-empty string")
	}

	return nil
}

// AuthorizationBackend configures the data source backend for
// additional data supplied to the policy evaluation engine
type AuthorizationBackend string

const (
	AuthorizationBackendFilesystem = "filesystem"
)

// AuthorizationFilesystemConfig configures the filesystem backend source
// for the authorization evaluation engines policies and data
type AuthorizationFilesystemConfig struct {
	Path string `json:"path,omitempty" mapstructure:"path" yaml:"path,omitempty"`
}
