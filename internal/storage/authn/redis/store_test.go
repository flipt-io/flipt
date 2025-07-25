package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.flipt.io/flipt/internal/storage/authn"
	authtesting "go.flipt.io/flipt/internal/storage/authn/testing"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
)

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

func TestAuthenticationStoreRedis(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	var (
		redisAddr   = os.Getenv("REDIS_HOST")
		redisCancel = func(context.Context, ...testcontainers.TerminateOption) error { return nil }
	)

	if redisAddr == "" {
		t.Log("Starting redis container.")

		redisContainer, err := setupRedis(context.Background())
		require.NoError(t, err, "Failed to start redis container.")

		redisCancel = redisContainer.Terminate
		redisAddr = fmt.Sprintf("%s:%s", redisContainer.host, redisContainer.port)
	}

	rdb := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})

	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		t.Log("Shutting down redis container.")

		if redisCancel != nil {
			err := redisCancel(shutdownCtx)
			if err != nil {
				t.Logf("Error terminating container: %v", err)
			}
		}
	})

	authtesting.TestAuthenticationStoreHarness(t, func(t *testing.T) authn.Store {
		return NewStore(rdb, zap.NewNop())
	})
}

func TestPrefixKeys(t *testing.T) {
	t.Run("authIDKey", func(t *testing.T) {
		require.Equal(t, "flipt:auth:id:123", authIDKey("flipt", "123"))
		require.Equal(t, "auth:id:123", authIDKey("", "123"))
	})
	t.Run("authTokenKey", func(t *testing.T) {
		require.Equal(t, "flipt:auth:token:123", authTokenKey("flipt", "123"))
		require.Equal(t, "auth:token:123", authTokenKey("", "123"))
	})
	t.Run("authMethodKey", func(t *testing.T) {
		require.Equal(t, "flipt:auth:method:METHOD_TOKEN", authMethodKey("flipt", auth.Method_METHOD_TOKEN))
		require.Equal(t, "auth:method:METHOD_TOKEN", authMethodKey("", auth.Method_METHOD_TOKEN))
	})
	t.Run("authAllKey", func(t *testing.T) {
		require.Equal(t, "flipt:auth:all", authAllKey("flipt"))
		require.Equal(t, "auth:all", authAllKey(""))
	})
}
