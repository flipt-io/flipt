package audit

import (
	"context"
	"fmt"
	"testing"
	"time"

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

func (s *sampleSink) SendAudits(ctx context.Context, es []Event) error {
	go func() {
		s.ch <- es[0]
	}()

	return nil
}

func (s *sampleSink) Close() error { return nil }

func TestSinkSpanExporter(t *testing.T) {
	cases := []struct {
		name      string
		resource  flipt.Resource
		subject   flipt.Subject
		action    flipt.Action
		expectErr bool
	}{
		{
			name:      "Valid",
			resource:  flipt.ResourceFlag,
			subject:   flipt.SubjectFlag,
			action:    flipt.ActionCreate,
			expectErr: false,
		},
		{
			name:      "Valid Rule",
			resource:  flipt.ResourceFlag,
			subject:   flipt.SubjectRule,
			action:    flipt.ActionCreate,
			expectErr: false,
		},
		{
			name:      "Invalid",
			subject:   flipt.Subject(""),
			action:    flipt.Action(""),
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

			r := flipt.NewRequest(c.resource, c.action, flipt.WithSubject(c.subject))
			e := NewEvent(
				r,
				&Actor{
					Authentication: "token",
					IP:             "127.0.0.1",
				}, &Flag{
					Key:         "this-flag",
					Name:        "this-flag",
					Description: "this description",
					Enabled:     false,
				})

			span.AddEvent("event", trace.WithAttributes(e.DecodeToAttributes()...))
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
