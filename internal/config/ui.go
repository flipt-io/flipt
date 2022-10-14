package config

import "github.com/spf13/viper"

const (
	// configuration keys
	uiEnabled = "ui.enabled"
)

// UIConfig contains fields, which control the behaviour
// of Flipt's user interface.
type UIConfig struct {
	Enabled bool `json:"enabled"`
}

func (c *UIConfig) init() (_ []string, _ error) {
	if viper.IsSet(uiEnabled) {
		c.Enabled = viper.GetBool(uiEnabled)
	}

	return
}
