package config

import (
	"errors"
	"time"

	"github.com/spf13/viper"
)

// AnalyticsConfig defines the configuration for various mechanisms for
// reporting and querying analytical data for Flipt.
type AnalyticsConfig struct {
	Enabled    bool             `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
	Clickhouse ClickhouseConfig `json:"clickhouse,omitempty" mapstructure:"clickhouse" yaml:"clickhouse,omitempty"`
	Buffer     BufferConfig     `json:"buffer,omitempty" mapstructure:"buffer" yaml:"buffer,omitempty"`
}

// ClickhouseConfig defines the connection details for connecting Flipt to Clickhouse.
type ClickhouseConfig struct {
	Enabled bool   `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
	URL     string `json:"url,omitempty" mapstructure:"url" yaml:"url,omitempty"`
}

//nolint:unparam
func (a *AnalyticsConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("analytics.clickhouse.enabled", "false")
	v.SetDefault("analytics.clickhouse.url", "")
	v.SetDefault("analytics.buffer.capacity", 10)
	v.SetDefault("analytics.buffer.flush_period", "10s")

	return nil
}

func (a *AnalyticsConfig) validate() error {
	if a.Clickhouse.Enabled && a.Clickhouse.URL == "" {
		return errors.New("clickhouse url not provided")
	}

	if a.Buffer.Capacity < 10 {
		return errors.New("capacity below 10")
	}

	if a.Buffer.FlushPeriod < time.Second*10 {
		return errors.New("flush period below 10 seconds")
	}

	return nil
}
