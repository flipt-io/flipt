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
	Enabled          bool   `json:"enabled.omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
	ConnectionString string `json:"connectionString,omitempty" mapstructure:"connection_string" yaml:"connection_string,omitempty"`
}

func (m *AnalyticsConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("metrics.clickhouse.enabled", "false")
	v.SetDefault("metrics.clickhouse.connection_string", "")

	return nil
}

func (m *AnalyticsConfig) validate() error {
	if m.Clickhouse.Enabled && m.Clickhouse.ConnectionString == "" {
		return errors.New("clickhouse connection string not provided")
	}

	return nil
}
