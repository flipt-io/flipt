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
