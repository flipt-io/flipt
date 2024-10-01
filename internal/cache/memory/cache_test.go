package memory

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
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
	require.NoError(t, err)

	v, ok, err := c.Get(ctx, "key")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, []byte("value"), v)

	v, ok, err = c.Get(ctx, "foo")
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Nil(t, v)

	v, ok, err = c.Get(ctx, "key")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, []byte("value"), v)
}

func TestDelete(t *testing.T) {
	var (
		c   = NewCache(config.CacheConfig{})
		ctx = context.Background()
	)

	err := c.Set(ctx, "key", []byte("value"))
	require.NoError(t, err)

	v, ok, err := c.Get(ctx, "key")
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, []byte("value"), v)

	err = c.Delete(ctx, "key")
	require.NoError(t, err)

	v, ok, err = c.Get(ctx, "key")
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Nil(t, v)
}
