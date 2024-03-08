package memory

import (
	"testing"

	"go.flipt.io/flipt/internal/storage/authn"
	authtesting "go.flipt.io/flipt/internal/storage/authn/testing"
)

func TestAuthenticationStoreHarness(t *testing.T) {
	authtesting.TestAuthenticationStoreHarness(t, func(t *testing.T) authn.Store {
		return NewStore()
	})
}
