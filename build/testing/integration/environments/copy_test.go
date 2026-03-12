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

func TestBulkApplyResources_CopyEquivalent(t *testing.T) {
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

		srcResource, err := envClient.GetResource(ctx, &environments.GetResourceRequest{
			EnvironmentKey: env,
			NamespaceKey:   integration.DefaultNamespace,
			TypeUrl:        "flipt.core.Flag",
			Key:            "copy-me",
		})
		require.NoError(t, err)

		t.Run("copy to different namespace", func(t *testing.T) {
			resp, err := envClient.BulkApplyResources(ctx, &environments.BulkApplyResourcesRequest{
				EnvironmentKey: env,
				NamespaceKeys:  []string{"copy-target"},
				Operation:      environments.BulkOperation_BULK_OPERATION_CREATE,
				TypeUrl:        "flipt.core.Flag",
				Key:            "copy-me",
				Payload:        srcResource.Resource.Payload,
				Revision:       revision,
			})
			require.NoError(t, err)

			require.Len(t, resp.Results, 1)
			assert.Equal(t, environments.OperationStatus_OPERATION_STATUS_SUCCESS, resp.Results[0].Status)
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
			resp, err := envClient.BulkApplyResources(ctx, &environments.BulkApplyResourcesRequest{
				EnvironmentKey: env,
				NamespaceKeys:  []string{"copy-target"},
				Operation:      environments.BulkOperation_BULK_OPERATION_CREATE,
				TypeUrl:        "flipt.core.Flag",
				Key:            "copy-me",
				Payload:        srcResource.Resource.Payload,
				OnConflict:     environments.ConflictStrategy_CONFLICT_STRATEGY_FAIL,
				Revision:       revision,
			})
			require.NoError(t, err)
			require.Len(t, resp.Results, 1)
			assert.Equal(t, environments.OperationStatus_OPERATION_STATUS_FAILED, resp.Results[0].Status)
			require.NotNil(t, resp.Results[0].Error)
			assert.Contains(t, *resp.Results[0].Error, "already exists")
		})

		t.Run("conflict strategy SKIP", func(t *testing.T) {
			resp, err := envClient.BulkApplyResources(ctx, &environments.BulkApplyResourcesRequest{
				EnvironmentKey: env,
				NamespaceKeys:  []string{"copy-target"},
				Operation:      environments.BulkOperation_BULK_OPERATION_CREATE,
				TypeUrl:        "flipt.core.Flag",
				Key:            "copy-me",
				Payload:        srcResource.Resource.Payload,
				OnConflict:     environments.ConflictStrategy_CONFLICT_STRATEGY_SKIP,
				Revision:       revision,
			})
			require.NoError(t, err)
			require.Len(t, resp.Results, 1)
			assert.Equal(t, environments.OperationStatus_OPERATION_STATUS_SKIPPED, resp.Results[0].Status)
			revision = resp.Revision
		})

		t.Run("conflict strategy OVERWRITE", func(t *testing.T) {
			resp, err := envClient.BulkApplyResources(ctx, &environments.BulkApplyResourcesRequest{
				EnvironmentKey: env,
				NamespaceKeys:  []string{"copy-target"},
				Operation:      environments.BulkOperation_BULK_OPERATION_CREATE,
				TypeUrl:        "flipt.core.Flag",
				Key:            "copy-me",
				Payload:        srcResource.Resource.Payload,
				OnConflict:     environments.ConflictStrategy_CONFLICT_STRATEGY_OVERWRITE,
				Revision:       revision,
			})
			require.NoError(t, err)
			require.Len(t, resp.Results, 1)
			assert.Equal(t, environments.OperationStatus_OPERATION_STATUS_SUCCESS, resp.Results[0].Status)
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
