package config

import (
	"errors"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/spf13/viper"
)

// AnalyticsConfig defines the configuration for various mechanisms for
// reporting and querying analytical data for Flipt.
type AnalyticsConfig struct {
	Clickhouse ClickhouseConfig `json:"clickhouse,omitempty" mapstructure:"clickhouse" yaml:"clickhouse,omitempty"`
	Buffer     BufferConfig     `json:"buffer,omitempty" mapstructure:"buffer" yaml:"buffer,omitempty"`
}

// ClickhouseConfig defines the connection details for connecting Flipt to Clickhouse.
type ClickhouseConfig struct {
	Enabled  bool   `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
	URL      string `json:"url,omitempty" mapstructure:"url" yaml:"url,omitempty"`
	Username string `json:"-" mapstructure:"username" yaml:"username,omitempty"`
	Password string `json:"-" mapstructure:"password" yaml:"password,omitempty"`
}

func (a *AnalyticsConfig) Enabled() bool {
	return a.Clickhouse.Enabled
}

func (c *ClickhouseConfig) Auth() clickhouse.Auth {
	return clickhouse.Auth{
		Username: c.Username,
		Password: c.Password,
		Database: "flipt_analytics",
	}
}

//nolint:unparam
func (a *AnalyticsConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("analytics", map[string]any{
		"enabled": "false",
		"clickhouse": map[string]any{
			"enabled": "false",
			"url":     "",
		},
		"buffer": map[string]any{
			"flush_period": "10s",
		},
	})

	return nil
}

func (a *AnalyticsConfig) validate() error {
	if a.Clickhouse.Enabled && a.Clickhouse.URL == "" {
		return errors.New("clickhouse url not provided")
	}

	if a.Buffer.FlushPeriod < time.Second*10 {
		return errors.New("flush period below 10 seconds")
	}

	return nil
}
