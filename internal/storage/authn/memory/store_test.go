package memory

import (
	"context"
	stdErrors "errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage/authn"
	authtesting "go.flipt.io/flipt/internal/storage/authn/testing"
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
