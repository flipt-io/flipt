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
}
