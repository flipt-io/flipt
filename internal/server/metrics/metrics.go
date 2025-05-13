package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.flipt.io/flipt/internal/otel/metrics"
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
	AttributeMatch       = attribute.Key("flipt_match")
	AttributeFlag        = attribute.Key("flipt_flag")
	AttributeFlagType    = attribute.Key("flipt_flag_type")
	AttributeSegments    = attribute.Key("flipt_segments")
	AttributeReason      = attribute.Key("flipt_reason")
	AttributeValue       = attribute.Key("flipt_value")
	AttributeNamespace   = attribute.Key("flipt_namespace")
	AttributeEnvironment = attribute.Key("flipt_environment")
)

// Snapshot (Unary) metrics
var (
	EvaluationsSnapshotRequestsTotal = metrics.MustInt64().Counter(
		prometheus.BuildFQName(namespace, evaluationsSubsystem, "snapshot_requests"),
		metric.WithDescription("Total number of snapshot (unary) evaluation requests"),
	)
	EvaluationsSnapshotErrorsTotal = metrics.MustInt64().Counter(
		prometheus.BuildFQName(namespace, evaluationsSubsystem, "snapshot_errors"),
		metric.WithDescription("Total number of errors in snapshot (unary) evaluation requests"),
	)
	EvaluationsSnapshotLatency = metrics.MustFloat64().Histogram(
		prometheus.BuildFQName(namespace, evaluationsSubsystem, "snapshot_latency"),
		metric.WithDescription("Latency of snapshot (unary) evaluation requests in milliseconds"),
		metric.WithUnit("ms"),
	)
)

// Stream metrics
var (
	EvaluationsStreamRequestsTotal = metrics.MustInt64().Counter(
		prometheus.BuildFQName(namespace, evaluationsSubsystem, "stream_requests"),
		metric.WithDescription("Total number of stream (subscribe) evaluation requests"),
	)
	EvaluationsStreamErrorsTotal = metrics.MustInt64().Counter(
		prometheus.BuildFQName(namespace, evaluationsSubsystem, "stream_errors"),
		metric.WithDescription("Total number of errors in stream (subscribe) evaluation requests"),
	)
	EvaluationsStreamMessagesTotal = metrics.MustInt64().Counter(
		prometheus.BuildFQName(namespace, evaluationsSubsystem, "stream_messages"),
		metric.WithDescription("Total number of messages sent over evaluation streams"),
	)
	EvaluationsStreamLatency = metrics.MustFloat64().Histogram(
		prometheus.BuildFQName(namespace, evaluationsSubsystem, "stream_latency"),
		metric.WithDescription("Latency of stream (subscribe) evaluation setup in milliseconds"),
		metric.WithUnit("ms"),
	)
)
