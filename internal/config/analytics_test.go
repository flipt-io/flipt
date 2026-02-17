package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalyticsClickhouseConfiguration(t *testing.T) {
	cfg := &AnalyticsConfig{
		Storage: AnalyticsStorageConfig{
			Clickhouse: ClickhouseConfig{
				Enabled: true,
				URL:     "clickhouse://localhost/db",
			},
		},
	}
	assert.True(t, cfg.Enabled())
	options, err := cfg.Storage.Clickhouse.Options()
	require.NoError(t, err)
	assert.Equal(t, []string{"localhost"}, options.Addr)
	assert.Equal(t, "db", options.Auth.Database)

	cfg.Storage.Clickhouse.URL = "something"
	_, err = cfg.Storage.Clickhouse.Options()
	require.Error(t, err)
	assert.ErrorContains(t, err, "parse dsn address failed")
}

func TestAnalyticsPrometheusConfiguration(t *testing.T) {
	cfg := &AnalyticsConfig{
		Storage: AnalyticsStorageConfig{
			Prometheus: PrometheusConfig{
				Enabled: true,
				URL:     "http://localhost:9090",
			},
		},
	}
	assert.True(t, cfg.Enabled())
	assert.Equal(t, "prometheus", cfg.Storage.String())
}

func TestAnalyticsPrometheusSigV4Configuration(t *testing.T) {
	cfg := &AnalyticsConfig{
		Storage: AnalyticsStorageConfig{
			Prometheus: PrometheusConfig{
				Enabled: true,
				URL:     "http://localhost:9090",
				SigV4: PrometheusSigV4Config{
					Enabled:     true,
					Region:      "us-east-1",
					AccessKey:   "AKIAIOSFODNN7EXAMPLE",
					SecretKey:   "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
					ServiceName: "aps",
				},
			},
		},
	}
	assert.True(t, cfg.Enabled())
	assert.True(t, cfg.Storage.Prometheus.SigV4.Enabled)
	assert.Equal(t, "us-east-1", cfg.Storage.Prometheus.SigV4.Region)
	assert.Equal(t, "aps", cfg.Storage.Prometheus.SigV4.ServiceName)
}
