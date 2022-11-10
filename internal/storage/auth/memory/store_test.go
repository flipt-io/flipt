package memory

import (
	"testing"

	"go.flipt.io/flipt/internal/storage/auth"
	storagetesting "go.flipt.io/flipt/internal/storage/testing"
)

func TestAuthenticationStoreHarness(t *testing.T) {
	storagetesting.TestAuthenticationStoreHarness(t, func(t *testing.T) auth.Store {
		return NewStore()
	})
}
