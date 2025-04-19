package tracing

import (
	"context"
	"sync"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

// newResource constructs a trace resource with Flipt-specific attributes.
// It incorporates schema URL, service name, service version, and OTLP environment data
func newResource(ctx context.Context, fliptVersion string) (*resource.Resource, error) {
	return resource.New(
		ctx,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(
			semconv.ServiceName("flipt"),
			semconv.ServiceVersion(fliptVersion),
		),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithContainer(),
		resource.WithHost(),
		resource.WithProcessRuntimeVersion(),
		resource.WithProcessRuntimeName(),
		resource.WithProcessRuntimeDescription(),
	)
}

// NewProvider creates a new TracerProvider configured for Flipt tracing.
func NewProvider(ctx context.Context, fliptVersion string) (*tracesdk.TracerProvider, error) {
	traceResource, err := newResource(ctx, fliptVersion)
	if err != nil {
		return nil, err
	}

	return tracesdk.NewTracerProvider(
		tracesdk.WithResource(traceResource),
	), nil
}

var (
	traceExpOnce sync.Once
	traceExp     tracesdk.SpanExporter
	traceExpFunc func(context.Context) error = func(context.Context) error { return nil }
	traceExpErr  error
)

// GetExporter retrieves a configured tracesdk.SpanExporter based on the provided configuration.
// Supports Jaeger, Zipkin and OTLP
func GetExporter(ctx context.Context) (tracesdk.SpanExporter, func(context.Context) error, error) {
	traceExpOnce.Do(func() {
		var err error
		traceExp, err = autoexport.NewSpanExporter(ctx)
		if err != nil {
			traceExpErr = err
			return
		}

		traceExpFunc = func(ctx context.Context) error {
			return traceExp.Shutdown(ctx)
		}
	})

	return traceExp, traceExpFunc, traceExpErr
}
