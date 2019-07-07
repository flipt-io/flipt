package cache

import (
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ErrCacheCorrupt represents a corrupt cache error
var ErrCacheCorrupt = errors.New("cache corrupted")

// Prometheus variables used throughout the cache package
var (
	CacheHitTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_hit_total",
		Help: "The total number of cache hits",
	}, []string{"type", "cache"})

	CacheMissTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "cache_miss_total",
		Help: "The total number of cache misses",
	}, []string{"type", "cache"})
)

// Cacher modifies and queries a cache
type Cacher interface {
	Get(key interface{}) (interface{}, bool)
	Add(key, value interface{}) bool
	Remove(key interface{})
}
