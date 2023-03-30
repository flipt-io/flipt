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

		require.Len(t, flags.Flags, 1)

		flag := flags.Flags[0]
		assert.Equal(t, namespace, flag.NamespaceKey)
		assert.Equal(t, "color", flag.Key)
		assert.Equal(t, "Color", flag.Name)
		assert.Equal(t, "This flag represents two colors.", flag.Description)
		assert.True(t, flag.Enabled)

		require.Len(t, flag.Variants, 2)

		assert.Equal(t, "blue", flag.Variants[0].Key)
		assert.Equal(t, "Blue", flag.Variants[0].Name)

		assert.Equal(t, "red", flag.Variants[1].Key)
		assert.Equal(t, "Red", flag.Variants[1].Name)

		segments, err := sdk.Flipt().ListSegments(ctx, &flipt.ListSegmentRequest{
			NamespaceKey: namespace,
		})

		require.Len(t, segments.Segments, 2)
	})
}
