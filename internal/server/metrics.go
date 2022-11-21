package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.flipt.io/flipt/internal/metrics"
	"go.opentelemetry.io/otel/metric/instrument"
)

const (
	namespace = "flipt"
	subsystem = "server"
)

// Prometheus metrics used throughout the server package
var (
	errorsTotal = metrics.MustSyncInt64().
		Counter(
			prometheus.BuildFQName(namespace, subsystem, "errors_total"),
			instrument.WithDescription("The total number of server errors"),
		)
)
