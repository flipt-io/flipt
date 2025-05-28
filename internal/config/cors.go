package config

import (
	"github.com/spf13/viper"
	"go.flipt.io/flipt/internal/server/common"
)

var _ defaulter = (*CorsConfig)(nil)

// CorsConfig contains fields, which configure behaviour in the
// HTTPServer relating to the CORS header-based mechanisms.
type CorsConfig struct {
	Enabled        bool     `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
	AllowedOrigins []string `json:"allowedOrigins,omitempty" mapstructure:"allowed_origins" yaml:"allowed_origins,omitempty"`
	AllowedHeaders []string `json:"allowedHeaders,omitempty" mapstructure:"allowed_headers" yaml:"allowed_headers,omitempty"`
}

func (c *CorsConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("cors.enabled", false)
	v.SetDefault("cors.allowed_origins", "*")
	v.SetDefault("cors.allowed_headers", []string{
		"Accept",
		"Authorization",
		"Content-Type",
		"X-CSRF-Token",
		common.HeaderFliptEnvironment,
		common.HeaderFliptNamespace,
	})
	return nil
}

// IsZero returns true if the cors config is not enabled.
// This is used for marshalling to YAML for `config init`.
func (c CorsConfig) IsZero() bool {
	return !c.Enabled
}
