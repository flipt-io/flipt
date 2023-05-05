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
		ns, err := sdk.Flipt().GetNamespace(ctx, &flipt.GetNamespaceRequest{
			Key: namespace,
		})
		require.NoError(t, err)

		assert.Equal(t, namespace, ns.Key)

		expected := "Default"
		if namespace != "" && namespace != "default" {
			expected = namespace
		}
		assert.Equal(t, expected, ns.Name)

		require.NoError(t, err)
		flags, err := sdk.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{
			NamespaceKey: namespace,
		})
		require.NoError(t, err)

		require.Len(t, flags.Flags, 50)

		flag := flags.Flags[0]
		assert.Equal(t, namespace, flag.NamespaceKey)
		assert.Equal(t, "flag_001", flag.Key)
		assert.Equal(t, "FLAG_001", flag.Name)
		assert.Equal(t, "Some Description", flag.Description)
		assert.True(t, flag.Enabled)

		require.Len(t, flag.Variants, 2)

		assert.Equal(t, "variant_1", flag.Variants[0].Key)
		assert.Equal(t, "variant_2", flag.Variants[1].Key)

		segments, err := sdk.Flipt().ListSegments(ctx, &flipt.ListSegmentRequest{
			NamespaceKey: namespace,
		})

		require.Len(t, segments.Segments, 50)
	})
}
