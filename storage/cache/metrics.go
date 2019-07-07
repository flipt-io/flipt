package cache

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics used throughout the cache package
var (
	cacheHitTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flipt",
		Subsystem: "storage",
		Name:      "cache_hit_total",
		Help:      "The total number of cache hits",
	}, []string{"type", "cache"})

	cacheMissTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flipt",
		Subsystem: "storage",
		Name:      "cache_miss_total",
		Help:      "The total number of cache misses",
	}, []string{"type", "cache"})
)
