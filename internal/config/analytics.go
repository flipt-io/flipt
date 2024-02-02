package config

import (
	"errors"

	"github.com/spf13/viper"
)

// AnalyticsConfig defines the configuration for various mechanisms for
// reporting and querying analytical data for Flipt.
type AnalyticsConfig struct {
	Enabled    bool             `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
	Clickhouse ClickhouseConfig `json:"clickhouse,omitempty" mapstructure:"clickhouse" yaml:"clickhouse,omitempty"`
}

// ClickhouseConfig defines the connection details for connecting Flipt to Clickhouse.
type ClickhouseConfig struct {
	Enabled bool   `json:"enabled.omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
	URL     string `json:"url,omitempty" mapstructure:"url" yaml:"url,omitempty"`
}

//nolint:unparam
func (m *AnalyticsConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("analytics.clickhouse.enabled", "false")
	v.SetDefault("analytics.clickhouse.url", "")

	return nil
}

func (m *AnalyticsConfig) validate() error {
	if m.Clickhouse.Enabled && m.Clickhouse.URL == "" {
		return errors.New("clickhouse url not provided")
	}

	return nil
}
