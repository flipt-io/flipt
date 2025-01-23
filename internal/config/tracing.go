package config

import (
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
	Propagators   []TracingPropagator `json:"propagators,omitempty" mapstructure:"propagators" yaml:"propagators,omitempty"`
	SamplingRatio float64             `json:"samplingRatio,omitempty" mapstructure:"sampling_ratio" yaml:"sampling_ratio,omitempty"`
	OTLP          OTLPTracingConfig   `json:"otlp,omitempty" mapstructure:"otlp" yaml:"otlp,omitempty"`
}

func (c *TracingConfig) setDefaults(v *viper.Viper) error {
	v.SetDefault("tracing", map[string]any{
		"enabled":        false,
		"sampling_ratio": 1,
		"propagators": []TracingPropagator{
			TracingPropagatorTraceContext,
			TracingPropagatorBaggage,
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

// IsZero returns true if the tracing config is not enabled.
// This is used for marshalling to YAML for `config init`.
func (c TracingConfig) IsZero() bool {
	return !c.Enabled
}

// TracingExporter represents the supported tracing exporters.
type TracingExporter string

const (
	TracingOTLP TracingExporter = "otlp"
)

func (e TracingExporter) String() string {
	return string(e)
}

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
