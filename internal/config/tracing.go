package config

import "github.com/spf13/viper"

// JaegerTracingConfig contains fields, which configure specifically
// Jaeger span and tracing output destination.
type JaegerTracingConfig struct {
	Enabled bool   `json:"enabled,omitempty" mapstructure:"enabled"`
	Host    string `json:"host,omitempty" mapstructure:"host"`
	Port    int    `json:"port,omitempty" mapstructure:"port"`
}

// TracingConfig contains fields, which configure tracing telemetry
// output destinations.
type TracingConfig struct {
	Jaeger JaegerTracingConfig `json:"jaeger,omitempty"`
}

func (c *TracingConfig) setDefaults(v *viper.Viper) []string {
	v.SetDefault("tracing", map[string]any{
		"jaeger": map[string]any{
			"host": "localhost",
			"port": 6831,
		},
	})

	return nil
}
