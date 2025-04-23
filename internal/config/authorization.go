package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

var (
	_ defaulter = (*AuthorizationConfig)(nil)
	_ validator = (*AuthorizationConfig)(nil)
)

// AuthorizationConfig configures Flipts authorization mechanisms
type AuthorizationConfig struct {
	// Required designates whether authorization credentials are validated.
	// If required == true, then authorization is required for all API endpoints.
	Required bool                      `json:"required,omitempty" mapstructure:"required" yaml:"required,omitempty"`
	Backend  AuthorizationBackend      `json:"backend,omitempty" mapstructure:"backend" yaml:"backend,omitempty"`
	Local    *AuthorizationLocalConfig `json:"local,omitempty" mapstructure:"local,omitempty" yaml:"local,omitempty"`
}

func (c *AuthorizationConfig) setDefaults(v *viper.Viper) error {
	auth := map[string]any{"required": false}
	if v.GetBool("authorization.required") {

		auth["backend"] = AuthorizationBackendLocal
		if v.GetString("authorization.local.data.path") == "" {
			auth["local"] = map[string]any{
				"policy": map[string]any{
					"poll_interval": "5m",
				},
			}
		} else {
			auth["local"] = map[string]any{
				"policy": map[string]any{
					"poll_interval": "5m",
				},
				"data": map[string]any{
					"poll_interval": "30s",
				},
			}
		}

		v.SetDefault("authorization", auth)
	}

	return nil
}

func (c *AuthorizationConfig) validate() error {
	if c.Required {
		switch c.Backend {
		case AuthorizationBackendLocal:
			if c.Local == nil {
				return errors.New("authorization: local backend must be configured")
			}

			if err := c.Local.validate(); err != nil {
				return fmt.Errorf("authorization: local: %w", err)
			}

		default:
			return errString("authorization", "backend must be one of local")
		}
	}

	return nil
}

// AuthorizationBackend configures the data source backend for
// additional data supplied to the policy evaluation engine
type AuthorizationBackend string

const (
	AuthorizationBackendLocal = AuthorizationBackend("local")
)

// AuthorizationLocalConfig configures the local backend source
// for the authorization evaluation engines policies and data
type AuthorizationLocalConfig struct {
	Policy *AuthorizationSourceLocalConfig `json:"policy,omitempty" mapstructure:"policy,omitempty" yaml:"policy,omitempty"`
	Data   *AuthorizationSourceLocalConfig `json:"data,omitempty" mapstructure:"data,omitempty" yaml:"data,omitempty"`
}

func (c *AuthorizationLocalConfig) validate() error {
	if c.Policy == nil {
		return errString("authorization", "policy source must be configured")
	}

	if err := c.Policy.validate(); err != nil {
		return errFieldWrap("authorization", "policy", err)
	}

	if c.Data != nil {
		if err := c.Data.validate(); err != nil {
			return errFieldWrap("authorization", "data", err)
		}
	}

	return nil
}

type AuthorizationSourceLocalConfig struct {
	Path         string        `json:"path,omitempty" mapstructure:"path" yaml:"path,omitempty"`
	PollInterval time.Duration `json:"pollInterval,omitempty" mapstructure:"poll_interval" yaml:"poll_interval,omitempty"`
}

func (a *AuthorizationSourceLocalConfig) validate() error {
	if a == nil {
		return nil
	}

	if a.Path == "" {
		return errFieldRequired("authorization", "path")
	}

	if a.PollInterval <= 0 {
		return errFieldPositiveDuration("authorization", "poll_interval")
	}

	return nil
}
