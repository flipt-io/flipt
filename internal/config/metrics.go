package config

import (
	"github.com/spf13/viper"
)

var _ defaulter = (*MetricsConfig)(nil)

type MetricsExporter string

const (
	MetricsPrometheus MetricsExporter = "prometheus"
	MetricsOTLP       MetricsExporter = "otlp"
)

type MetricsConfig struct {
	Enabled  bool            `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
	Exporter MetricsExporter `json:"exporter,omitempty" mapstructure:"exporter" yaml:"exporter,omitempty"`
}

func (c *MetricsConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("metrics.enabled", true)
	v.SetDefault("metrics.exporter", MetricsPrometheus)
	return nil
}
