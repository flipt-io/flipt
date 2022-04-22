package cache

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	namespace = "flipt"
	subsystem = "cache"
)

// Prometheus metrics used throughout the cache package
var (
	cacheHitTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "hit_total",
		Help:      "The number of cache hits",
	}, []string{"cache"})

	cacheMissTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "miss_total",
		Help:      "The number of cache misses",
	}, []string{"cache"})

	cacheFlushTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "flush_total",
		Help:      "The number of times the cache is flushed",
	}, []string{"cache"})
)
