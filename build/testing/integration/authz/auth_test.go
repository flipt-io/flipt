package authz

import (
	"testing"

	"go.flipt.io/build/testing/integration"
)

func TestAuthz(t *testing.T) {
	integration.Harness(t, func(t *testing.T, opts integration.TestOpts) {
		Common(t, opts)
	})
}
