package config

import "github.com/spf13/viper"

var (
	_ defaulter = (*CorsConfig)(nil)
)

// CorsConfig contains fields, which configure behaviour in the
// HTTPServer relating to the CORS header-based mechanisms.
type CorsConfig struct {
	Enabled        bool     `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
	AllowedOrigins []string `json:"allowedOrigins,omitempty" mapstructure:"allowed_origins" yaml:"allowed_origins,omitempty"`
	AllowedHeaders []string `json:"allowedHeaders,omitempty" mapstructure:"allowed_headers" yaml:"allowed_headers,omitempty"`
}

func (c *CorsConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("cors", map[string]any{
		"enabled":         false,
		"allowed_origins": "*",
		"allowed_headers": []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Fern-Language",
			"X-Fern-SDK-Name",
			"X-Fern-SDK-Version",
			"X-Flipt-Namespace",
			"X-Flipt-Accept-Server-Version",
		},
	})

	return nil
}
