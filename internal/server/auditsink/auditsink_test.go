package auditsink

import (
	"context"
	"fmt"
	"testing"

	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	"gotest.tools/assert"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type sampleSink struct {
	ch chan AuditEvent
	fmt.Stringer
}

func (s *sampleSink) SendAudits(aes []AuditEvent) error {
	go func() {
		s.ch <- aes[0]
	}()

	return nil
}

func TestSinkSpanExporter(t *testing.T) {
	ss := &sampleSink{ch: make(chan AuditEvent)}
	sse := NewSinkSpanExporter(zap.NewNop(), []AuditSink{ss})

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(sse))

	tr := tp.Tracer("SpanProcessor")

	_, span := tr.Start(context.Background(), "OnStart")

	ae := NewAuditEvent(FlagType, CreateAction, &flipt.CreateFlagRequest{
		Key:         "this-flag",
		Name:        "this-flag",
		Description: "this description",
		Enabled:     false,
	}, "v0.1")

	span.AddEvent("auditEvent", trace.WithAttributes(ae.DecodeToAttributes()...))
	span.End()

	sae := <-ss.ch
	assert.Equal(t, ae.Metadata, sae.Metadata)
	assert.Equal(t, ae.Version, sae.Version)
}
