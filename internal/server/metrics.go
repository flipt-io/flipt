package server

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
var errorsTotal = metrics.MustSyncInt64().
	Counter(
		prometheus.BuildFQName(namespace, serverSubsystem, "errors"),
		instrument.WithDescription("The total number of server errors"),
	)

// Evaluation specific metrics
var (
	evaluationsTotal = metrics.MustSyncInt64().
				Counter(
			prometheus.BuildFQName(namespace, evaluationsSubsystem, "requests"),
			instrument.WithDescription("The total number of requested evaluations"),
		)

	evaluationErrorsTotal = metrics.MustSyncInt64().
				Counter(
			prometheus.BuildFQName(namespace, evaluationsSubsystem, "errors"),
			instrument.WithDescription("The total number of requested evaluations"),
		)

	evaluationResultsTotal = metrics.MustSyncInt64().
				Counter(
			prometheus.BuildFQName(namespace, evaluationsSubsystem, "results"),
			instrument.WithDescription("Count of results including match, flag, segment, reason and value attributes"),
		)

	evaluationLatency syncfloat64.Histogram

	attributeMatch   = attribute.Key("match")
	attributeFlag    = attribute.Key("flag")
	attributeSegment = attribute.Key("segment")
	attributeReason  = attribute.Key("reason")
	attributeValue   = attribute.Key("value")
)

func init() {
	var err error
	evaluationLatency, err = metrics.Meter.SyncFloat64().Histogram(
		prometheus.BuildFQName(namespace, evaluationsSubsystem, "latency"),
		instrument.WithDescription("The latency of inidividual evaluations in milliseconds"),
		instrument.WithUnit(unit.Milliseconds),
	)

	if err != nil {
		panic(err)
	}
}
