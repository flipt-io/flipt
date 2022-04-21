package memory

import (
	"time"

	"github.com/markphelps/flipt/storage/cache/metrics"
	gocache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// InMemoryCache wraps gocache.Cache in order to implement Cacher
type InMemoryCache struct {
	c *gocache.Cache
}

// NewCache creates a new InMemoryCache with the provided expiration and evictionInterval
func NewCache(expiration time.Duration, evictionInterval time.Duration, logger logrus.FieldLogger) *InMemoryCache {
	logger = logger.WithField("cache", "memory")

	c := gocache.New(expiration, evictionInterval)
	c.OnEvicted(func(s string, _ interface{}) {
		metrics.CacheEvictionTotal.WithLabelValues("memory").Inc()
		metrics.CacheItemCount.WithLabelValues("memory").Dec()
		logger.Debugf("evicted key: %q", s)
	})

	return &InMemoryCache{c: c}
}

func (i *InMemoryCache) Get(key string) (interface{}, bool) {
	return i.c.Get(key)
}

func (i *InMemoryCache) Set(key string, value interface{}) {
	i.c.SetDefault(key, value)
	metrics.CacheItemCount.WithLabelValues("memory").Inc()
}

func (i *InMemoryCache) Delete(key string) {
	i.c.Delete(key)
	metrics.CacheItemCount.WithLabelValues("memory").Dec()
}

func (i *InMemoryCache) Flush() {
	i.c.Flush()
	metrics.CacheFlushTotal.WithLabelValues("memory").Inc()
	metrics.CacheItemCount.WithLabelValues("memory").Set(0)
}

func (i *InMemoryCache) String() string {
	return "memory"
}
