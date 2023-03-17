package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/sdk"
)

func Harness(t *testing.T, fn func(t *testing.T) sdk.SDK) {
	client := fn(t)

	ctx := context.Background()

	t.Log("Create a new flag with key \"test\".")

	created, err := client.Flipt().CreateFlag(ctx, &flipt.CreateFlagRequest{
		Key:         "test",
		Name:        "Test",
		Description: "This is a test flag",
		Enabled:     true,
	})
	require.NoError(t, err)

	assert.Equal(t, "test", created.Key)
	assert.Equal(t, "Test", created.Name)
	assert.Equal(t, "This is a test flag", created.Description)
	assert.True(t, created.Enabled, "Flag should be enabled")

	t.Log("Retrieve flag with key \"test\".")

	flag, err := client.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{Key: "test"})
	require.NoError(t, err)

	assert.Equal(t, created, flag)

	t.Log("Update flag with key \"test\".")

	updated, err := client.Flipt().UpdateFlag(ctx, &flipt.UpdateFlagRequest{
		Key:  "test",
		Name: "Test 2",
	})
	require.NoError(t, err)

	assert.Equal(t, "Test 2", updated.Name)

	t.Log("List all flags.")

	flags, err := client.Flipt().ListFlags(ctx, &flipt.ListFlagRequest{})
	require.NoError(t, err)

	assert.Len(t, flags.Flags, 1)
	assert.Equal(t, updated.Key, flags.Flags[0].Key)
	assert.Equal(t, updated.Name, flags.Flags[0].Name)
	assert.Equal(t, updated.Description, flags.Flags[0].Description)

	for _, key := range []string{"one", "two"} {
		t.Logf("Create variant with key %q.", key)

		createdVariant, err := client.Flipt().CreateVariant(ctx, &flipt.CreateVariantRequest{
			Key:     key,
			Name:    key,
			FlagKey: "test",
		})
		require.NoError(t, err)

		assert.Equal(t, key, createdVariant.Key)
		assert.Equal(t, key, createdVariant.Name)
	}

	t.Log("Get flag \"test\" and check variants.")

	flag, err = client.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{Key: "test"})
	require.NoError(t, err)

	assert.Len(t, flag.Variants, 2)

	t.Log("Update variant \"one\" (rename from \"one\" to \"One\")")

	updatedVariant, err := client.Flipt().UpdateVariant(ctx, &flipt.UpdateVariantRequest{
		FlagKey: "test",
		Id:      flag.Variants[0].Id,
		Key:     "one",
		Name:    "One",
	})
	require.NoError(t, err)

	assert.Equal(t, "one", updatedVariant.Key)
	assert.Equal(t, "One", updatedVariant.Name)
}
