package config

import "github.com/spf13/viper"

// UIConfig contains fields, which control the behaviour
// of Flipt's user interface.
type UIConfig struct {
	Enabled bool `json:"enabled" mapstructure:"enabled"`
}

func (c *UIConfig) setDefaults(v *viper.Viper) []string {
	v.SetDefault("ui", map[string]any{
		"enabled": true,
	})

	return nil
}
