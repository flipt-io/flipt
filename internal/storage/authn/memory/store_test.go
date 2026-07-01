package memory

import (
	"context"
	stdErrors "errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage/authn"
	authtesting "go.flipt.io/flipt/internal/storage/authn/testing"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAuthenticationStoreHarness(t *testing.T) {
	store := NewStore(zap.NewNop())
	cleanup(t, store)
	authtesting.TestAuthenticationStoreHarness(t, func(t *testing.T) authn.Store {
		return store
	})
}

func cleanup(t *testing.T, authStore *Store) {
	t.Helper()
	t.Cleanup(func() {
		require.NoError(t, authStore.Shutdown(context.Background())) // nolint: usetesting
	})
}

func TestSIDAndIDTokenStorage(t *testing.T) {
	store := NewStore(zap.NewNop())
	cleanup(t, store)

	ctx := t.Context()

	t.Run("Create authentication with SID and IDToken", func(t *testing.T) {
		token, created, err := store.CreateAuthentication(ctx, &authn.CreateAuthenticationRequest{
			Method:    rpcauth.Method_METHOD_OIDC,
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
				Method:    rpcauth.Method_METHOD_TOKEN,
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

func TestOAuthChallengePopUsesInjectedClock(t *testing.T) {
	var (
		nowMu sync.Mutex
		now   = time.Now().UTC()
	)

	store := NewStore(
		zap.NewNop(),
		WithNowFunc(func() *timestamppb.Timestamp {
			nowMu.Lock()
			defer nowMu.Unlock()
			return timestamppb.New(now)
		}),
		WithCleanupInterval(24*time.Hour),
	)
	cleanup(t, store)

	setNow := func(v time.Time) {
		nowMu.Lock()
		now = v.UTC()
		nowMu.Unlock()
	}

	ctx := t.Context()
	base := time.Now().UTC()

	setNow(base.Add(2 * time.Minute))

	require.NoError(t, store.PutOAuthChallenge(ctx, "state-pop", "verifier-pop", timestamppb.New(base.Add(1*time.Minute))))

	_, err := store.PopOAuthChallenge(ctx, "state-pop")
	var notFound errors.ErrNotFound
	require.ErrorAs(t, err, &notFound)
}

func TestOAuthChallengeCleanupRemovesExpiredChallenges(t *testing.T) {
	var (
		nowMu sync.Mutex
		now   = time.Now().UTC()
	)

	store := NewStore(
		zap.NewNop(),
		WithNowFunc(func() *timestamppb.Timestamp {
			nowMu.Lock()
			defer nowMu.Unlock()
			return timestamppb.New(now)
		}),
		WithCleanupInterval(10*time.Millisecond),
	)
	cleanup(t, store)

	setNow := func(v time.Time) {
		nowMu.Lock()
		now = v.UTC()
		nowMu.Unlock()
	}

	ctx := t.Context()
	base := time.Now().UTC()
	expiresAt := timestamppb.New(base.Add(1 * time.Minute))
	future := base.Add(2 * time.Minute)

	setNow(base)
	require.NoError(t, store.PutOAuthChallenge(ctx, "state-cleanup", "verifier-cleanup", expiresAt))

	setNow(future)

	require.Eventually(t, func() bool {
		setNow(base)
		defer setNow(future)

		value, err := store.PopOAuthChallenge(ctx, "state-cleanup")
		if err == nil {
			require.NoError(t, store.PutOAuthChallenge(ctx, "state-cleanup", value, expiresAt))
			return false
		}

		var notFound errors.ErrNotFound
		return stdErrors.As(err, &notFound)
	}, time.Second, 10*time.Millisecond)
}
