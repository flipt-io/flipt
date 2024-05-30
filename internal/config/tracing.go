package config

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/viper"
)

var (
	_ defaulter = (*TracingConfig)(nil)
	_ validator = (*TracingConfig)(nil)
)

// TracingConfig contains fields, which configure tracing telemetry
// output destinations.
type TracingConfig struct {
	Enabled       bool                `json:"enabled" mapstructure:"enabled" yaml:"enabled"`
	Exporter      TracingExporter     `json:"exporter,omitempty" mapstructure:"exporter" yaml:"exporter,omitempty"`
	Propagators   []TracingPropagator `json:"propagators,omitempty" mapstructure:"propagators" yaml:"propagators,omitempty"`
	SamplingRatio float64             `json:"samplingRatio,omitempty" mapstructure:"sampling_ratio" yaml:"sampling_ratio,omitempty"`
	Jaeger        JaegerTracingConfig `json:"jaeger,omitempty" mapstructure:"jaeger" yaml:"jaeger,omitempty"`
	Zipkin        ZipkinTracingConfig `json:"zipkin,omitempty" mapstructure:"zipkin" yaml:"zipkin,omitempty"`
	OTLP          OTLPTracingConfig   `json:"otlp,omitempty" mapstructure:"otlp" yaml:"otlp,omitempty"`
}

func (c *TracingConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("tracing", map[string]any{
		"enabled":        false,
		"exporter":       TracingJaeger,
		"sampling_ratio": 1,
		"propagators": []TracingPropagator{
			TracingPropagatorTraceContext,
			TracingPropagatorBaggage,
		},
		"jaeger": map[string]any{
			"host": "localhost",
			"port": 6831,
		},
		"zipkin": map[string]any{
			"endpoint": "http://localhost:9411/api/v2/spans",
		},
		"otlp": map[string]any{
			"endpoint": "localhost:4317",
		},
	})

	return nil
}

func (c *TracingConfig) validate() error {
	if c.SamplingRatio < 0 || c.SamplingRatio > 1 {
		return errors.New("sampling ratio should be a number between 0 and 1")
	}

	for _, propagator := range c.Propagators {
		if !propagator.isValid() {
			return fmt.Errorf("invalid propagator option: %s", propagator)
		}
	}

	return nil
}

func (c *TracingConfig) deprecations(v *viper.Viper) []deprecated {
	var deprecations []deprecated

	if v.GetString("tracing.exporter") == TracingJaeger.String() && v.GetBool("tracing.enabled") {
		deprecations = append(deprecations, "tracing.exporter.jaeger")
	}

	return deprecations
}

// IsZero returns true if the tracing config is not enabled.
// This is used for marshalling to YAML for `config init`.
func (c TracingConfig) IsZero() bool {
	return !c.Enabled
}

// TracingExporter represents the supported tracing exporters.
// TODO: can we use a string here instead?
type TracingExporter uint8

func (e TracingExporter) String() string {
	return tracingExporterToString[e]
}

func (e TracingExporter) MarshalJSON() ([]byte, error) {
	return json.Marshal(e.String())
}

func (e TracingExporter) MarshalYAML() (interface{}, error) {
	return e.String(), nil
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

type TracingPropagator string

const (
	TracingPropagatorTraceContext TracingPropagator = "tracecontext"
	TracingPropagatorBaggage      TracingPropagator = "baggage"
	TracingPropagatorB3           TracingPropagator = "b3"
	TracingPropagatorB3Multi      TracingPropagator = "b3multi"
	TracingPropagatorJaeger       TracingPropagator = "jaeger"
	TracingPropagatorXRay         TracingPropagator = "xray"
	TracingPropagatorOtTrace      TracingPropagator = "ottrace"
	TracingPropagatorNone         TracingPropagator = "none"
)

func (t TracingPropagator) isValid() bool {
	validOptions := map[TracingPropagator]bool{
		TracingPropagatorTraceContext: true,
		TracingPropagatorBaggage:      true,
		TracingPropagatorB3:           true,
		TracingPropagatorB3Multi:      true,
		TracingPropagatorJaeger:       true,
		TracingPropagatorXRay:         true,
		TracingPropagatorOtTrace:      true,
		TracingPropagatorNone:         true,
	}

	return validOptions[t]
}

// JaegerTracingConfig contains fields, which configure
// Jaeger span and tracing output destination.
type JaegerTracingConfig struct {
	Host string `json:"host,omitempty" mapstructure:"host" yaml:"host,omitempty"`
	Port int    `json:"port,omitempty" mapstructure:"port" yaml:"port,omitempty"`
}

// ZipkinTracingConfig contains fields, which configure
// Zipkin span and tracing output destination.
type ZipkinTracingConfig struct {
	Endpoint string `json:"endpoint,omitempty" mapstructure:"endpoint" yaml:"endpoint,omitempty"`
}

// OTLPTracingConfig contains fields, which configure
// OTLP span and tracing output destination.
type OTLPTracingConfig struct {
	Endpoint string            `json:"endpoint,omitempty" mapstructure:"endpoint" yaml:"endpoint,omitempty"`
	Headers  map[string]string `json:"headers,omitempty" mapstructure:"headers" yaml:"headers,omitempty"`
}
