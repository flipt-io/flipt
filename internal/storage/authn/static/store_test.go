package static

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/internal/storage/authn/memory"
	authtesting "go.flipt.io/flipt/internal/storage/authn/testing"
	"go.uber.org/zap/zaptest"
)

func TestAuthenticationStoreStatic(t *testing.T) {
	authtesting.TestAuthenticationStoreHarness(t, func(t *testing.T) authn.Store {
		store, err := NewStore(memory.NewStore(zaptest.NewLogger(t)), config.AuthenticationMethodTokenStorage{})
		require.NoError(t, err)

		return store
	})
}

func TestAuthenticationStoreStatic_HasToken(t *testing.T) {
	store, err := NewStore(memory.NewStore(zaptest.NewLogger(t)), config.AuthenticationMethodTokenStorage{
		Type: config.AuthenticationMethodTokenStorageTypeStatic,
		Tokens: []config.AuthenticationMethodStaticToken{
			{
				Name:       "sometoken",
				Credential: "somesecretstring",
			},
		},
	})
	require.NoError(t, err)

	auth, err := store.GetAuthenticationByClientToken(context.TODO(), "somesecretstring")
	require.NoError(t, err)

	assert.Equal(t, "sometoken", auth.Id)
}
