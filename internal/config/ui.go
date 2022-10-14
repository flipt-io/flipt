package config

import "github.com/spf13/viper"

// UIConfig contains fields, which control the behaviour
// of Flipt's user interface.
type UIConfig struct {
	Enabled bool `json:"enabled" mapstructure:"enabled"`
}

func (c *UIConfig) viperKey() string {
	return "ui"
}

func (c *UIConfig) unmarshalViper(v *viper.Viper) (_ []string, _ error) {
	v.SetDefault("enabled", true)

	return nil, v.Unmarshal(c)
}
