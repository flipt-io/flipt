package config

import "github.com/spf13/viper"

var (
	_ defaulter = (*DiagnosticConfig)(nil)
)

type DiagnosticConfig struct {
	Profiling ProfilingDiagnosticConfig `json:"profiling,omitempty" mapstructure:"profiling" yaml:"profiling,omitempty"`
}

type ProfilingDiagnosticConfig struct {
	Enabled bool `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
}

func (c *DiagnosticConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("diagnostics", map[string]any{
		"profiling": map[string]any{
			"enabled": true,
		},
	})

	return nil
}
