package readonly_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/build/testing/integration"
	"go.flipt.io/flipt/rpc/flipt"
	sdk "go.flipt.io/flipt/sdk/go"
)

// TestReadOnly is a suite of tests which presumes all the data found in the local testdata
// folder has been loaded into the target instance being tested.
// It then exercises a bunch of read operations via the provided SDK in the target namespace.
func TestReadOnly(t *testing.T) {
	integration.Harness(t, func(t *testing.T, sdk sdk.SDK, namespace string, _ bool) {
		ctx := context.Background()
		flags, err := sdk.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
			NamespaceKey: namespace,
		})
		require.NoError(t, err)

		assert.Len(t, flags.Flags, 1)
	})
}
