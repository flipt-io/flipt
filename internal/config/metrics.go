package config

import (
	"github.com/spf13/viper"
)

var (
	_ defaulter = (*MetricsConfig)(nil)
)

type MetricsExporter string

const (
	MetricsPrometheus MetricsExporter = "prometheus"
	MetricsOTLP       MetricsExporter = "otlp"
)

type MetricsConfig struct {
	Enabled  bool              `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
	Exporter MetricsExporter   `json:"exporter,omitempty" mapstructure:"exporter" yaml:"exporter,omitempty"`
	OTLP     OTLPMetricsConfig `json:"otlp,omitempty" mapstructure:"otlp,omitempty" yaml:"otlp,omitempty"`
}

type OTLPMetricsConfig struct {
	Endpoint string            `json:"endpoint,omitempty" mapstructure:"endpoint" yaml:"endpoint,omitempty"`
	Headers  map[string]string `json:"headers,omitempty" mapstructure:"headers" yaml:"headers,omitempty"`
}

func (c *MetricsConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("metrics", map[string]interface{}{
		"enabled":  true,
		"exporter": MetricsPrometheus,
	})

	return nil
}
