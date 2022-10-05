package config

import "github.com/spf13/viper"

const (
	// configuration keys
	metaCheckForUpdates  = "meta.check_for_updates"
	metaTelemetryEnabled = "meta.telemetry_enabled"
	metaStateDirectory   = "meta.state_directory"
)

// MetaConfig contains a variety of meta configuration fields.
type MetaConfig struct {
	CheckForUpdates  bool   `json:"checkForUpdates"`
	TelemetryEnabled bool   `json:"telemetryEnabled"`
	StateDirectory   string `json:"stateDirectory"`
}

func (c *MetaConfig) init() (warnings []string, _ error) {
	if viper.IsSet(metaCheckForUpdates) {
		c.CheckForUpdates = viper.GetBool(metaCheckForUpdates)
	}

	if viper.IsSet(metaTelemetryEnabled) {
		c.TelemetryEnabled = viper.GetBool(metaTelemetryEnabled)
	}

	if viper.IsSet(metaStateDirectory) {
		c.StateDirectory = viper.GetString(metaStateDirectory)
	}

	return
}
