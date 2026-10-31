package config

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
)

func TestCorsDefaultsAllowFliptReferenceHeader(t *testing.T) {
	const referenceHeader = "X-Flipt-Reference"

	require.Contains(t, Default().Cors.AllowedHeaders, referenceHeader)

	v := viper.New()
	require.NoError(t, (&CorsConfig{}).setDefaults(v))
	require.Contains(t, v.GetStringSlice("cors.allowed_headers"), referenceHeader)
}
