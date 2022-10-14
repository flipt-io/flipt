package config

import "github.com/spf13/viper"

// CorsConfig contains fields, which configure behaviour in the
// HTTPServer relating to the CORS header-based mechanisms.
type CorsConfig struct {
	Enabled        bool     `json:"enabled" mapstructure:"enabled"`
	AllowedOrigins []string `json:"allowedOrigins,omitempty" mapstructure:"allowed_origins"`
}

func (c *CorsConfig) viperKey() string {
	return "cors"
}

func (c *CorsConfig) unmarshalViper(v *viper.Viper) (_ []string, _ error) {
	return nil, v.Unmarshal(c)
}
