package config

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
