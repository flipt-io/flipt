package config

import (
	"errors"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/spf13/viper"
)

var (
	_ defaulter = (*AnalyticsConfig)(nil)
	_ validator = (*AnalyticsConfig)(nil)
)

// AnalyticsConfig defines the configuration for various mechanisms for
// reporting and querying analytical data for Flipt.
type AnalyticsConfig struct {
	Storage AnalyticsStorageConfig `json:"storage,omitempty" mapstructure:"storage" yaml:"storage,omitempty"`
	Buffer  BufferConfig           `json:"buffer,omitempty" mapstructure:"buffer" yaml:"buffer,omitempty"`
}

// AnalyticsStorageConfig is a collection of configuration option for storage backends.
type AnalyticsStorageConfig struct {
	Clickhouse ClickhouseConfig `json:"clickhouse,omitempty" mapstructure:"clickhouse" yaml:"clickhouse,omitempty"`
}

func (a *AnalyticsStorageConfig) String() string {
	// TODO: make this more dynamic if we add more storage options
	if a.Clickhouse.Enabled {
		return "clickhouse"
	}
	return ""
}

// ClickhouseConfig defines the connection details for connecting Flipt to Clickhouse.
type ClickhouseConfig struct {
	Enabled bool   `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
	URL     string `json:"-" mapstructure:"url" yaml:"url,omitempty"`
}

func (a *AnalyticsConfig) Enabled() bool {
	return a.Storage.Clickhouse.Enabled
}

// Options returns the connection option details for Clickhouse.
func (c *ClickhouseConfig) Options() (*clickhouse.Options, error) {
	options, err := clickhouse.ParseDSN(c.URL)
	if err != nil {
		return nil, err
	}

	return options, nil
}

//nolint:unparam
func (a *AnalyticsConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("analytics", map[string]any{
		"storage": map[string]any{
			"clickhouse": map[string]any{
				"enabled": "false",
				"url":     "",
			},
		},
		"buffer": map[string]any{
			"flush_period": "10s",
		},
	})

	return nil
}

func (a *AnalyticsConfig) validate() error {
	if a.Storage.Clickhouse.Enabled && a.Storage.Clickhouse.URL == "" {
		return errors.New("clickhouse url not provided")
	}

	if a.Buffer.FlushPeriod < time.Second*10 {
		return errors.New("flush period below 10 seconds")
	}

	return nil
}
