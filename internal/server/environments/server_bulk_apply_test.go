package environments

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/coss/license"
	"go.flipt.io/flipt/internal/product"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestBulkApplyResources_PartialFailurePerEnvironment(t *testing.T) {
	ctx := t.Context()

	licenseManager := license.NewMockManager(t)
	licenseManager.EXPECT().Product().Return(product.Pro)

	envA := NewMockEnvironment(t)
	envB := NewMockEnvironment(t)

	envA.EXPECT().
		Update(ctx, "rev-1", NewResourceType("flipt.core", "Flag"), mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, _ ResourceType, fn UpdateFunc) (string, error) {
			store := newTestResourceStore()
			require.NoError(t, fn(ctx, store))
			return "rev-2", nil
		}).
		Once()

	envB.EXPECT().
		Update(ctx, "", NewResourceType("flipt.core", "Flag"), mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, _ ResourceType, fn UpdateFunc) (string, error) {
			store := newTestResourceStore()
			require.NoError(t, fn(ctx, store))
			return "", errors.New("commit failed")
		}).
		Once()

	s := &Server{
		logger:         zap.NewNop(),
		licenseManager: licenseManager,
		envs: &EnvironmentStore{
			byKey: map[string]Environment{
				"env-a": envA,
				"env-b": envB,
			},
		},
	}

	resp, err := s.BulkApplyResources(ctx, &rpcenvironments.BulkApplyResourcesRequest{
		EnvironmentKey: "env-a",
		EnvironmentKeys: []string{
			"env-a",
			"missing-env",
			"env-b",
		},
		NamespaceKeys: []string{"default"},
		Operation:     rpcenvironments.BulkOperation_BULK_OPERATION_CREATE,
		TypeUrl:       "flipt.core.Flag",
		Key:           "feature_a",
		Payload:       &anypb.Any{TypeUrl: "flipt.core.Flag"},
		OnConflict:    rpcenvironments.ConflictStrategy_CONFLICT_STRATEGY_OVERWRITE,
		Revision:      "rev-1",
	})

	require.NoError(t, err)
	assert.Equal(t, "rev-2", resp.Revision)
	require.Len(t, resp.Results, 3)

	byEnv := map[string]*rpcenvironments.BulkApplyNamespaceResult{}
	for _, result := range resp.Results {
		byEnv[result.EnvironmentKey] = result
	}

	require.Contains(t, byEnv, "env-a")
	require.Contains(t, byEnv, "missing-env")
	require.Contains(t, byEnv, "env-b")

	assert.Equal(
		t,
		rpcenvironments.OperationStatus_OPERATION_STATUS_SUCCESS,
		byEnv["env-a"].Status,
	)
	assert.Equal(
		t,
		rpcenvironments.OperationStatus_OPERATION_STATUS_FAILED,
		byEnv["missing-env"].Status,
	)
	assert.Contains(t, byEnv["missing-env"].GetError(), "environment")
	assert.Equal(
		t,
		rpcenvironments.OperationStatus_OPERATION_STATUS_FAILED,
		byEnv["env-b"].Status,
	)
	assert.Contains(t, byEnv["env-b"].GetError(), "commit failed")
}

func TestBulkApplyResources_NoOpEmptyCommitPreservesSkippedResults(t *testing.T) {
	ctx := t.Context()

	licenseManager := license.NewMockManager(t)
	licenseManager.EXPECT().Product().Return(product.Pro)

	env := NewMockEnvironment(t)
	env.EXPECT().
		Update(ctx, "rev-1", NewResourceType("flipt.core", "Flag"), mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, _ ResourceType, fn UpdateFunc) (string, error) {
			store := newTestResourceStore()
			store.resources[testResourceKey("default", "feature_a")] = &rpcenvironments.Resource{
				NamespaceKey: "default",
				Key:          "feature_a",
			}
			require.NoError(t, fn(ctx, store))
			return "", errors.New("cannot create empty commit")
		}).
		Once()

	s := &Server{
		logger:         zap.NewNop(),
		licenseManager: licenseManager,
		envs: &EnvironmentStore{
			byKey: map[string]Environment{
				"env-a": env,
			},
		},
	}

	resp, err := s.BulkApplyResources(ctx, &rpcenvironments.BulkApplyResourcesRequest{
		EnvironmentKey: "env-a",
		NamespaceKeys:  []string{"default"},
		Operation:      rpcenvironments.BulkOperation_BULK_OPERATION_CREATE,
		TypeUrl:        "flipt.core.Flag",
		Key:            "feature_a",
		Payload:        &anypb.Any{TypeUrl: "flipt.core.Flag"},
		OnConflict:     rpcenvironments.ConflictStrategy_CONFLICT_STRATEGY_SKIP,
		Revision:       "rev-1",
	})

	require.NoError(t, err)
	assert.Equal(t, "rev-1", resp.Revision)
	require.Len(t, resp.Results, 1)
	assert.Equal(
		t,
		rpcenvironments.OperationStatus_OPERATION_STATUS_SKIPPED,
		resp.Results[0].Status,
	)
	assert.Equal(t, "env-a", resp.Results[0].EnvironmentKey)
	assert.Equal(t, "default", resp.Results[0].NamespaceKey)
}

func TestBulkApplyResources_DeduplicatesEnvironmentKeys(t *testing.T) {
	ctx := t.Context()

	licenseManager := license.NewMockManager(t)
	licenseManager.EXPECT().Product().Return(product.Pro)

	envA := NewMockEnvironment(t)
	envB := NewMockEnvironment(t)

	envA.EXPECT().
		Update(ctx, "rev-1", NewResourceType("flipt.core", "Flag"), mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, _ ResourceType, fn UpdateFunc) (string, error) {
			store := newTestResourceStore()
			require.NoError(t, fn(ctx, store))
			return "rev-2", nil
		}).
		Once()

	envB.EXPECT().
		Update(ctx, "", NewResourceType("flipt.core", "Flag"), mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, _ ResourceType, fn UpdateFunc) (string, error) {
			store := newTestResourceStore()
			require.NoError(t, fn(ctx, store))
			return "rev-b", nil
		}).
		Once()

	s := &Server{
		logger:         zap.NewNop(),
		licenseManager: licenseManager,
		envs: &EnvironmentStore{
			byKey: map[string]Environment{
				"env-a": envA,
				"env-b": envB,
			},
		},
	}

	resp, err := s.BulkApplyResources(ctx, &rpcenvironments.BulkApplyResourcesRequest{
		EnvironmentKey: "env-a",
		EnvironmentKeys: []string{
			"env-a",
			"env-a",
			"env-b",
		},
		NamespaceKeys: []string{"default"},
		Operation:     rpcenvironments.BulkOperation_BULK_OPERATION_CREATE,
		TypeUrl:       "flipt.core.Flag",
		Key:           "feature_a",
		Payload:       &anypb.Any{TypeUrl: "flipt.core.Flag"},
		OnConflict:    rpcenvironments.ConflictStrategy_CONFLICT_STRATEGY_OVERWRITE,
		Revision:      "rev-1",
	})

	require.NoError(t, err)
	assert.Equal(t, "rev-2", resp.Revision)
	require.Len(t, resp.Results, 2)
}
