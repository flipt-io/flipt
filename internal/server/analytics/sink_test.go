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
	go func() {
		s.ch <- responses
	}()
	return nil
}

func (s *sampleSink) Close() error { return nil }

func TestSinkSpanExporter(t *testing.T) {
	ss := &sampleSink{
		ch: make(chan []*EvaluationResponse),
	}
	analyticsSinkSpanExporter := NewAnalyticsSinkSpanExporter(zap.NewNop(), ss)

	tp := sdktrace.NewTracerProvider(sdktrace.WithSampler(sdktrace.AlwaysSample()))
	tp.RegisterSpanProcessor(sdktrace.NewSimpleSpanProcessor(analyticsSinkSpanExporter))

	tr := tp.Tracer("SpanProcesser")

	_, span := tr.Start(context.Background(), "OnStart")

	evaluationResponse := &EvaluationResponse{
		FlagKey:      "hello",
		NamespaceKey: "default",
		Reason:       "MATCH_EVALUATION_REASON",
		Match:        true,
		Timestamp:    time.Now().UTC(),
	}

	evaluationResponseBytes, err := json.Marshal(evaluationResponse)
	require.NoError(t, err)

	attrs := []attribute.KeyValue{
		{
			Key:   "flipt.evaluation.response",
			Value: attribute.StringValue(string(evaluationResponseBytes)),
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
		assert.Equal(t, evaluationResponse.FlagKey, evaluationResponseActual.FlagKey)
		assert.Equal(t, evaluationResponse.NamespaceKey, evaluationResponseActual.NamespaceKey)
		assert.Equal(t, evaluationResponse.Reason, evaluationResponseActual.Reason)
		assert.Equal(t, evaluationResponse.Match, evaluationResponseActual.Match)
		assert.Equal(t, evaluationResponse.Timestamp, evaluationResponseActual.Timestamp)
	case <-timeoutCtx.Done():
		require.Fail(t, "message should have been sent on the channel")
	}
}
