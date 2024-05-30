package config

import "github.com/spf13/viper"

var (
	_ defaulter = (*MetaConfig)(nil)
)

// MetaConfig contains a variety of meta configuration fields.
type MetaConfig struct {
	CheckForUpdates  bool   `json:"checkForUpdates" mapstructure:"check_for_updates" yaml:"check_for_updates"`
	TelemetryEnabled bool   `json:"telemetryEnabled" mapstructure:"telemetry_enabled" yaml:"telemetry_enabled"`
	StateDirectory   string `json:"stateDirectory,omitempty" mapstructure:"state_directory" yaml:"state_directory,omitempty"`
}

func (c *MetaConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("meta", map[string]any{
		"check_for_updates": true,
		"telemetry_enabled": true,
	})

	return nil
}
