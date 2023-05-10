package audit

import (
	"context"
	"fmt"
	"testing"
	"time"

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
	cases := []struct {
		name      string
		fType     Type
		action    Action
		expectErr bool
	}{
		{
			name:      "Valid",
			fType:     FlagType,
			action:    Create,
			expectErr: false,
		},
		{
			name:      "Invalid",
			fType:     Type(""),
			action:    Action(""),
			expectErr: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx := context.Background()

			ss := &sampleSink{ch: make(chan Event)}
			sse := NewSinkSpanExporter(zap.NewNop(), []Sink{ss})

			defer func() {
				_ = sse.Shutdown(ctx)
			}()

			tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
			tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(sse))

			tr := tp.Tracer("SpanProcessor")

			_, span := tr.Start(ctx, "OnStart")

			e := NewEvent(
				c.fType,
				c.action,
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

			timeoutCtx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
			defer cancel()

			select {
			case se := <-ss.ch:
				assert.Equal(t, e.Metadata, se.Metadata)
				assert.Equal(t, e.Version, se.Version)
			case <-timeoutCtx.Done():
				if !c.expectErr {
					assert.Fail(t, "send audits should have been called")
				}
			}
		})
	}
}
func TestGRPCMethodToAction(t *testing.T) {
	a := GRPCMethodToAction("CreateNamespace")
	assert.Equal(t, Create, a)

	a = GRPCMethodToAction("UpdateSegment")
	assert.Equal(t, Update, a)

	a = GRPCMethodToAction("NoMethodMatched")
	assert.Equal(t, "", string(a))
}
