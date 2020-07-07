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
	cacheItemCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "item_count",
		Help:      "The number of items currently in the cache",
	}, []string{"type", "cache"})

	cacheHitTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "hit_total",
		Help:      "The number of cache hits",
	}, []string{"type", "cache"})

	cacheMissTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "miss_total",
		Help:      "The number of cache misses",
	}, []string{"type", "cache"})

	cacheFlushTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "flush_total",
		Help:      "The number of times the cache is flushed",
	}, []string{"type", "cache"})

	cacheEvictionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "eviction_total",
		Help:      "The number of times an item is evicted from the cache",
	}, []string{"type", "cache"})
)
