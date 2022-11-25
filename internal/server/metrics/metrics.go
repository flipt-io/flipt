package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.flipt.io/flipt/internal/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric/instrument"
	"go.opentelemetry.io/otel/metric/instrument/syncfloat64"
	"go.opentelemetry.io/otel/metric/unit"
)

const (
	namespace            = "flipt"
	serverSubsystem      = "server"
	evaluationsSubsystem = "evaluations"
)

// Prometheus metrics used throughout the server package
var (
	// ErrorsTotal is the total number of errors in the server
	ErrorsTotal = metrics.MustSyncInt64().
		Counter(
			prometheus.BuildFQName(namespace, serverSubsystem, "errors"),
			instrument.WithDescription("The total number of server errors"),
		)
)

// Evaluation specific metrics
var (
	// EvaluationsTotal is the total number of evaluation requests
	EvaluationsTotal = metrics.MustSyncInt64().
				Counter(
			prometheus.BuildFQName(namespace, evaluationsSubsystem, "requests"),
			instrument.WithDescription("The total number of requested evaluations"),
		)

	// EvaluationErrorsTotal is the total number of evaluation errors
	EvaluationErrorsTotal = metrics.MustSyncInt64().
				Counter(
			prometheus.BuildFQName(namespace, evaluationsSubsystem, "errors"),
			instrument.WithDescription("The total number of requested evaluations"),
		)

	// EvaluationResultsTotal is the total number of evaluation results
	EvaluationResultsTotal = metrics.MustSyncInt64().
				Counter(
			prometheus.BuildFQName(namespace, evaluationsSubsystem, "results"),
			instrument.WithDescription("Count of results including match, flag, segment, reason and value attributes"),
		)

	// EvaluationLatency is the latency of individual evaluations
	EvaluationLatency syncfloat64.Histogram

	// Attributes used in evaluation metrics
	//nolint
	AttributeMatch   = attribute.Key("match")
	AttributeFlag    = attribute.Key("flag")
	AttributeSegment = attribute.Key("segment")
	AttributeReason  = attribute.Key("reason")
	AttributeValue   = attribute.Key("value")
)

func init() {
	var err error
	EvaluationLatency, err = metrics.Meter.SyncFloat64().Histogram(
		prometheus.BuildFQName(namespace, evaluationsSubsystem, "latency"),
		instrument.WithDescription("The latency of inidividual evaluations in milliseconds"),
		instrument.WithUnit(unit.Milliseconds),
	)

	if err != nil {
		panic(err)
	}
}
