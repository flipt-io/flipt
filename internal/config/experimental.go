package config

// ExperimentalConfig allows for experimental features to be enabled
// and disabled.
type ExperimentalConfig struct {
}

// ExperimentalFlag is a structure which has properties to configure
// experimental feature enablement.
type ExperimentalFlag struct {
	Enabled bool `json:"enabled,omitempty" mapstructure:"enabled"`
}
