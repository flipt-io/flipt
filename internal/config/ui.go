package config

import "github.com/spf13/viper"

type UITheme string

const (
	SystemUITheme = UITheme("system")
	DarkUITheme   = UITheme("dark")
	LightUITheme  = UITheme("light")
)

var (
	_ defaulter = (*UIConfig)(nil)
)

// UIConfig contains fields, which control the behaviour
// of Flipt's user interface.
type UIConfig struct {
	DefaultTheme UITheme `json:"defaultTheme" mapstructure:"default_theme" yaml:"default_theme"`
}

func (c *UIConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("ui", map[string]any{
		"default_theme": SystemUITheme,
	})

	return nil
}
