package logs

import (
	"context"
	"sync"

	"go.opentelemetry.io/contrib/exporters/autoexport"
	logsdk "go.opentelemetry.io/otel/sdk/log"
)

var (
	logExpOnce sync.Once
	logExp     logsdk.Exporter
	logExpFunc func(context.Context) error = func(context.Context) error { return nil }
	logExpErr  error
)

// GetExporter retrieves a configured logsdk.Exporter based on the provided configuration.
// Supports OTLP
func GetExporter(ctx context.Context) (logsdk.Exporter, func(context.Context) error, error) {
	logExpOnce.Do(func() {
		fallbackExporterFactory := func(ctx context.Context) (logsdk.Exporter, error) {
			return noopLogExporter{}, nil
		}

		logExp, logExpErr = autoexport.NewLogExporter(ctx, autoexport.WithFallbackLogExporter(fallbackExporterFactory))
		if logExpErr != nil {
			return
		}

		logExpFunc = func(ctx context.Context) error {
			return logExp.Shutdown(ctx)
		}
	})

	return logExp, logExpFunc, logExpErr
}

// noopLogExporter is an implementation of logsdk.Exporter that performs no operations.
type noopLogExporter struct{}

var _ logsdk.Exporter = noopLogExporter{}

// ExportSpans is part of logsdk.Exporter interface.
func (e noopLogExporter) Export(ctx context.Context, records []logsdk.Record) error {
	return nil
}

// Shutdown is part of log.Exporter interface.
func (e noopLogExporter) Shutdown(ctx context.Context) error {
	return nil
}

// ForceFlush is part of log.Exporter interface.
func (e noopLogExporter) ForceFlush(ctx context.Context) error {
	return nil
}
