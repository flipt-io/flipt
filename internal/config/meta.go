package config

import "github.com/spf13/viper"

// cheers up the unparam linter
var _ defaulter = (*MetaConfig)(nil)

// MetaConfig contains a variety of meta configuration fields.
type MetaConfig struct {
	CheckForUpdates  bool   `json:"checkForUpdates" mapstructure:"check_for_updates"`
	TelemetryEnabled bool   `json:"telemetryEnabled" mapstructure:"telemetry_enabled"`
	StateDirectory   string `json:"stateDirectory" mapstructure:"state_directory"`
}

func (c *MetaConfig) setDefaults(v *viper.Viper) []string {
	v.SetDefault("meta", map[string]any{
		"check_for_updates": true,
		"telemetry_enabled": true,
	})

	return nil
}
