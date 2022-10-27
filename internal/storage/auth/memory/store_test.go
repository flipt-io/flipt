package memory

import (
	"testing"

	"go.flipt.io/flipt/internal/storage"
	storagetesting "go.flipt.io/flipt/internal/storage/testing"
)

func TestAuthenticationStoreHarness(t *testing.T) {
	storagetesting.TestAuthenticationStoreHarness(t, func(t *testing.T) storage.AuthenticationStore {
		return NewStore()
	})
}
