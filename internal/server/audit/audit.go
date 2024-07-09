package audit

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/hashicorp/go-multierror"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

// Sink is the abstraction for various audit sink configurations
// that Flipt will support.
type Sink interface {
	SendAudits(context.Context, []Event) error
	io.Closer
	fmt.Stringer
}

// SinkSpanExporter sends audit logs to configured sinks through intercepting span events.
type SinkSpanExporter struct {
	sinks  []Sink
	logger *zap.Logger
}

// EventExporter provides an API for exporting spans as Event(s).
type EventExporter interface {
	ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error
	Shutdown(ctx context.Context) error
	SendAudits(ctx context.Context, es []Event) error
}

// NewSinkSpanExporter is the constructor for a SinkSpanExporter.
func NewSinkSpanExporter(logger *zap.Logger, sinks []Sink) EventExporter {
	return &SinkSpanExporter{
		sinks:  sinks,
		logger: logger,
	}
}

// ExportSpans completes one part of the implementation of a SpanExporter. Decodes span events to audit events.
func (s *SinkSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	es := make([]Event, 0)

	for _, span := range spans {
		for _, e := range span.Events() {
			e, err := decodeToEvent(e.Attributes)
			if err != nil {
				if !errors.Is(err, errEventNotValid) {
					s.logger.Error("audit event not decodable", zap.Error(err))
				}
				continue
			}
			es = append(es, *e)
		}
	}

	return s.SendAudits(ctx, es)
}

// Shutdown will close all the registered sinks.
func (s *SinkSpanExporter) Shutdown(ctx context.Context) error {
	var result error

	for _, sink := range s.sinks {
		err := sink.Close()
		if err != nil {
			result = multierror.Append(result, err)
		}
	}

	return result
}

// SendAudits wraps the methods of sending audits events to various sinks.
func (s *SinkSpanExporter) SendAudits(ctx context.Context, es []Event) error {
	if len(es) < 1 {
		return nil
	}

	for _, sink := range s.sinks {
		s.logger.Debug("performing batched sending of audit events", zap.Stringer("sink", sink), zap.Int("batch size", len(es)))
		err := sink.SendAudits(ctx, es)
		if err != nil {
			s.logger.Error("failed to send audits to sink", zap.Stringer("sink", sink), zap.Error(err))
		}
	}

	return nil
}
