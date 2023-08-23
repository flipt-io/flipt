package config

import "github.com/spf13/viper"

// cheers up the unparam linter
var _ defaulter = (*UIConfig)(nil)

// UIConfig contains fields, which control the behaviour
// of Flipt's user interface.
type UIConfig struct {
	Enabled bool `json:"enabled" mapstructure:"enabled"`
}

func (c *UIConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("ui", map[string]any{
		"enabled": true,
	})

	return nil
}

func (c *UIConfig) deprecations(v *viper.Viper) []deprecated {
	var deprecations []deprecated

	if v.InConfig("ui.enabled") {
		deprecations = append(deprecations, "ui.enabled")
	}

	return deprecations
}
