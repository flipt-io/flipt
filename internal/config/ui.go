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

// UITopbar represents the configuration of a user interface top bar component.
type UITopbar struct {
	Color string `json:"color" mapstructure:"color" yaml:"color"`
	Label string `json:"label" mapstructure:"label" yaml:"label"`
}

// UIConfig contains fields, which control the behaviour
// of Flipt's user interface.
type UIConfig struct {
	DefaultTheme UITheme  `json:"defaultTheme" mapstructure:"default_theme" yaml:"default_theme"`
	Topbar       UITopbar `json:"topbar,omitempty" mapstructure:"topbar" yaml:"topbar,omitempty"`
}

func (c *UIConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("ui", map[string]any{
		"default_theme": SystemUITheme,
	})

	return nil
}
