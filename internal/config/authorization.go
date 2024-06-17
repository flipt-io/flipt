package config

import (
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
		auth["local.policy"] = map[string]any{
			"poll_interval": "5m",
		}

		if v.GetString("authorization.local.data.path") != "" {
			auth["local.data"] = map[string]any{
				"poll_interval": "30s",
			}
		}

		v.SetDefault("authorization", auth)
	}

	return nil
}

func (c *AuthorizationConfig) validate() error {
	// if c.Required {
	// 	if c.Policy == nil {
	// 		return errors.New("authorization: policy source must be configured")
	// 	}

	// 	if err := c.Policy.validate(); err != nil {
	// 		return fmt.Errorf("authorization: policy: %w", err)
	// 	}

	// 	if err := c.Data.validate(); err != nil {
	// 		return fmt.Errorf("authorization: data: %w", err)
	// 	}
	// }

	return nil
}

// AuthorizationBackend configures the data source backend for
// additional data supplied to the policy evaluation engine
type AuthorizationBackend string

const (
	AuthorizationBackendLocal  = AuthorizationBackend("local")
	AuthorizationBackendObject = AuthorizationBackend("object")
	AuthorizationBackendOCI    = AuthorizationBackend("oci")
)

// AuthorizationLocalConfig configures the local backend source
// for the authorization evaluation engines policies and data
type AuthorizationLocalConfig struct {
	Policy *AuthorizationSourceLocalConfig `json:"policy,omitempty" mapstructure:"policy,omitempty" yaml:"policy,omitempty"`
	Data   *AuthorizationSourceLocalConfig `json:"data,omitempty" mapstructure:"data,omitempty" yaml:"data,omitempty"`
}

type AuthorizationSourceLocalConfig struct {
	Path         string        `json:"path,omitempty" mapstructure:"path" yaml:"backend,path"`
	PollInterval time.Duration `json:"pollInterval,omitempty" mapstructure:"poll_interval" yaml:"poll_interval,omitempty"`
}

func (a *AuthorizationSourceLocalConfig) validate() (err error) {
	// defer func() {
	// 	if err != nil {
	// 		err = fmt.Errorf("source: %w", err)
	// 	}
	// }()

	// if a == nil {
	// 	return nil
	// }

	// if a.Backend != AuthorizationBackendLocal {
	// 	return errors.New("backend must be one of [local]")
	// }

	// if a.Local == nil || a.Local.Path == "" {
	// 	return errors.New("local: path must be non-empty string")
	// }

	// if a.PollInterval <= 0 {
	// 	return errors.New("local: poll_interval must be non-zero")
	// }

	return nil
}
