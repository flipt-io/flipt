package memory

import (
	"testing"

	"go.flipt.io/flipt/internal/storage/auth"
	authtesting "go.flipt.io/flipt/internal/storage/auth/testing"
)

func TestAuthenticationStoreHarness(t *testing.T) {
	authtesting.TestAuthenticationStoreHarness(t, func(t *testing.T) auth.Store {
		return NewStore()
	})
}
