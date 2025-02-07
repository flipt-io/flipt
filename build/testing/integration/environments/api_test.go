package environments_test

import (
	"context"
	"testing"

	"go.flipt.io/build/testing/integration"
	"go.flipt.io/build/testing/integration/environments"
)

func TestAPI(t *testing.T) {
	integration.Harness(t, func(t *testing.T, opts integration.TestOpts) {
		ctx := context.Background()

		environments.API(t, ctx, opts)
	})
}
