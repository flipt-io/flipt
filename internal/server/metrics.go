package server

import (
	"go.flipt.io/flipt/internal/metrics"
	"go.opentelemetry.io/otel/metric/instrument"
)

// Prometheus metrics used throughout the server package
var (
	errorsTotal = metrics.MustSyncInt64().
		Counter(
			"flipt_server_errors_total",
			instrument.WithDescription("The total number of server errors"),
		)
)
