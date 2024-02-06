package analytics

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	evaluationResponsesBytes, err := json.Marshal(evaluationResponses)
	require.NoError(t, err)

	attrs := []attribute.KeyValue{
		{
			Key:   "flipt.evaluation.response",
			Value: attribute.StringValue(string(evaluationResponsesBytes)),
		},
	}
	span.AddEvent("evaluation_response", trace.WithAttributes(attrs...))
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
