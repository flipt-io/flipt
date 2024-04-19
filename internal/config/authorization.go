package config

import "github.com/spf13/viper"

var (
	_ defaulter = (*AuthorizationConfig)(nil)
)

// AuthorizationConfig configures Flipts authorization mechanisms
type AuthorizationConfig struct {
	// Required designates whether authorization credentials are validated.
	// If required == true, then authorization is required for all API endpoints.
	Required bool `json:"required" mapstructure:"required" yaml:"required"`
}

func (c *AuthorizationConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("authorization", map[string]interface{}{
		"required": false,
	})

	return nil
}
