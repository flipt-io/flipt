package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	goredis_cache "github.com/go-redis/cache/v9"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.flipt.io/flipt/internal/config"
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
		ctx         = context.Background()
		c, teardown = newCache(t, ctx)
	)

	defer teardown()

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

type redisContainer struct {
	testcontainers.Container
	host string
	port string
}

func setupRedis(ctx context.Context) (*redisContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:alpine",
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
		t.Skip("skipping test in short mode")
	}

	var (
		redisAddr   = os.Getenv("REDIS_HOST")
		redisCancel = func(context.Context) error { return nil }
	)

	if redisAddr == "" {
		t.Log("Starting redis container.")

		redisContainer, err := setupRedis(ctx)
		require.NoError(t, err, "Failed to start redis container.")

		redisCancel = redisContainer.Terminate
		redisAddr = fmt.Sprintf("%s:%s", redisContainer.host, redisContainer.port)
	}

	rdb := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})

	cache := NewCache(config.CacheConfig{
		TTL: 30 * time.Second,
	}, goredis_cache.New(&goredis_cache.Options{
		Redis: rdb,
	}))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	teardown := func() {
		_ = redisCancel(shutdownCtx)
		cancel()
	}

	return cache, teardown
}
