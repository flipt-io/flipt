package config

import "github.com/spf13/viper"

var _ defaulter = (*AuthenticationConfig)(nil)

// AuthenticationConfig configures Flipts authentication mechanisms
type AuthenticationConfig struct {
	Enabled bool `json:"enabled,omitempty" mapstructure:"enabled"`
}

func (a *AuthenticationConfig) setDefaults(v *viper.Viper) []string {
	v.SetDefault("authentication", map[string]any{
		"enabled": false,
	})

	return nil
}
