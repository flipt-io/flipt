package config

import "github.com/spf13/viper"

// cheers up the unparam linter
var _ defaulter = (*CorsConfig)(nil)

// CorsConfig contains fields, which configure behaviour in the
// HTTPServer relating to the CORS header-based mechanisms.
type CorsConfig struct {
	Enabled        bool     `json:"enabled" mapstructure:"enabled"`
	AllowedOrigins []string `json:"allowedOrigins,omitempty" mapstructure:"allowed_origins"`
}

func (c *CorsConfig) setDefaults(v *viper.Viper) []string {
	v.SetDefault("cors", map[string]any{
		"allowed_origins": "*",
	})

	return nil
}
