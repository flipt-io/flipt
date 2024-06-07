package cleanup

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	authstorage "go.flipt.io/flipt/internal/storage/authn"
	inmemauth "go.flipt.io/flipt/internal/storage/authn/memory"
	inmemoplock "go.flipt.io/flipt/internal/storage/oplock/memory"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap/zaptest"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCleanup(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	var (
		ctx        = context.Background()
		logger     = zaptest.NewLogger(t)
		authstore  = inmemauth.NewStore()
		lock       = inmemoplock.New()
		authConfig = config.AuthenticationConfig{
			Methods: config.AuthenticationMethods{},
		}
	)

	// enable all methods and set their cleanup configuration
	for _, info := range authConfig.Methods.AllMethods(ctx) {
		info.Enable(t)
		info.SetCleanup(t, config.AuthenticationCleanupSchedule{
			Interval:    time.Second,
			GracePeriod: 5 * time.Second,
		})
	}

	// create an initial non-expiring token
	clientToken, storedAuth, err := authstore.CreateAuthentication(
		ctx,
		&authstorage.CreateAuthenticationRequest{Method: auth.Method_METHOD_TOKEN},
	)
	require.NoError(t, err)

	for i := 0; i < 5; i++ {
		// run five instances of service
		// it should be a safe operation given they share the same lock service
		service := NewAuthenticationService(logger, lock, authstore, authConfig)
		service.Run(ctx)
		t.Cleanup(func() {
			require.NoError(t, service.Shutdown(context.TODO()))
		})
	}

	t.Run("ensure non-expiring token exists", func(t *testing.T) {
		retrievedAuth, err := authstore.GetAuthenticationByClientToken(ctx, clientToken)
		require.NoError(t, err)
		assert.Equal(t, storedAuth, retrievedAuth)
	})

	for _, info := range authConfig.Methods.AllMethods(ctx) {
		info := info
		if !info.RequiresDatabase {
			continue
		}

		t.Run(fmt.Sprintf("Authentication Method %q", info.Method), func(t *testing.T) {
			t.Parallel()

			t.Log("create an expiring token and ensure it exists")
			clientToken, storedAuth, err := authstore.CreateAuthentication(
				ctx,
				&authstorage.CreateAuthenticationRequest{
					Method:    info.Method,
					ExpiresAt: timestamppb.New(time.Now().UTC().Add(5 * time.Second)),
				},
			)
			require.NoError(t, err)

			retrievedAuth, err := authstore.GetAuthenticationByClientToken(ctx, clientToken)
			require.NoError(t, err)
			assert.Equal(t, storedAuth, retrievedAuth)

			t.Log("ensure grace period protects token from being deleted")
			// token should still exist as it wont be deleted until
			// expiry + grace period (5s + 5s == 10s)
			time.Sleep(5 * time.Second)

			retrievedAuth, err = authstore.GetAuthenticationByClientToken(ctx, clientToken)
			require.NoError(t, err)
			assert.Equal(t, storedAuth, retrievedAuth)

			// ensure authentication is expired but still persisted
			assert.True(t, retrievedAuth.ExpiresAt.AsTime().Before(time.Now().UTC()))

			t.Log("once expiry and grace period ellapses ensure token is deleted")
			time.Sleep(10 * time.Second)

			_, err = authstore.GetAuthenticationByClientToken(ctx, clientToken)
			require.Error(t, err, "token should not be fetchable")
		})
	}
}
