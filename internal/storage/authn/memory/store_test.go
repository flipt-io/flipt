package memory

import (
	"testing"

	"go.flipt.io/flipt/internal/storage/authn"
	authtesting "go.flipt.io/flipt/internal/storage/authn/testing"
	"go.uber.org/zap"
)

func TestAuthenticationStoreHarness(t *testing.T) {
	authtesting.TestAuthenticationStoreHarness(t, func(t *testing.T) authn.Store {
		return NewStore(zap.NewNop())
	})
}
