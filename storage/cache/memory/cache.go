package memory

import (
	"context"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
)

// InMemoryCache wraps gocache.Cache in order to implement Cacher
type InMemoryCache struct {
	c *gocache.Cache
}

const (
	namespace = "flipt"
	subsystem = "cache"
)

var (
	cacheItemCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "item_count",
		Help:      "The number of items currently in the cache",
	}, []string{"cache"})

	cacheEvictionTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Subsystem: subsystem,
		Name:      "eviction_total",
		Help:      "The number of times an item is evicted from the cache",
	}, []string{"cache"})
)

// NewCache creates a new InMemoryCache with the provided expiration and evictionInterval
func NewCache(expiration time.Duration, evictionInterval time.Duration, logger logrus.FieldLogger) *InMemoryCache {
	logger = logger.WithField("cache", "memory")

	c := gocache.New(expiration, evictionInterval)
	c.OnEvicted(func(s string, _ interface{}) {
		cacheEvictionTotal.WithLabelValues("memory").Inc()
		cacheItemCount.WithLabelValues("memory").Dec()
		logger.Debugf("evicted key: %q", s)
	})

	return &InMemoryCache{c: c}
}

func (i *InMemoryCache) Get(_ context.Context, key string) (interface{}, bool, error) {
	v, ok := i.c.Get(key)
	if !ok {
		return nil, false, nil
	}
	return v, true, nil
}

func (i *InMemoryCache) Set(_ context.Context, key string, value interface{}) error {
	i.c.SetDefault(key, value)
	cacheItemCount.WithLabelValues("memory").Inc()
	return nil
}

func (i *InMemoryCache) Delete(_ context.Context, key string) error {
	i.c.Delete(key)
	cacheItemCount.WithLabelValues("memory").Dec()
	return nil
}

func (i *InMemoryCache) Flush(_ context.Context) error {
	i.c.Flush()
	cacheItemCount.WithLabelValues("memory").Set(0)
	return nil
}

func (i *InMemoryCache) String() string {
	return "memory"
}
