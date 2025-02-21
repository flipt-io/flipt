package git

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	serverenvironments "go.flipt.io/flipt/internal/server/environments"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
)

func TestReadOnlyEnvironment(t *testing.T) {
	ctx := context.Background()
	env := ReadOnlyEnvironment{}

	t.Run("PutNamespace returns read-only error", func(t *testing.T) {
		description := "A test namespace"
		ns := &rpcenvironments.Namespace{
			Key:         "test",
			Name:        "Test Namespace",
			Description: &description,
		}

		ref, err := env.PutNamespace(ctx, "main", ns)
		assert.Empty(t, ref)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrReadOnly)
		assert.Contains(t, err.Error(), "put namespace \"test\"")
	})

	t.Run("DeleteNamespace returns read-only error", func(t *testing.T) {
		ref, err := env.DeleteNamespace(ctx, "main", "test")
		assert.Empty(t, ref)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrReadOnly)
		assert.Contains(t, err.Error(), "delete namespace \"test\"")
	})

	t.Run("Update returns read-only error", func(t *testing.T) {
		ref, err := env.Update(ctx, "main", serverenvironments.NewResourceType("test", "Flag"), func(context.Context, serverenvironments.ResourceStore) error {
			return nil
		})
		assert.Empty(t, ref)
		require.Error(t, err)
		require.ErrorIs(t, err, ErrReadOnly)
		assert.Contains(t, err.Error(), "update resource type \"test.Flag\"")
	})
}
