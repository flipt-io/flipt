package redis

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage/authn"
	authtesting "go.flipt.io/flipt/internal/storage/authn/testing"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
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

		redisContainer, err := setupRedis(t.Context())
		require.NoError(t, err, "Failed to start redis container.")

		redisCancel = redisContainer.Terminate
		redisAddr = fmt.Sprintf("%s:%s", redisContainer.host, redisContainer.port)
	}

	rdb := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})

	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
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

func TestSIDAndIDTokenStorage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode")
	}

	var (
		redisAddr   = os.Getenv("REDIS_HOST")
		redisCancel = func(context.Context, ...testcontainers.TerminateOption) error { return nil }
	)

	if redisAddr == "" {
		t.Log("Starting redis container.")

		redisContainer, err := setupRedis(t.Context())
		require.NoError(t, err, "Failed to start redis container.")

		redisCancel = redisContainer.Terminate
		redisAddr = fmt.Sprintf("%s:%s", redisContainer.host, redisContainer.port)
	}

	rdb := goredis.NewClient(&goredis.Options{
		Addr: redisAddr,
	})

	t.Cleanup(func() {
		shutdownCtx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()

		if redisCancel != nil {
			err := redisCancel(shutdownCtx)
			if err != nil {
				t.Logf("Error terminating container: %v", err)
			}
		}
	})

	store := NewStore(rdb, zap.NewNop())
	ctx := t.Context()

	t.Run("Create authentication with SID and IDToken", func(t *testing.T) {
		token, created, err := store.CreateAuthentication(ctx, &authn.CreateAuthenticationRequest{
			Method:    auth.Method_METHOD_OIDC,
			ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
			SessionID: "google:session-123",
			IDToken:   "my-id-token-jwt",
		})
		require.NoError(t, err)
		require.NotEmpty(t, token)
		require.NotNil(t, created)

		t.Run("GetAuthenticationIDBySID returns the auth ID", func(t *testing.T) {
			id, err := store.GetAuthenticationIDBySID(ctx, "google:session-123")
			require.NoError(t, err)
			assert.Equal(t, created.Id, id)
		})

		t.Run("GetAuthenticationIDBySID returns not found for unknown SID", func(t *testing.T) {
			_, err := store.GetAuthenticationIDBySID(ctx, "unknown-sid")
			var notFound errors.ErrNotFound
			require.ErrorAs(t, err, &notFound)
		})

		t.Run("GetIDToken returns the stored ID token", func(t *testing.T) {
			data, err := store.GetIDToken(ctx, created.Id)
			require.NoError(t, err)
			require.NotNil(t, data)
			assert.Equal(t, "my-id-token-jwt", data.IDToken)
		})

		t.Run("GetIDToken returns not found for auth without ID token", func(t *testing.T) {
			_, otherAuth, err := store.CreateAuthentication(ctx, &authn.CreateAuthenticationRequest{
				Method:    auth.Method_METHOD_TOKEN,
				ExpiresAt: timestamppb.New(time.Now().Add(time.Hour)),
			})
			require.NoError(t, err)

			_, err = store.GetIDToken(ctx, otherAuth.Id)
			var notFound errors.ErrNotFound
			require.ErrorAs(t, err, &notFound)
		})

		t.Run("GetIDToken returns not found for unknown auth ID", func(t *testing.T) {
			_, err := store.GetIDToken(ctx, "non-existent")
			var notFound errors.ErrNotFound
			require.ErrorAs(t, err, &notFound)
		})

		t.Run("Delete authentication cleans up SID and IDToken", func(t *testing.T) {
			err := store.DeleteAuthentications(ctx, authn.Delete(authn.WithID(created.Id)))
			require.NoError(t, err)

			_, err = store.GetAuthenticationIDBySID(ctx, "google:session-123")
			var notFound errors.ErrNotFound
			require.ErrorAs(t, err, &notFound)

			_, err = store.GetIDToken(ctx, created.Id)
			require.ErrorAs(t, err, &notFound)
		})
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
	t.Run("oauthChallengeKey", func(t *testing.T) {
		require.Equal(t, "flipt:auth:oauth_challenge:123", oauthChallengeKey("flipt", "123"))
		require.Equal(t, "auth:oauth_challenge:123", oauthChallengeKey("", "123"))
	})
	t.Run("authSessionKey", func(t *testing.T) {
		require.Equal(t, "flipt:auth:session:google:sid-123", authSessionKey("flipt", "google:sid-123"))
		require.Equal(t, "auth:session:google:sid-123", authSessionKey("", "google:sid-123"))
	})
}
