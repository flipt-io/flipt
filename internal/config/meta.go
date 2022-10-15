package config

import "github.com/spf13/viper"

// MetaConfig contains a variety of meta configuration fields.
type MetaConfig struct {
	CheckForUpdates  bool   `json:"checkForUpdates" mapstructure:"check_for_updates"`
	TelemetryEnabled bool   `json:"telemetryEnabled" mapstructure:"telemetry_enabled"`
	StateDirectory   string `json:"stateDirectory" mapstructure:"state_directory"`
}

func (c *MetaConfig) viperKey() string {
	return "meta"
}

func (c *MetaConfig) unmarshalViper(v *viper.Viper) (warnings []string, _ error) {
	v.SetDefault("check_for_updates", true)
	v.SetDefault("telemetry_enabled", true)
	return nil, v.Unmarshal(c)
}
