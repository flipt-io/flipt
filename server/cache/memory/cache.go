package memory

import (
	"context"
	"sync/atomic"

	gocache "github.com/patrickmn/go-cache"
	"go.flipt.io/flipt/config"
	"go.flipt.io/flipt/server/cache"
)

// Cache wraps gocache.Cache in order to implement Cacher
type Cache struct {
	c         *gocache.Cache
	missTotal uint64
	hitTotal  uint64
}

// NewCache creates a new in memory cache with the provided cache config
func NewCache(cfg config.CacheConfig) *Cache {
	var (
		gc = gocache.New(cfg.TTL, cfg.Memory.EvictionInterval)
		c  = &Cache{c: gc}
	)

	cache.RegisterMetrics(c)
	return c
}

func (c *Cache) Get(_ context.Context, key string) ([]byte, bool, error) {
	v, ok := c.c.Get(key)
	if !ok {
		atomic.AddUint64(&c.missTotal, 1)
		return nil, false, nil
	}

	atomic.AddUint64(&c.hitTotal, 1)
	return v.([]byte), true, nil
}

func (c *Cache) Set(_ context.Context, key string, value []byte) error {
	c.c.SetDefault(key, value)
	return nil
}

func (c *Cache) Delete(_ context.Context, key string) error {
	c.c.Delete(key)
	return nil
}

func (c *Cache) String() string {
	return "memory"
}

func (c *Cache) Stats() cache.Stats {
	return cache.Stats{
		MissTotal:  c.missTotal,
		HitTotal:   c.hitTotal,
		ErrorTotal: 0,
	}
}
