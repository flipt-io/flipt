package config

import "github.com/spf13/viper"

const (
	corsEnabled        = "cors.enabled"
	corsAllowedOrigins = "cors.allowed_origins"
)

// CorsConfig contains fields, which configure behaviour in the
// HTTPServer relating to the CORS header-based mechanisms.
type CorsConfig struct {
	Enabled        bool     `json:"enabled"`
	AllowedOrigins []string `json:"allowedOrigins,omitempty"`
}

func (c *CorsConfig) init() (_ []string, _ error) {
	if viper.IsSet(corsEnabled) {
		c.Enabled = viper.GetBool(corsEnabled)

		if viper.IsSet(corsAllowedOrigins) {
			c.AllowedOrigins = viper.GetStringSlice(corsAllowedOrigins)
		}
	}

	return
}
