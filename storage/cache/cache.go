package cache

import (
	"errors"
	"time"

	gocache "github.com/patrickmn/go-cache"
	"github.com/sirupsen/logrus"
)

// ErrCacheCorrupt represents a corrupt cache error
var ErrCacheCorrupt = errors.New("cache corrupted")

// Cacher modifies and queries a cache
type Cacher interface {
	Get(key string) (interface{}, bool)
	Set(key string, value interface{})
	Delete(key string)
	Flush()
}

// InMemoryCache wraps gocache.Cache in order to implement Cacher
type InMemoryCache struct {
	c *gocache.Cache
}

// NewInMemoryCache creates a new InMemoryCache with the provided expirationDuration
// and 10 minute cleanupDuration
func NewInMemoryCache(expDuration time.Duration, logger logrus.FieldLogger) *InMemoryCache {
	logger = logger.WithField("cache", "memory")

	c := gocache.New(expDuration, 10*time.Minute)
	c.OnEvicted(func(s string, _ interface{}) {
		logger.Debugf("evicted key: %q", s)
	})

	return &InMemoryCache{c: c}
}

func (i *InMemoryCache) Get(key string) (interface{}, bool) {
	return i.c.Get(key)
}

func (i *InMemoryCache) Set(key string, value interface{}) {
	i.c.SetDefault(key, value)
}

func (i *InMemoryCache) Delete(key string) {
	i.c.Delete(key)
}

func (i *InMemoryCache) Flush() {
	i.c.Flush()
}
