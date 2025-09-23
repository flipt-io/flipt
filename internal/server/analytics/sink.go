package analytics

import (
	"context"
	"errors"
	"time"

	"go.flipt.io/flipt/internal/server/tracing"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/zap"
	"k8s.io/utils/ptr"
)

var errNotTransformable = errors.New("event not transformable into evaluation response")

type AnalyticsStoreMutator interface {
	IncrementFlagEvaluationCounts(ctx context.Context, responses []*EvaluationResponse) error
	Close() error
}

type EvaluationResponse struct {
	FlagKey         string    `json:"flagKey,omitempty"`
	FlagType        string    `json:"flagType,omitempty"`
	EnvironmentKey  string    `json:"environmentKey,omitempty"`
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

// transformSpanEventToEvaluationResponses is a convenience function to transform a span event into an []*EvaluationResponse.
func transformSpanEventToEvaluationResponses(event sdktrace.Event) ([]*EvaluationResponse, error) {
	if event.Name != tracing.Event {
		return nil, errNotTransformable
	}

	r := &EvaluationResponse{
		Timestamp: event.Time.UTC(),
	}
	for _, v := range event.Attributes {
		switch v.Key {
		case tracing.AttributeEnvironment:
			r.EnvironmentKey = v.Value.AsString()
		case tracing.AttributeNamespace:
			r.NamespaceKey = v.Value.AsString()
		case tracing.AttributeFlag:
			r.FlagKey = v.Value.AsString()
		case tracing.AttributeFlagType:
			r.FlagType = v.Value.AsString()
		case tracing.AttributeEntityID:
			r.EntityId = v.Value.AsString()
		case tracing.AttributeMatch:
			if v.Value.Type() == attribute.BOOL {
				r.Match = ptr.To(v.Value.AsBool())
			}
		case tracing.AttributeReason:
			r.Reason = v.Value.AsString()
		case tracing.AttributeValue:
			r.EvaluationValue = ptr.To(v.Value.AsString())
		}
	}
	return []*EvaluationResponse{r}, nil
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
	if a != nil && a.analyticsStoreMutator != nil {
		return a.analyticsStoreMutator.Close()
	}

	return nil
}
