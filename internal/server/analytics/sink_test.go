package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/tracing"
	"go.flipt.io/flipt/rpc/v2/evaluation"
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

	_, span := tr.Start(t.Context(), "OnStart")

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
		{Key: "flipt.key", Value: attribute.StringValue("hello")},
		{Key: "flipt.type", Value: attribute.StringValue("variant")},
		{Key: "flipt.namespace", Value: attribute.StringValue("default")},
		{Key: "flipt.result.reason", Value: attribute.StringValue("match")},
		{Key: "flipt.result.match", Value: attribute.BoolValue(b)},
	}
	span.AddEvent(tracing.Event, trace.WithAttributes(attrs...), trace.WithTimestamp(evaluationResponses[0].Timestamp))
	span.End()

	timeoutCtx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
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
		assert.Equal(t, evaluation.EvaluationFlagType_VARIANT_FLAG_TYPE.String(), evaluationResponseActual.FlagType)
	case <-timeoutCtx.Done():
		require.Fail(t, "message should have been sent on the channel")
	}
}

func TestTransformSpanEvent_BooleanEvaluationValue(t *testing.T) {
	const now = 1700000000

	tests := []struct {
		name     string
		variant  attribute.KeyValue
		expected string
	}{
		{
			name:     "enabled",
			variant:  tracing.AttributeVariant.Bool(true),
			expected: "true",
		},
		{
			name:     "disabled",
			variant:  tracing.AttributeVariant.Bool(false),
			expected: "false",
		},
		{
			name:     "variant string unchanged",
			variant:  tracing.AttributeVariant.String("variant-key"),
			expected: "variant-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event := sdktrace.Event{
				Name: tracing.Event,
				Time: time.Unix(now, 0),
				Attributes: []attribute.KeyValue{
					tracing.AttributeFlag.String("hello"),
					tracing.AttributeFlagTypeBoolean,
					tracing.AttributeNamespace.String("default"),
					tracing.AttributeReason.String("match"),
					tt.variant,
				},
			}

			responses, err := transformSpanEventToEvaluationResponses(event)
			require.NoError(t, err)
			require.Len(t, responses, 1)

			require.NotNil(t, responses[0].EvaluationValue)
			assert.Equal(t, tt.expected, *responses[0].EvaluationValue)
		})
	}
}
