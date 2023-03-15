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

	created, err := client.Flipt().CreateFlag(ctx, &flipt.CreateFlagRequest{
		Key:         "test",
		Name:        "Test",
		Description: "This is a test flag",
	})
	require.NoError(t, err)

	assert.Equal(t, "test", created.Key)
	assert.Equal(t, "Test", created.Name)
	assert.Equal(t, "This is a test flag", created.Description)

	flag, err := client.Flipt().GetFlag(ctx, &flipt.GetFlagRequest{Key: "test"})
	require.NoError(t, err)

	assert.Equal(t, created, flag)
}
