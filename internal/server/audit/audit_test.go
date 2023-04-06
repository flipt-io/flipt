package audit

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type sampleSink struct {
	ch chan Event
	fmt.Stringer
}

func (s *sampleSink) SendAudits(aes []Event) error {
	go func() {
		s.ch <- aes[0]
	}()

	return nil
}

func (s *sampleSink) Close() error { return nil }

func TestSinkSpanExporter(t *testing.T) {
	ss := &sampleSink{ch: make(chan Event)}
	sse := NewSinkSpanExporter(zap.NewNop(), []Sink{ss})

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(sse))

	tr := tp.Tracer("SpanProcessor")

	_, span := tr.Start(context.Background(), "OnStart")

	ae := NewEvent(Flag, Create, &flipt.CreateFlagRequest{
		Key:         "this-flag",
		Name:        "this-flag",
		Description: "this description",
		Enabled:     false,
	})

	span.AddEvent("auditEvent", trace.WithAttributes(ae.DecodeToAttributes()...))
	span.End()

	sae := <-ss.ch
	assert.Equal(t, ae.Metadata, sae.Metadata)
	assert.Equal(t, ae.Version, sae.Version)
}
