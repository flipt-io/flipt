package environments

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/build/testing/integration"
	"go.flipt.io/flipt/rpc/flipt/core"
	"go.flipt.io/flipt/rpc/v2/environments"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestCopyResource(t *testing.T) {
	integration.Harness(t, func(t *testing.T, opts integration.TestOpts) {
		ctx := context.Background()

		envClient := opts.TokenClientV2(t).Environments()

		const env = integration.DefaultEnvironment

		// Get initial revision.
		nl, err := envClient.ListNamespaces(ctx, &environments.ListNamespacesRequest{
			EnvironmentKey: env,
		})
		require.NoError(t, err)
		revision := nl.Revision

		// Ensure "copy-target" namespace exists.
		created, err := envClient.CreateNamespace(ctx, &environments.UpdateNamespaceRequest{
			EnvironmentKey: env,
			Key:            "copy-target",
			Name:           "Copy Target",
			Revision:       revision,
		})
		require.NoError(t, err)
		revision = created.Revision

		// Create a flag in the default namespace to copy.
		flagPayload, err := anypb.New(&core.Flag{
			Key:     "copy-me",
			Name:    "Copy Me",
			Enabled: true,
		})
		require.NoError(t, err)

		stripAnyTypePrefix(t, flagPayload)

		srcFlag, err := envClient.CreateResource(ctx, &environments.UpdateResourceRequest{
			EnvironmentKey: env,
			NamespaceKey:   integration.DefaultNamespace,
			Key:            "copy-me",
			Payload:        flagPayload,
			Revision:       revision,
		})
		require.NoError(t, err)
		revision = srcFlag.Revision

		t.Run("copy to different namespace", func(t *testing.T) {
			resp, err := envClient.CopyResource(ctx, &environments.CopyResourceRequest{
				EnvironmentKey:       env,
				NamespaceKey:         "copy-target",
				SourceEnvironmentKey: env,
				SourceNamespaceKey:   integration.DefaultNamespace,
				TypeUrl:              "flipt.core.Flag",
				Key:                  "copy-me",
				Revision:             revision,
			})
			require.NoError(t, err)

			assert.Equal(t, environments.OperationStatus_OPERATION_STATUS_SUCCESS, resp.Status)
			assert.Equal(t, "copy-me", resp.Resource.Key)
			assert.Equal(t, "copy-target", resp.Resource.NamespaceKey)
			assert.NotEmpty(t, resp.Revision)

			revision = resp.Revision

			// Verify the resource exists in the target namespace.
			got, err := envClient.GetResource(ctx, &environments.GetResourceRequest{
				EnvironmentKey: env,
				NamespaceKey:   "copy-target",
				TypeUrl:        "flipt.core.Flag",
				Key:            "copy-me",
			})
			require.NoError(t, err)
			assert.Equal(t, "copy-me", got.Resource.Key)
		})

		t.Run("conflict strategy FAIL", func(t *testing.T) {
			_, err := envClient.CopyResource(ctx, &environments.CopyResourceRequest{
				EnvironmentKey:       env,
				NamespaceKey:         "copy-target",
				SourceEnvironmentKey: env,
				SourceNamespaceKey:   integration.DefaultNamespace,
				TypeUrl:              "flipt.core.Flag",
				Key:                  "copy-me",
				OnConflict:           environments.ConflictStrategy_CONFLICT_STRATEGY_FAIL,
				Revision:             revision,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "AlreadyExists")
		})

		t.Run("conflict strategy SKIP", func(t *testing.T) {
			resp, err := envClient.CopyResource(ctx, &environments.CopyResourceRequest{
				EnvironmentKey:       env,
				NamespaceKey:         "copy-target",
				SourceEnvironmentKey: env,
				SourceNamespaceKey:   integration.DefaultNamespace,
				TypeUrl:              "flipt.core.Flag",
				Key:                  "copy-me",
				OnConflict:           environments.ConflictStrategy_CONFLICT_STRATEGY_SKIP,
				Revision:             revision,
			})
			require.NoError(t, err)
			assert.Equal(t, environments.OperationStatus_OPERATION_STATUS_SKIPPED, resp.Status)
			revision = resp.Revision
		})

		t.Run("conflict strategy OVERWRITE", func(t *testing.T) {
			resp, err := envClient.CopyResource(ctx, &environments.CopyResourceRequest{
				EnvironmentKey:       env,
				NamespaceKey:         "copy-target",
				SourceEnvironmentKey: env,
				SourceNamespaceKey:   integration.DefaultNamespace,
				TypeUrl:              "flipt.core.Flag",
				Key:                  "copy-me",
				OnConflict:           environments.ConflictStrategy_CONFLICT_STRATEGY_OVERWRITE,
				Revision:             revision,
			})
			require.NoError(t, err)
			assert.Equal(t, environments.OperationStatus_OPERATION_STATUS_SUCCESS, resp.Status)
			revision = resp.Revision
		})

		// Cleanup: delete copy-target namespace.
		_, err = envClient.DeleteNamespace(ctx, &environments.DeleteNamespaceRequest{
			EnvironmentKey: env,
			Key:            "copy-target",
			Revision:       revision,
		})
		require.NoError(t, err)
	})
}
