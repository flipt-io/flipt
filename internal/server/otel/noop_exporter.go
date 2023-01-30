package otel

import (
	"context"

	"go.opentelemetry.io/otel/sdk/trace"
)

var _ trace.SpanExporter = (*noopSpanExporter)(nil)

type noopSpanExporter struct{}

// NewNoopSpanExporter returns a new noop span exporter.
func NewNoopSpanExporter() trace.SpanExporter {
	return &noopSpanExporter{}
}

func (n *noopSpanExporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	return nil
}

func (n *noopSpanExporter) Shutdown(ctx context.Context) error {
	return nil
}
