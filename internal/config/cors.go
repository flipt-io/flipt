package config

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
