package environments

import (
	"context"
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

func TestCopyNamespace_RequiresPro(t *testing.T) {
	ctx := t.Context()
	licenseManager := license.NewMockManager(t)
	licenseManager.EXPECT().Product().Return(product.OSS)

	s := &Server{
		logger:         zap.NewNop(),
		licenseManager: licenseManager,
		envs:           &EnvironmentStore{byKey: map[string]Environment{}},
	}

	_, err := s.CopyNamespace(ctx, &rpcenvironments.CopyNamespaceRequest{
		EnvironmentKey:       "target",
		NamespaceKey:         "default",
		SourceEnvironmentKey: "source",
		SourceNamespaceKey:   "default",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "requires a Flipt Pro license")
}

func TestCopyNamespace_SuccessCreatesResource(t *testing.T) {
	ctx := t.Context()
	licenseManager := license.NewMockManager(t)
	licenseManager.EXPECT().Product().Return(product.Pro)

	sourceEnv := NewMockEnvironment(t)
	targetEnv := NewMockEnvironment(t)

	sourceEnv.EXPECT().
		View(ctx, NewResourceType("flipt.core", "Flag"), mock.Anything).
		RunAndReturn(func(ctx context.Context, _ ResourceType, fn ViewFunc) error {
			store := newTestResourceStore()
			store.resources[testResourceKey("default", "flag-a")] = &rpcenvironments.Resource{
				NamespaceKey: "default",
				Key:          "flag-a",
				Payload:      &anypb.Any{TypeUrl: "flipt.core.Flag"},
			}
			return fn(ctx, store)
		}).
		Once()
	sourceEnv.EXPECT().
		View(ctx, NewResourceType("flipt.core", "Segment"), mock.Anything).
		RunAndReturn(func(ctx context.Context, _ ResourceType, fn ViewFunc) error {
			return fn(ctx, newTestResourceStore())
		}).
		Once()

	targetEnv.EXPECT().
		GetNamespace(ctx, "default").
		Return(&rpcenvironments.NamespaceResponse{}, nil).
		Once()
	targetEnv.EXPECT().
		View(ctx, NewResourceType("flipt.core", "Flag"), mock.Anything).
		RunAndReturn(func(ctx context.Context, _ ResourceType, fn ViewFunc) error {
			return fn(ctx, newTestResourceStore())
		}).
		Once()
	targetEnv.EXPECT().
		Update(ctx, "rev-1", NewResourceType("flipt.core", "Flag"), mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, _ ResourceType, fn UpdateFunc) (string, error) {
			store := newTestResourceStore()
			require.NoError(t, fn(ctx, store))
			assert.Equal(t, 1, store.createCalls)
			return "rev-2", nil
		}).
		Once()

	s := &Server{
		logger:         zap.NewNop(),
		licenseManager: licenseManager,
		envs: &EnvironmentStore{
			byKey: map[string]Environment{
				"source": sourceEnv,
				"target": targetEnv,
			},
		},
	}

	resp, err := s.CopyNamespace(ctx, &rpcenvironments.CopyNamespaceRequest{
		EnvironmentKey:       "target",
		NamespaceKey:         "default",
		SourceEnvironmentKey: "source",
		SourceNamespaceKey:   "default",
		OnConflict:           rpcenvironments.ConflictStrategy_CONFLICT_STRATEGY_FAIL,
		Revision:             "rev-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "rev-2", resp.Revision)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, rpcenvironments.OperationStatus_OPERATION_STATUS_SUCCESS, resp.Results[0].Status)
}

func TestCopyNamespace_SkipExistingResourceOnConflictSkip(t *testing.T) {
	ctx := t.Context()
	licenseManager := license.NewMockManager(t)
	licenseManager.EXPECT().Product().Return(product.Pro)

	sourceEnv := NewMockEnvironment(t)
	targetEnv := NewMockEnvironment(t)

	sourceEnv.EXPECT().
		View(ctx, NewResourceType("flipt.core", "Flag"), mock.Anything).
		RunAndReturn(func(ctx context.Context, _ ResourceType, fn ViewFunc) error {
			store := newTestResourceStore()
			store.resources[testResourceKey("default", "flag-a")] = &rpcenvironments.Resource{
				NamespaceKey: "default",
				Key:          "flag-a",
				Payload:      &anypb.Any{TypeUrl: "flipt.core.Flag"},
			}
			return fn(ctx, store)
		}).
		Once()
	sourceEnv.EXPECT().
		View(ctx, NewResourceType("flipt.core", "Segment"), mock.Anything).
		RunAndReturn(func(ctx context.Context, _ ResourceType, fn ViewFunc) error {
			return fn(ctx, newTestResourceStore())
		}).
		Once()

	targetEnv.EXPECT().
		GetNamespace(ctx, "default").
		Return(&rpcenvironments.NamespaceResponse{}, nil).
		Once()
	targetEnv.EXPECT().
		View(ctx, NewResourceType("flipt.core", "Flag"), mock.Anything).
		RunAndReturn(func(ctx context.Context, _ ResourceType, fn ViewFunc) error {
			store := newTestResourceStore()
			store.resources[testResourceKey("default", "flag-a")] = &rpcenvironments.Resource{
				NamespaceKey: "default",
				Key:          "flag-a",
				Payload:      &anypb.Any{TypeUrl: "flipt.core.Flag"},
			}
			return fn(ctx, store)
		}).
		Once()

	s := &Server{
		logger:         zap.NewNop(),
		licenseManager: licenseManager,
		envs: &EnvironmentStore{
			byKey: map[string]Environment{
				"source": sourceEnv,
				"target": targetEnv,
			},
		},
	}

	resp, err := s.CopyNamespace(ctx, &rpcenvironments.CopyNamespaceRequest{
		EnvironmentKey:       "target",
		NamespaceKey:         "default",
		SourceEnvironmentKey: "source",
		SourceNamespaceKey:   "default",
		OnConflict:           rpcenvironments.ConflictStrategy_CONFLICT_STRATEGY_SKIP,
		Revision:             "rev-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "rev-1", resp.Revision)
	require.Len(t, resp.Results, 1)
	assert.Equal(t, rpcenvironments.OperationStatus_OPERATION_STATUS_SKIPPED, resp.Results[0].Status)
}
