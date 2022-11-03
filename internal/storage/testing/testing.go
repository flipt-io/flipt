package testing

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/auth"
)

func TestAuthenticationStoreHarness(t *testing.T, fn func(t *testing.T) storage.AuthenticationStore) {
	t.Helper()

	store := fn(t)

	const authCount = 100
	var (
		ctx   = context.TODO()
		auths = make([]*auth.Authentication, authCount)
		index = make(map[string]int, authCount)
	)

	t.Run(fmt.Sprintf("Create %d authentications", authCount), func(t *testing.T) {
		for i := 0; i < 100; i++ {
			token, auth, err := store.CreateAuthentication(ctx, &storage.CreateAuthenticationRequest{
				Method: auth.Method_TOKEN,
				Metadata: map[string]string{
					"name":        fmt.Sprintf("foo_%d", i),
					"description": "bar",
				},
			})
			require.NoError(t, err)

			auths[i] = auth

			// ensure token does not already exists
			_, ok := index[token]
			require.False(t, ok, "Token already exists")
			index[token] = i
		}
	})

	t.Run("Get each authentication by ID", func(t *testing.T) {
		// ensure each auth can be re-retrieved by the client token
		for token, i := range index {
			auth, err := store.GetAuthenticationByClientToken(ctx, token)
			require.NoError(t, err)

			assert.Equal(t, auths[i], auth)
		}
	})

	t.Run("List all authentications (3 per page ascending)", func(t *testing.T) {
		// ensure all auths match
		all, err := storage.ListAll(ctx, store.ListAuthentications, storage.ListAllParams{
			PerPage: 3,
		})
		require.NoError(t, err)
		assert.Equal(t, auths, all)
	})

	t.Run("List all authentications (6 per page descending)", func(t *testing.T) {
		// ensure order descending matches
		all, err := storage.ListAll(ctx, store.ListAuthentications, storage.ListAllParams{
			PerPage: 6,
			Order:   storage.OrderDesc,
		})
		require.NoError(t, err)
		for i := 0; i < len(auths)/2; i++ {
			j := len(auths) - 1 - i
			auths[i], auths[j] = auths[j], auths[i]
		}
		assert.Equal(t, auths, all)
	})
}
