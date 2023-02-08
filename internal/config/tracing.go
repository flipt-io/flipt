package config

import (
	"encoding/json"

	"github.com/spf13/viper"
)

// cheers up the unparam linter
var _ defaulter = (*TracingConfig)(nil)

// TracingConfig contains fields, which configure tracing telemetry
// output destinations.
type TracingConfig struct {
	Enabled  bool                `json:"enabled,omitempty" mapstructure:"enabled"`
	Exporter TracingExporter     `json:"exporter,omitempty" mapstructure:"exporter"`
	Jaeger   JaegerTracingConfig `json:"jaeger,omitempty" mapstructure:"jaeger"`
	Zipkin   ZipkinTracingConfig `json:"zipkin,omitempty" mapstructure:"zipkin"`
	OTLP     OTLPTracingConfig   `json:"otlp,omitempty" mapstructure:"otlp"`
}

func (c *TracingConfig) setDefaults(v *viper.Viper) {
	v.SetDefault("tracing", map[string]any{
		"enabled":  false,
		"exporter": TracingJaeger,
		"jaeger": map[string]any{
			"enabled": false, // deprecated (see below)
			"host":    "localhost",
			"port":    6831,
		},
		"zipkin": map[string]any{
			"endpoint": "http://localhost:9411/api/v2/spans",
		},
		"otlp": map[string]any{
			"endpoint": "localhost:4317",
		},
	})

	if v.GetBool("tracing.jaeger.enabled") {
		// forcibly set top-level `enabled` to true
		v.Set("tracing.enabled", true)
		v.Set("tracing.exporter", TracingJaeger)
	}
}

func (c *TracingConfig) deprecations(v *viper.Viper) []deprecation {
	var deprecations []deprecation

	if v.InConfig("tracing.jaeger.enabled") {
		deprecations = append(deprecations, deprecation{
			option:            "tracing.jaeger.enabled",
			additionalMessage: deprecatedMsgTracingJaegerEnabled,
		})
	}

	return deprecations
}

// TracingExporter represents the supported tracing exporters.
type TracingExporter uint8

func (e TracingExporter) String() string {
	return tracingExporterToString[e]
}

func (e TracingExporter) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

const (
	_ TracingExporter = iota
	// TracingJaeger ...
	TracingJaeger
	// TracingZipkin ...
	TracingZipkin
	// TracingOTLP ...
	TracingOTLP
)

var (
	tracingExporterToString = map[TracingExporter]string{
		TracingJaeger: "jaeger",
		TracingZipkin: "zipkin",
		TracingOTLP:   "otlp",
	}

	stringToTracingExporter = map[string]TracingExporter{
		"jaeger": TracingJaeger,
		"zipkin": TracingZipkin,
		"otlp":   TracingOTLP,
	}
)

// JaegerTracingConfig contains fields, which configure
// Jaeger span and tracing output destination.
type JaegerTracingConfig struct {
	Host string `json:"host,omitempty" mapstructure:"host"`
	Port int    `json:"port,omitempty" mapstructure:"port"`
}

// ZipkinTracingConfig contains fields, which configure
// Zipkin span and tracing output destination.
type ZipkinTracingConfig struct {
	Endpoint string `json:"endpoint,omitempty" mapstructure:"endpoint"`
}

// OTLPTracingConfig contains fields, which configure
// OTLP span and tracing output destination.
type OTLPTracingConfig struct {
	Endpoint string `json:"endpoint,omitempty" mapstructure:"endpoint"`
}
