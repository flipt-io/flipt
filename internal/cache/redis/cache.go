package redis

import (
	"context"
	"errors"

	redis "github.com/go-redis/cache/v9"
	"go.flipt.io/flipt/internal/cache"
	"go.flipt.io/flipt/internal/config"
)

const cacheType = "redis"

type Cache struct {
	c   *redis.Cache
	cfg config.CacheConfig
}

// NewCache creates a new redis cache with the provided cache config
func NewCache(cfg config.CacheConfig, r *redis.Cache) *Cache {
	return &Cache{cfg: cfg, c: r}
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	var value []byte
	key = cache.Key(key)
	if err := c.c.Get(ctx, key, &value); err != nil {
		if errors.Is(err, redis.ErrCacheMiss) {
			cache.Observe(ctx, cacheType, cache.Miss)
			return nil, false, nil
		}

		cache.Observe(ctx, cacheType, cache.Error)
		return nil, false, err
	}

	cache.Observe(ctx, cacheType, cache.Hit)
	return value, true, nil
}

func (c *Cache) Set(ctx context.Context, key string, value []byte) error {
	key = cache.Key(key)
	if err := c.c.Set(&redis.Item{
		Ctx:   ctx,
		Key:   key,
		Value: value,
		TTL:   c.cfg.TTL,
	}); err != nil {
		cache.Observe(ctx, cacheType, cache.Error)
		return err
	}

	return nil
}

func (c *Cache) Delete(ctx context.Context, key string) error {
	key = cache.Key(key)
	if err := c.c.Delete(ctx, key); err != nil {
		cache.Observe(ctx, cacheType, cache.Error)
		return err
	}

	return nil
}

func (c *Cache) String() string {
	return cacheType
}
