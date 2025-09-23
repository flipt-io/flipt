package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/tracing"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type sampleSink struct {
	ch chan []*EvaluationResponse
}

func (s *sampleSink) IncrementFlagEvaluationCounts(ctx context.Context, responses []*EvaluationResponse) error {
	s.ch <- responses
	return nil
}

func (s *sampleSink) Close() error { return nil }

func TestSinkSpanExporter(t *testing.T) {
	ss := &sampleSink{
		ch: make(chan []*EvaluationResponse, 1),
	}
	analyticsSinkSpanExporter := NewAnalyticsSinkSpanExporter(zap.NewNop(), ss)

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(analyticsSinkSpanExporter))

	tr := tp.Tracer("SpanProcesser")

	_, span := tr.Start(context.Background(), "OnStart")

	b := true
	evaluationResponses := []*EvaluationResponse{
		{
			FlagKey:      "hello",
			NamespaceKey: "default",
			Reason:       "MATCH_EVALUATION_REASON",
			Match:        &b,
			Timestamp:    time.Now().UTC(),
		},
	}

	attrs := []attribute.KeyValue{
		{Key: "flipt_flag", Value: attribute.StringValue("hello")},
		{Key: "flipt_namespace", Value: attribute.StringValue("default")},
		{Key: "flipt_reason", Value: attribute.StringValue("MATCH_EVALUATION_REASON")},
		{Key: "flipt_match", Value: attribute.BoolValue(b)},
	}
	span.AddEvent(tracing.Event, trace.WithAttributes(attrs...), trace.WithTimestamp(evaluationResponses[0].Timestamp))
	span.End()

	timeoutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	select {
	case e := <-ss.ch:
		assert.Len(t, e, 1)
		evaluationResponseActual := e[0]
		assert.Equal(t, evaluationResponses[0].FlagKey, evaluationResponseActual.FlagKey)
		assert.Equal(t, evaluationResponses[0].NamespaceKey, evaluationResponseActual.NamespaceKey)
		assert.Equal(t, evaluationResponses[0].Reason, evaluationResponseActual.Reason)
		assert.Equal(t, *evaluationResponses[0].Match, *evaluationResponseActual.Match)
		assert.Equal(t, evaluationResponses[0].Timestamp, evaluationResponseActual.Timestamp)
	case <-timeoutCtx.Done():
		require.Fail(t, "message should have been sent on the channel")
	}
}
