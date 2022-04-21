package metrics

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
	CacheItemCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "item_count",
		Help:      "The number of items currently in the cache",
	}, []string{"cache"})

	CacheHitTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "hit_total",
		Help:      "The number of cache hits",
	}, []string{"cache"})

	CacheMissTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "miss_total",
		Help:      "The number of cache misses",
	}, []string{"cache"})

	CacheFlushTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "flush_total",
		Help:      "The number of times the cache is flushed",
	}, []string{"cache"})

	CacheEvictionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "eviction_total",
		Help:      "The number of times an item is evicted from the cache",
	}, []string{"cache"})
)
