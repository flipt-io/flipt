package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.flipt.io/flipt/internal/metrics"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
)

const (
	namespace            = "flipt"
	serverSubsystem      = "server"
	evaluationsSubsystem = "evaluations"
)

// Prometheus metrics used throughout the server package
var (
	// ErrorsTotal is the total number of errors in the server
	ErrorsTotal = metrics.MustInt64().
		Counter(
			prometheus.BuildFQName(namespace, serverSubsystem, "errors"),
			metric.WithDescription("The total number of server errors"),
		)
)

// Evaluation specific metrics
var (
	// EvaluationsTotal is the total number of evaluation requests
	EvaluationsTotal = metrics.MustInt64().
				Counter(
			prometheus.BuildFQName(namespace, evaluationsSubsystem, "requests"),
			metric.WithDescription("The total number of requested evaluations"),
		)

	// EvaluationErrorsTotal is the total number of evaluation errors
	EvaluationErrorsTotal = metrics.MustInt64().
				Counter(
			prometheus.BuildFQName(namespace, evaluationsSubsystem, "errors"),
			metric.WithDescription("The total number of requested evaluations"),
		)

	// EvaluationResultsTotal is the total number of evaluation results
	EvaluationResultsTotal = metrics.MustInt64().
				Counter(
			prometheus.BuildFQName(namespace, evaluationsSubsystem, "results"),
			metric.WithDescription("Count of results including match, flag, segment, reason and value attributes"),
		)

	// EvaluationLatency is the latency of individual evaluations
	EvaluationLatency = metrics.MustFloat64().Histogram(
		prometheus.BuildFQName(namespace, evaluationsSubsystem, "latency"),
		metric.WithDescription("The latency of inidividual evaluations in milliseconds"),
		metric.WithUnit("ms"),
	)

	// Attributes used in evaluation metrics
	//nolint
	AttributeMatch     = attribute.Key("match")
	AttributeFlag      = attribute.Key("flag")
	AttributeSegment   = attribute.Key("segment")
	AttributeReason    = attribute.Key("reason")
	AttributeValue     = attribute.Key("value")
	AttributeNamespace = attribute.Key("namespace")
	AttributeType      = attribute.Key("type")
)
