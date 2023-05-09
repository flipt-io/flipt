package audit

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type sampleSink struct {
	ch chan Event
	fmt.Stringer
}

func (s *sampleSink) SendAudits(es []Event) error {
	go func() {
		s.ch <- es[0]
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

	e := NewEvent(
		FlagType,
		Create,
		map[string]string{
			"authentication": "token",
			"ip":             "127.0.0.1",
		}, &Flag{
			Key:         "this-flag",
			Name:        "this-flag",
			Description: "this description",
			Enabled:     false,
		})

	span.AddEvent("auditEvent", trace.WithAttributes(e.DecodeToAttributes()...))
	span.End()

	se := <-ss.ch
	assert.Equal(t, e.Metadata, se.Metadata)
	assert.Equal(t, e.Version, se.Version)
}

func TestGRPCMethodToAction(t *testing.T) {
	a := GRPCMethodToAction("CreateNamespace")
	assert.Equal(t, Create, a)

	a = GRPCMethodToAction("UpdateSegment")
	assert.Equal(t, Update, a)
}
