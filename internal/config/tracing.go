package config

const (
	// configuration keys
	tracingJaegerEnabled = "tracing.jaeger.enabled"
	tracingJaegerHost    = "tracing.jaeger.host"
	tracingJaegerPort    = "tracing.jaeger.port"
)

// JaegerTracingConfig contains fields, which configure specifically
// Jaeger span and tracing output destination.
type JaegerTracingConfig struct {
	Enabled bool   `json:"enabled,omitempty"`
	Host    string `json:"host,omitempty"`
	Port    int    `json:"port,omitempty"`
}

// TracingConfig contains fields, which configure tracing telemetry
// output destinations.
type TracingConfig struct {
	Jaeger JaegerTracingConfig `json:"jaeger,omitempty"`
}
