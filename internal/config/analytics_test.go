package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnalytics_WrongDatabase(t *testing.T) {
	clickhouseConfig := ClickhouseConfig{
		Enabled: true,
		URL:     "clickhouse://localhost:9000/wrong_database",
	}

	options, err := clickhouseConfig.Options()
	require.Nil(t, options)

	assert.True(t, errors.Is(err, ErrWrongDatabase))
}
