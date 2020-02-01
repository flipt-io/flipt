package cache

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Prometheus metrics used throughout the cache package
var (
	cacheItemCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "flipt",
		Subsystem: "storage",
		Name:      "cache_item_count",
		Help:      "The number of items currently in the cache",
	}, []string{"cache"})

	cacheHitTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flipt",
		Subsystem: "storage",
		Name:      "cache_hit_total",
		Help:      "The number of cache hits",
	}, []string{"type", "cache"})

	cacheMissTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flipt",
		Subsystem: "storage",
		Name:      "cache_miss_total",
		Help:      "The number of cache misses",
	}, []string{"type", "cache"})

	cacheFlushTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flipt",
		Subsystem: "storage",
		Name:      "cache_flush_total",
		Help:      "The number of times the cache is flushed",
	}, []string{"cache"})

	cacheEvictionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "flipt",
		Subsystem: "storage",
		Name:      "cache_eviction_total",
		Help:      "The number of times an item is evicted from the cache",
	}, []string{"cache"})
)
