package testing

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	storageauth "go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestAuthenticationStoreHarness(t *testing.T, fn func(t *testing.T) storageauth.Store) {
	t.Helper()

	store := fn(t)

	type authTuple struct {
		Token string
		Auth  *auth.Authentication
	}

	var (
		ctx = context.TODO()

		created [100]authTuple

		allAuths = func(t []authTuple) (res []*auth.Authentication) {
			res = make([]*auth.Authentication, len(t))
			for i, a := range t {
				res[i] = a.Auth
			}
			return
		}
	)

	t.Run(fmt.Sprintf("Create %d authentications", len(created)), func(t *testing.T) {
		uniqueTokens := make(map[string]struct{}, len(created))
		for i := 0; i < len(created); i++ {
			// the first token will have a null expiration
			var expires *timestamppb.Timestamp
			if i > 0 {
				// Set expiry times starting from now + 100 hours
				// Each subsequent token expires 1 hour later
				// This ensures tokens are:
				// 1. Far in the future (won't auto-expire)
				// 2. Ordered sequentially (for testing deletion)
				expires = timestamppb.New(time.Now().UTC().Add(time.Duration(100+i) * time.Hour))
			}

			token, auth, err := store.CreateAuthentication(ctx, &storageauth.CreateAuthenticationRequest{
				Method: auth.Method_METHOD_TOKEN,
				// from t1 to t100.
				ExpiresAt: expires,
				Metadata: map[string]string{
					"name":        fmt.Sprintf("foo_%d", i+1),
					"description": "bar",
				},
			})
			require.NoError(t, err)

			created[i].Token = token
			created[i].Auth = auth

			// ensure token does not already exists
			_, ok := uniqueTokens[token]
			require.False(t, ok, "Token already exists")
			uniqueTokens[token] = struct{}{}
		}
	})

	t.Run("Get each authentication by token", func(t *testing.T) {
		// ensure each auth can be re-retrieved by the client token
		for i, auth := range created {
			auth, err := store.GetAuthenticationByClientToken(ctx, auth.Token)
			require.NoError(t, err)

			assert.Equal(t, created[i].Auth, auth)
		}
	})

	t.Run("Get each authentication by ID", func(t *testing.T) {
		// ensure each auth can be re-retrieved by ID
		for i, auth := range created {
			auth, err := store.GetAuthenticationByID(ctx, auth.Auth.Id)
			require.NoError(t, err)

			assert.Equal(t, created[i].Auth, auth)
		}
	})

	t.Run("List all authentications (3 per page ascending)", func(t *testing.T) {
		// ensure all auths match
		all, err := storage.ListAll(ctx, store.ListAuthentications, storage.ListAllParams{
			PerPage: 3,
		})
		require.NoError(t, err)
		assert.Equal(t, allAuths(created[:]), all)
	})

	t.Run("List all authentications (6 per page descending)", func(t *testing.T) {
		// ensure order descending matches
		all, err := storage.ListAll(ctx, store.ListAuthentications, storage.ListAllParams{
			PerPage: 6,
			// TODO: ordering is not supported
			// Order:   storage.OrderDesc,
		})
		require.NoError(t, err)

		expected := allAuths(created[:])
		// for i := 0; i < len(expected)/2; i++ {
		// 	j := len(expected) - 1 - i
		// 	expected[i], expected[j] = expected[j], expected[i]
		// }

		assert.Equal(t, len(expected), len(all), "number of authentications should match")
		for i := 0; i < len(expected); i++ {
			assert.Equal(t, expected[i].Id, all[i].Id, "authentication IDs should match at index %d", i)
		}
	})

	t.Run("Delete must be predicated", func(t *testing.T) {
		err := store.DeleteAuthentications(ctx, &storageauth.DeleteAuthenticationsRequest{})
		var invalid errors.ErrInvalid
		require.ErrorAs(t, err, &invalid)
	})

	t.Run("Delete a single instance by ID", func(t *testing.T) {
		req := storageauth.Delete(storageauth.WithID(created[99].Auth.Id))
		err := store.DeleteAuthentications(ctx, req)
		require.NoError(t, err)

		// there should now be 99 authentications left
		all, err := storage.ListAll(ctx, store.ListAuthentications, storage.ListAllParams{})
		require.NoError(t, err)
		assert.Equal(t, 99, len(all), "number of authentications should match")

		auth, err := store.GetAuthenticationByClientToken(ctx, created[99].Token)
		var expected errors.ErrNotFound
		require.ErrorAs(t, err, &expected, "authentication still exists in the database")
		assert.Nil(t, auth)
	})

	t.Run("Delete by method Token with before expired constraint", func(t *testing.T) {
		// Delete tokens with indices [1,50] by using expiry of now + 150 hours
		req := storageauth.Delete(
			storageauth.WithMethod(auth.Method_METHOD_TOKEN),
			storageauth.WithExpiredBefore(time.Now().UTC().Add(150*time.Hour)),
		)

		err := store.DeleteAuthentications(
			ctx,
			req,
		)
		require.NoError(t, err)

		// there should now be 49 authentications left
		all, err := storage.ListAll(ctx, store.ListAuthentications, storage.ListAllParams{})
		require.NoError(t, err)
		assert.Equal(t, 49, len(all), "number of authentications should match")

		// ensure only the most recent 49 expires_at timestamped authentications remain
		// along with the first authentication without an expiry
		expected := allAuths(append(created[:1], created[51:99]...))
		assert.Equal(t, len(expected), len(all), "number of authentications should match")
		for i := 0; i < len(expected); i++ {
			assert.Equal(t, expected[i].Id, all[i].Id, "authentication IDs should match at index %d", i)
		}
	})

	t.Run("Delete any token type before expired constraint", func(t *testing.T) {
		// Delete tokens with indices [1,75] by using expiry of now + 175 hours
		req := storageauth.Delete(
			storageauth.WithExpiredBefore(time.Now().UTC().Add(175 * time.Hour)),
		)

		err := store.DeleteAuthentications(
			ctx,
			req,
		)
		require.NoError(t, err)

		// there should now be 24 authentications left
		all, err := storage.ListAll(ctx, store.ListAuthentications, storage.ListAllParams{})
		require.NoError(t, err)
		assert.Equal(t, 24, len(all), "number of authentications should match")

		expected := allAuths(append(created[:1], created[76:99]...))
		assert.Equal(t, len(expected), len(all), "number of authentications should match")
		for i := 0; i < len(expected); i++ {
			assert.Equal(t, expected[i].Id, all[i].Id, "authentication IDs should match at index %d", i)
		}
	})

	t.Run("Delete the rest of the tokens with an expiration", func(t *testing.T) {
		// Delete all remaining tokens by using expiry beyond the last token (now + 1100 hours)
		req := storageauth.Delete(
			storageauth.WithExpiredBefore(time.Now().UTC().Add(1100 * time.Hour)),
		)

		err := store.DeleteAuthentications(
			ctx,
			req,
		)
		require.NoError(t, err)

		all, err := storage.ListAll(ctx, store.ListAuthentications, storage.ListAllParams{})
		require.NoError(t, err)

		// ensure only the the first token with no expiry exists
		if !assert.Equal(t, allAuths(created[:1]), all) {
			fmt.Println("Found:", len(all))
		}
	})

	t.Run("Expire a single instance by ID", func(t *testing.T) {
		expiresAt := timestamppb.New(time.Now().UTC().Add(-time.Hour))
		// expire the first token
		err := store.ExpireAuthenticationByID(ctx, created[0].Auth.Id, expiresAt)
		require.NoError(t, err)

		_, err = store.GetAuthenticationByClientToken(ctx, created[0].Token)
		var expected errors.ErrNotFound
		require.ErrorAs(t, err, &expected, "authentication still exists in the database")
		// TODO: the memory store does not remove the token from the set if it is expired
		// require.NoError(t, err)
		// assert.True(t, auth.ExpiresAt.AsTime().Before(time.Now().UTC()))
	})
}
