package otel

import (
	"context"
	"sync"

	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var (
	resourceOnce sync.Once
	rsrc         *resource.Resource
	resourceErr  error
)

// newResource constructs a trace resource with Flipt-specific attributes.
// It incorporates schema URL, service name, service version, and OTLP environment data
func NewResource(ctx context.Context, fliptVersion string) (*resource.Resource, error) {
	resourceOnce.Do(func() {
		rsrc, resourceErr = resource.New(
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
	})

	return rsrc, resourceErr
}
