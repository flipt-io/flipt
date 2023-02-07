package otel

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// TracerProvider is an interface that wraps the trace.TracerProvider and adds the Shutdown method
// that is not part of the interface but is part of the implementation.
type TracerProvider interface {
	trace.TracerProvider
	Shutdown(context.Context) error
}

type noopProvider struct {
	trace.TracerProvider
}

func NewNoopProvider() TracerProvider {
	return &noopProvider{
		TracerProvider: trace.NewNoopTracerProvider(),
	}
}

func (p noopProvider) Shutdown(context.Context) error {
	return nil
}
