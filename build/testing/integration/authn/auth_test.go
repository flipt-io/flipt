package authn

import (
	"testing"

	"go.flipt.io/flipt/build/testing/integration"
)

func TestAuthn(t *testing.T) {
	integration.Harness(t, func(t *testing.T, opts integration.TestOpts) {
		Common(t, opts)
	})
}
