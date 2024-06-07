package analytics

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
)

var errNotTransformable = errors.New("event not transformable into evaluation response")

type AnalyticsStoreMutator interface {
	IncrementFlagEvaluationCounts(ctx context.Context, responses []*EvaluationResponse) error
	Close() error
}

type EvaluationResponse struct {
	FlagKey         string    `json:"flagKey,omitempty"`
	FlagType        string    `json:"flagType,omitempty"`
	NamespaceKey    string    `json:"namespaceKey,omitempty"`
	Reason          string    `json:"reason,omitempty"`
	Match           *bool     `json:"match,omitempty"`
	EvaluationValue *string   `json:"evaluationValue,omitempty"`
	Timestamp       time.Time `json:"timestamp,omitempty"`
	EntityId        string    `json:"entityId,omitempty"`
}

// AnalyticsSinkSpanExporter implements SpanExporter.
type AnalyticsSinkSpanExporter struct {
	logger                *zap.Logger
	analyticsStoreMutator AnalyticsStoreMutator
}

// NewAnalyticsSinkSpanExporter is the constructor function for an AnalyticsSpanExporter.
func NewAnalyticsSinkSpanExporter(logger *zap.Logger, analyticsStoreMutator AnalyticsStoreMutator) *AnalyticsSinkSpanExporter {
	return &AnalyticsSinkSpanExporter{
		logger:                logger,
		analyticsStoreMutator: analyticsStoreMutator,
	}
}

const evaluationResponseKey = "flipt.evaluation.response"

// transformSpanEventToEvaluationResponses is a convenience function to transform a span event into an []*EvaluationResponse.
func transformSpanEventToEvaluationResponses(event sdktrace.Event) ([]*EvaluationResponse, error) {
	for _, attr := range event.Attributes {
		if string(attr.Key) == evaluationResponseKey {
			evaluationResponseBytes := []byte(attr.Value.AsString())
			var evaluationResponse []*EvaluationResponse

			if err := json.Unmarshal(evaluationResponseBytes, &evaluationResponse); err != nil {
				return nil, err
			}

			return evaluationResponse, nil
		}
	}

	return nil, errNotTransformable
}

// ExportSpans transforms the spans into []*EvaluationResponse which the mutator takes to store into an analytics store.
func (a *AnalyticsSinkSpanExporter) ExportSpans(ctx context.Context, spans []sdktrace.ReadOnlySpan) error {
	evaluationResponses := make([]*EvaluationResponse, 0)

	for _, span := range spans {
		for _, event := range span.Events() {
			evaluationResponsesFromSpan, err := transformSpanEventToEvaluationResponses(event)
			if err != nil && !errors.Is(err, errNotTransformable) {
				a.logger.Error("event not decodable into evaluation response", zap.Error(err))
				continue
			}

			evaluationResponses = append(evaluationResponses, evaluationResponsesFromSpan...)
		}
	}

	return a.analyticsStoreMutator.IncrementFlagEvaluationCounts(ctx, evaluationResponses)
}

// Shutdown closes resources for an AnalyticsStoreMutator.
func (a *AnalyticsSinkSpanExporter) Shutdown(_ context.Context) error {
	return a.analyticsStoreMutator.Close()
}
