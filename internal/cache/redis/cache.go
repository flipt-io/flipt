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
	k   *cache.Keyer
}

// NewCache creates a new redis cache with the provided cache config
func NewCache(cfg config.CacheConfig, r *redis.Cache) *Cache {
	keyer := cache.DefaultKeyer
	if cfg.Redis.Prefix != "" {
		keyer = cache.NewKeyer(cfg.Redis.Prefix)
	}
	return &Cache{cfg: cfg, c: r, k: keyer}
}

func (c *Cache) Get(ctx context.Context, key string) ([]byte, bool, error) {
	key = c.k.Key(key)
	var value []byte
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
	key = c.k.Key(key)

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
	key = c.k.Key(key)

	if err := c.c.Delete(ctx, key); err != nil {
		cache.Observe(ctx, cacheType, cache.Error)
		return err
	}

	return nil
}

func (c *Cache) String() string {
	return cacheType
}
