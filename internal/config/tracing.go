package config

import (
	"github.com/spf13/viper"
)

var (
	_ defaulter = (*TracingConfig)(nil)
)

// TracingConfig contains fields, which configure tracing telemetry
// output destinations.
type TracingConfig struct {
	Enabled bool `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
}

func (c *TracingConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("tracing", map[string]any{
		"enabled": false,
	})

	return nil
}
