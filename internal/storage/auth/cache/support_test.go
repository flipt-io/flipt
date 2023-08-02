package cache

import (
	"context"

	"go.flipt.io/flipt/internal/cache"
)

var _ cache.Cacher = &cacheSpy{}

type cacheSpy struct {
	cached      bool
	cachedValue []byte
	cacheKey    string
	getErr      error
	setErr      error
}

func (c *cacheSpy) String() string {
	return "mockCacher"
}

func (c *cacheSpy) Get(ctx context.Context, key string) ([]byte, bool, error) {
	c.cacheKey = key

	if c.getErr != nil || !c.cached {
		return nil, c.cached, c.getErr
	}

	return c.cachedValue, true, nil
}

func (c *cacheSpy) Set(ctx context.Context, key string, value []byte) error {
	c.cacheKey = key
	c.cachedValue = value

	if c.setErr != nil {
		return c.setErr
	}

	return nil
}

func (c *cacheSpy) Delete(ctx context.Context, key string) error {
	return nil
}
