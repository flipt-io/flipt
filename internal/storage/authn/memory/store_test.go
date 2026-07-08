package memory

import (
	"context"
	stdErrors "errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/authn"
	authtesting "go.flipt.io/flipt/internal/storage/authn/testing"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestOAuthEntityString(t *testing.T) {
	tests := []struct {
		name   string
		entity func() *oauthEntity
		check  func(t *testing.T, s string)
	}{
		{
			name: "does not leak idToken",
			entity: func() *oauthEntity {
				return &oauthEntity{
					Authentication: &rpcauth.Authentication{Id: "test-id"},
					idToken:        &authn.IDTokenData{IDToken: "secret-jwt"},
				}
			},
			check: func(t *testing.T, s string) {
				assert.Contains(t, s, "test-id")
				assert.NotContains(t, s, "secret-jwt")
			},
		},
		{
			name: "nil embedded proto",
			entity: func() *oauthEntity {
				return &oauthEntity{
					idToken: &authn.IDTokenData{IDToken: "secret-jwt"},
				}
			},
			check: func(t *testing.T, s string) {
				assert.NotContains(t, s, "secret-jwt")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := fmt.Sprint(tt.entity())
			tt.check(t, s)
		})
	}
}

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
			Sid:       "session-123",
			Provider:  "google",
			IDToken:   "my-id-token-jwt",
		})
		require.NoError(t, err)
		require.NotEmpty(t, token)
		require.NotNil(t, created)

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

			authStored, err := store.ListAuthentications(ctx, &storage.ListRequest[authn.ListAuthenticationsPredicate]{
				Predicate: authn.ListAuthenticationsPredicate{
					Method:   new(rpcauth.Method_METHOD_OIDC),
					Provider: new("google"),
					Sid:      new("session-123"),
				},
			})
			require.NoError(t, err)
			require.Empty(t, authStored.Results)

			_, err = store.GetIDToken(ctx, created.Id)
			var notFound errors.ErrNotFound
			require.ErrorAs(t, err, &notFound)
		})
	})
}

func TestListAuthentications(t *testing.T) {
	store := NewStore(zap.NewNop())
	cleanup(t, store)

	ctx := t.Context()
	now := time.Now().UTC()

	// seed various authentications with different predicate values
	seeded := make([]*rpcauth.Authentication, 0, 5)

	for _, s := range []struct {
		method   rpcauth.Method
		provider string
		sid      string
		sub      string
		expires  time.Duration
	}{
		{rpcauth.Method_METHOD_OIDC, "google", "sid-a", "sub-1", time.Hour},
		{rpcauth.Method_METHOD_OIDC, "google", "sid-b", "sub-2", 2 * time.Hour},
		{rpcauth.Method_METHOD_OIDC, "github", "sid-c", "sub-1", 3 * time.Hour},
		{rpcauth.Method_METHOD_TOKEN, "", "", "", time.Hour},
		{rpcauth.Method_METHOD_KUBERNETES, "", "", "", time.Hour},
	} {
		_, auth, err := store.CreateAuthentication(ctx, &authn.CreateAuthenticationRequest{
			Method:    s.method,
			ExpiresAt: timestamppb.New(now.Add(s.expires)),
			Provider:  s.provider,
			Sid:       s.sid,
			Sub:       s.sub,
		})
		require.NoError(t, err)
		seeded = append(seeded, auth)
	}

	tests := []struct {
		name      string
		predicate authn.ListAuthenticationsPredicate
		expected  []int // indices into seeded that should match
	}{
		{
			name:      "no predicate returns all",
			predicate: authn.ListAuthenticationsPredicate{},
			expected:  []int{0, 1, 2, 3, 4},
		},
		{
			name: "filter by method OIDC",
			predicate: authn.ListAuthenticationsPredicate{
				Method: new(rpcauth.Method_METHOD_OIDC),
			},
			expected: []int{0, 1, 2},
		},
		{
			name: "filter by method TOKEN",
			predicate: authn.ListAuthenticationsPredicate{
				Method: new(rpcauth.Method_METHOD_TOKEN),
			},
			expected: []int{3},
		},
		{
			name: "filter by unknown method returns none",
			predicate: authn.ListAuthenticationsPredicate{
				Method: new(rpcauth.Method_METHOD_NONE),
			},
			expected: nil,
		},
		{
			name: "filter by provider",
			predicate: authn.ListAuthenticationsPredicate{
				Provider: new("google"),
			},
			expected: []int{0, 1},
		},
		{
			name: "filter by provider no matches",
			predicate: authn.ListAuthenticationsPredicate{
				Provider: new("gitlab"),
			},
			expected: nil,
		},
		{
			name: "filter by sid",
			predicate: authn.ListAuthenticationsPredicate{
				Sid: new("sid-a"),
			},
			expected: []int{0},
		},
		{
			name: "filter by sid empty string",
			predicate: authn.ListAuthenticationsPredicate{
				Sid: new(""),
			},
			expected: []int{3, 4},
		},
		{
			name: "filter by sub",
			predicate: authn.ListAuthenticationsPredicate{
				Sub: new("sub-1"),
			},
			expected: []int{0, 2},
		},
		{
			name: "filter by sub empty string",
			predicate: authn.ListAuthenticationsPredicate{
				Sub: new(""),
			},
			expected: []int{3, 4},
		},
		{
			name: "filter by method + provider",
			predicate: authn.ListAuthenticationsPredicate{
				Method:   new(rpcauth.Method_METHOD_OIDC),
				Provider: new("google"),
			},
			expected: []int{0, 1},
		},
		{
			name: "filter by method + provider + sid",
			predicate: authn.ListAuthenticationsPredicate{
				Method:   new(rpcauth.Method_METHOD_OIDC),
				Provider: new("google"),
				Sid:      new("sid-a"),
			},
			expected: []int{0},
		},
		{
			name: "filter by provider + sub",
			predicate: authn.ListAuthenticationsPredicate{
				Provider: new("google"),
				Sub:      new("sub-1"),
			},
			expected: []int{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.ListAuthentications(ctx, &storage.ListRequest[authn.ListAuthenticationsPredicate]{
				Predicate: tt.predicate,
				QueryParams: storage.QueryParams{
					Limit: 100,
				},
			})
			require.NoError(t, err)

			if tt.expected == nil {
				require.Empty(t, got.Results)
				return
			}

			require.Len(t, got.Results, len(tt.expected))
			for i, idx := range tt.expected {
				assert.Equal(t, seeded[idx].Id, got.Results[i].Id, "result %d", i)
			}
		})
	}
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
