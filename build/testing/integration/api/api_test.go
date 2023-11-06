package api_test

import (
	"context"
	"testing"

	"go.flipt.io/flipt/build/testing/integration"
	"go.flipt.io/flipt/build/testing/integration/api"
	sdk "go.flipt.io/flipt/sdk/go"
)

func TestAPI(t *testing.T) {
	integration.Harness(t, func(t *testing.T, sdk sdk.SDK, opts integration.TestOpts) {
		ctx := context.Background()

		api.API(t, ctx, sdk, opts)

		// run extra tests in authenticated context
		if opts.Authenticated {
			api.Authenticated(t, sdk)
		}
	})
}
