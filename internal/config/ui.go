package config

import "github.com/spf13/viper"

type UITheme string

const (
	SystemUITheme = UITheme("system")
	DarkUITheme   = UITheme("dark")
	LightUITheme  = UITheme("light")
)

// cheers up the unparam linter
var _ defaulter = (*UIConfig)(nil)

// UIConfig contains fields, which control the behaviour
// of Flipt's user interface.
type UIConfig struct {
	Enabled      bool    `json:"enabled" mapstructure:"enabled"`
	DefaultTheme UITheme `json:"defaultTheme" mapstructure:"default_theme"`
}

func (c *UIConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("ui", map[string]any{
		"enabled":       true,
		"default_theme": SystemUITheme,
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
