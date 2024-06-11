package cache

import (
	"context"

	"go.flipt.io/flipt/internal/cache"
)

var (
	_ cache.Cacher = &cacheMock{}
	_ cache.Cacher = &cacheSpy{}
)

type cacheMock struct {
	cached      bool
	cachedValue []byte
	cacheKey    string
	getErr      error
	setErr      error
}

func (c *cacheMock) String() string {
	return "mockCacher"
}

func (c *cacheMock) Get(ctx context.Context, key string) ([]byte, bool, error) {
	c.cacheKey = key

	if c.getErr != nil || !c.cached {
		return nil, c.cached, c.getErr
	}

	return c.cachedValue, true, nil
}

func (c *cacheMock) Set(ctx context.Context, key string, value []byte) error {
	c.cacheKey = key
	c.cachedValue = value

	if c.setErr != nil {
		return c.setErr
	}

	return nil
}

func (c *cacheMock) Delete(ctx context.Context, key string) error {
	return nil
}

type cacheSpy struct {
	cache.Cacher

	getKeys   map[string]struct{}
	getCalled int

	setItems  map[string][]byte
	setCalled int

	deleteKeys   map[string]struct{}
	deleteCalled int
}

func newCacheSpy(c cache.Cacher) *cacheSpy {
	return &cacheSpy{
		Cacher:     c,
		getKeys:    make(map[string]struct{}),
		setItems:   make(map[string][]byte),
		deleteKeys: make(map[string]struct{}),
	}
}

func (c *cacheSpy) Get(ctx context.Context, key string) ([]byte, bool, error) {
	c.getCalled++
	c.getKeys[key] = struct{}{}
	return c.Cacher.Get(ctx, key)
}

func (c *cacheSpy) Set(ctx context.Context, key string, value []byte) error {
	c.setCalled++
	c.setItems[key] = value
	return c.Cacher.Set(ctx, key, value)
}

func (c *cacheSpy) Delete(ctx context.Context, key string) error {
	c.deleteCalled++
	c.deleteKeys[key] = struct{}{}
	return c.Cacher.Delete(ctx, key)
}
