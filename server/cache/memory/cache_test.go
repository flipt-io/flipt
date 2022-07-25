package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.flipt.io/flipt/config"
)

func TestNewCache(t *testing.T) {
	c := NewCache(config.CacheConfig{})
	assert.NotNil(t, c)
	assert.Equal(t, "memory", c.String())
}

func TestSet(t *testing.T) {
	var (
		c   = NewCache(config.CacheConfig{})
		ctx = context.Background()
	)

	err := c.Set(ctx, "key", []byte("value"))
	assert.NoError(t, err)
}

func TestGet(t *testing.T) {
	var (
		c   = NewCache(config.CacheConfig{})
		ctx = context.Background()
	)

	err := c.Set(ctx, "key", []byte("value"))
	assert.NoError(t, err)

	v, ok, err := c.Get(ctx, "key")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, []byte("value"), v)
	assert.Equal(t, uint64(1), c.Stats().HitTotal)
	assert.Equal(t, uint64(0), c.Stats().MissTotal)
	assert.Equal(t, uint64(0), c.Stats().ErrorTotal)

	v, ok, err = c.Get(ctx, "foo")
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Nil(t, v)
	assert.Equal(t, uint64(1), c.Stats().HitTotal)
	assert.Equal(t, uint64(1), c.Stats().MissTotal)
	assert.Equal(t, uint64(0), c.Stats().ErrorTotal)

	v, ok, err = c.Get(ctx, "key")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, []byte("value"), v)
	assert.Equal(t, uint64(2), c.Stats().HitTotal)
	assert.Equal(t, uint64(1), c.Stats().MissTotal)
	assert.Equal(t, uint64(0), c.Stats().ErrorTotal)
}

func TestDelete(t *testing.T) {
	var (
		c   = NewCache(config.CacheConfig{})
		ctx = context.Background()
	)

	err := c.Set(ctx, "key", []byte("value"))
	assert.NoError(t, err)

	v, ok, err := c.Get(ctx, "key")
	assert.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, []byte("value"), v)
	assert.Equal(t, uint64(1), c.Stats().HitTotal)
	assert.Equal(t, uint64(0), c.Stats().MissTotal)
	assert.Equal(t, uint64(0), c.Stats().ErrorTotal)

	err = c.Delete(ctx, "key")
	assert.NoError(t, err)

	v, ok, err = c.Get(ctx, "key")
	assert.NoError(t, err)
	assert.False(t, ok)
	assert.Nil(t, v)
	assert.Equal(t, uint64(1), c.Stats().HitTotal)
	assert.Equal(t, uint64(1), c.Stats().MissTotal)
	assert.Equal(t, uint64(0), c.Stats().ErrorTotal)
}
