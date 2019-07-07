package server

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics used throughout the server package
var (
	errorsTotal = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "flipt",
		Subsystem: "server",
		Name:      "errors_total",
		Help:      "The total number of server errors",
	})
)
