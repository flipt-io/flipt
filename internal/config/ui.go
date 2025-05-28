package config

import (
	"fmt"
	"regexp"

	"github.com/spf13/viper"
)

type UITheme string

const (
	SystemUITheme = UITheme("system")
	DarkUITheme   = UITheme("dark")
	LightUITheme  = UITheme("light")
)

var (
	_          defaulter = (*UIConfig)(nil)
	hexedColor           = regexp.MustCompile("^#(?:[0-9a-fA-F]{3}){1,2}$")
)

// UITopbar represents the configuration of a user interface top bar component.
type UITopbar struct {
	Color string `json:"color" mapstructure:"color" yaml:"color"`
}

func (u *UITopbar) validate() error {
	if u.Color != "" && !hexedColor.MatchString(u.Color) {
		return errFieldWrap("ui", "topbar", fmt.Errorf("expected valid hex color, got %s", u.Color))
	}
	return nil
}

// UIConfig contains fields, which control the behaviour
// of Flipt's user interface.
type UIConfig struct {
	Enabled      bool     `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
	DefaultTheme UITheme  `json:"defaultTheme" mapstructure:"default_theme" yaml:"default_theme"`
	Topbar       UITopbar `json:"topbar,omitempty" mapstructure:"topbar" yaml:"topbar,omitempty"`
}

func (c *UIConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("ui.enabled", true)
	v.SetDefault("ui.default_theme", SystemUITheme)
	return nil
}

func (c *UIConfig) validate() error {
	return c.Topbar.validate()
}
