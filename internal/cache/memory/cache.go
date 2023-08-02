package memory

import (
	"context"

	gocache "github.com/patrickmn/go-cache"
	"go.flipt.io/flipt/internal/cache"
	"go.flipt.io/flipt/internal/config"
)

const cacheType = "memory"

// Cache wraps gocache.Cache in order to implement Cacher
type Cache struct {
	c *gocache.Cache
}

// NewCache creates a new in memory cache with the provided cache config
func NewCache(cfg config.CacheConfig) *Cache {
	return &Cache{c: gocache.New(cfg.TTL, cfg.Memory.EvictionInterval)}
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	key = cache.Key(key)
	v, ok := c.c.Get(key)
	if !ok {
		cache.Observe(ctx, cacheType, cache.Miss)
		return nil, false, nil
	}

	cache.Observe(ctx, cacheType, cache.Hit)
	return v.([]byte), true, nil
}

func (c *Cache) Set(_ context.Context, key string, value []byte) error {
	key = cache.Key(key)
	c.c.SetDefault(key, value)
	return nil
}

func (c *Cache) Delete(_ context.Context, key string) error {
	key = cache.Key(key)
	c.c.Delete(key)
	return nil
}

func (c *Cache) String() string {
	return cacheType
}
