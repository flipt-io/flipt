package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/secrets"
)

func TestVaultProviderRegistration(t *testing.T) {
	t.Run("vault provider factory is registered", func(t *testing.T) {
		factory, exists := secrets.GetProviderFactory("vault")
		require.True(t, exists, "vault factory should be registered")
		assert.NotNil(t, factory, "vault factory should not be nil")
	})
}
