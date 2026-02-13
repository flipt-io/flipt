package config

import (
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/spf13/viper"
)

var (
	_ defaulter = (*AnalyticsConfig)(nil)
	_ validator = (*AnalyticsConfig)(nil)
)

// AnalyticsStorageConfig is a collection of configuration option for storage backends.
type AnalyticsStorageConfig struct {
	Clickhouse ClickhouseConfig `json:"clickhouse,omitempty" mapstructure:"clickhouse" yaml:"clickhouse,omitempty"`
	Prometheus PrometheusConfig `json:"prometheus,omitempty" mapstructure:"prometheus" yaml:"prometheus,omitempty"`
}

func (a AnalyticsStorageConfig) String() string {
	// TODO: make this more dynamic if we add more storage options
	if a.Clickhouse.Enabled {
		return "clickhouse"
	}

	if a.Prometheus.Enabled {
		return "prometheus"
	}
	return ""
}

// ClickhouseConfig defines the connection details for connecting Flipt to Clickhouse.
type ClickhouseConfig struct {
	Enabled bool   `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
	URL     string `json:"-" mapstructure:"url" yaml:"url,omitempty"`
}

// Options returns the connection option details for Clickhouse.
func (c ClickhouseConfig) Options() (*clickhouse.Options, error) {
	options, err := clickhouse.ParseDSN(c.URL)
	if err != nil {
		return nil, err
	}

	return options, nil
}

// PrometheusSigV4Config defines the SigV4 authentication configuration for Prometheus.
type PrometheusSigV4Config struct {
	Enabled            bool   `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
	Region             string `json:"-" mapstructure:"region" yaml:"-"`
	AccessKey          string `json:"-" mapstructure:"access_key" yaml:"-"`
	SecretKey          string `json:"-" mapstructure:"secret_key" yaml:"-"`
	Profile            string `json:"-" mapstructure:"profile" yaml:"-"`
	RoleARN            string `json:"-" mapstructure:"role_arn" yaml:"-"`
	ExternalID         string `json:"-" mapstructure:"external_id" yaml:"-"`
	UseFIPSSTSEndpoint bool   `json:"-" mapstructure:"use_fips_sts_endpoint" yaml:"-"`
	ServiceName        string `json:"-" mapstructure:"service_name" yaml:"-"`
}

// PrometheusConfig defines the connection details for connecting Flipt to Prometheus.
type PrometheusConfig struct {
	Enabled bool                  `json:"enabled,omitempty" mapstructure:"enabled" yaml:"enabled,omitempty"`
	URL     string                `json:"-" mapstructure:"url" yaml:"-"`
	Headers map[string]string     `json:"-" mapstructure:"headers" yaml:"-"`
	SigV4   PrometheusSigV4Config `json:"-" mapstructure:"sigv4" yaml:"-"`
}

// AnalyticsBufferConfig holds configuration for the buffering of sending the analytics
// events to the backend storage.
type AnalyticsBufferConfig struct {
	Capacity    int           `json:"capacity,omitempty" mapstructure:"capacity" yaml:"capacity,omitempty"`
	FlushPeriod time.Duration `json:"flushPeriod,omitempty" mapstructure:"flush_period" yaml:"flush_period,omitempty"`
}

// AnalyticsConfig defines the configuration for various mechanisms for
// reporting and querying analytical data for Flipt.
type AnalyticsConfig struct {
	Storage AnalyticsStorageConfig `json:"storage,omitempty" mapstructure:"storage" yaml:"storage,omitempty"`
	Buffer  AnalyticsBufferConfig  `json:"buffer,omitempty" mapstructure:"buffer" yaml:"buffer,omitempty"`
}

func (a *AnalyticsConfig) Enabled() bool {
	return a.Storage.Clickhouse.Enabled || a.Storage.Prometheus.Enabled
}

// IsZero returns true if the analytics config is not enabled.
// This is used for marshalling to YAML for `config init`.
func (a AnalyticsConfig) IsZero() bool {
	return !a.Enabled()
}

//nolint:unparam
func (a *AnalyticsConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("analytics.storage.clickhouse.enabled", "false")
	v.SetDefault("analytics.storage.clickhouse.url", "")
	v.SetDefault("analytics.storage.prometheus.enabled", "false")
	v.SetDefault("analytics.storage.prometheus.url", "")
	v.SetDefault("analytics.buffer.flush_period", "10s")
	return nil
}

func (a *AnalyticsConfig) validate() error {
	if a.Storage.Clickhouse.Enabled {
		if a.Storage.Clickhouse.URL == "" {
			return errFieldRequired("analytics", "clickhouse url")
		}

		if a.Buffer.FlushPeriod < time.Second*10 {
			return errString("analytics", "buffer flush period below 10 seconds")
		}
	}

	return nil
}
