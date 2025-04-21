package tracing

import (
	"context"
	"sync"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
)

var (
	traceExpOnce sync.Once
	traceExp     tracesdk.SpanExporter
	traceExpFunc func(context.Context) error = func(context.Context) error { return nil }
	traceExpErr  error
)

// GetExporter retrieves a configured tracesdk.SpanExporter based on the provided configuration.
// Supports OTLP
func GetExporter(ctx context.Context) (tracesdk.SpanExporter, func(context.Context) error, error) {
	traceExpOnce.Do(func() {
		traceExp, traceExpErr = autoexport.NewSpanExporter(ctx)
		if traceExpErr != nil {
			return
		}

		traceExpFunc = func(ctx context.Context) error {
			return traceExp.Shutdown(ctx)
		}
	})

	return traceExp, traceExpFunc, traceExpErr
}
