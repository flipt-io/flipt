package api_test

import (
	"testing"

	"go.flipt.io/flipt/build/testing/integration"
	"go.flipt.io/flipt/build/testing/integration/api"
	sdk "go.flipt.io/flipt/sdk/go"
)

func TestAPI(t *testing.T) {
	integration.Harness(t, func(t *testing.T, sdk sdk.SDK, namespace string, authentication bool) {
		api.API(t, sdk, namespace)

		// run extra tests in authenticated context
		if authentication {
			api.Authenticated(t, sdk)
		}
	})
}
