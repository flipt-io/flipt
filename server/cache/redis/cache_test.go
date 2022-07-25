package redis

import (
	"context"
	"fmt"
	"testing"
	"time"

	goredis_cache "github.com/go-redis/cache/v8"
	goredis "github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.flipt.io/flipt/config"
)

func TestSet(t *testing.T) {
	var (
		ctx         = context.Background()
		c, teardown = newCache(t, ctx)
	)

	defer teardown()

	err := c.Set(ctx, "key", []byte("value"))
	assert.NoError(t, err)
}

func TestGet(t *testing.T) {
	var (
		ctx         = context.Background()
		c, teardown = newCache(t, ctx)
	)

	defer teardown()

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
		ctx         = context.Background()
		c, teardown = newCache(t, ctx)
	)

	defer teardown()

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

type redisContainer struct {
	testcontainers.Container
	host string
	port string
}

func setupRedis(ctx context.Context) (*redisContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:latest",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("* Ready to accept connections"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, "6379")
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	return &redisContainer{Container: container, host: hostIP, port: mappedPort.Port()}, nil
}

func newCache(t *testing.T, ctx context.Context) (*Cache, func()) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping redis test in short mode")
	}

	redisContainer, err := setupRedis(ctx)
	if err != nil {
		assert.FailNowf(t, "failed to setup redis container: %s", err.Error())
	}

	rdb := goredis.NewClient(&goredis.Options{
		Addr: fmt.Sprintf("%s:%s", redisContainer.host, redisContainer.port),
	})

	cache := NewCache(config.CacheConfig{
		TTL: 30 * time.Second,
	}, goredis_cache.New(&goredis_cache.Options{
		Redis: rdb,
	}))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	teardown := func() {
		_ = rdb.Shutdown(shutdownCtx)
		_ = redisContainer.Terminate(shutdownCtx)
		cancel()
	}

	return cache, teardown
}
