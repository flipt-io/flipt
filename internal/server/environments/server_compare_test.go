package environments

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/coss/license"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestCompareEnvironments_ClassifiesResourceStates(t *testing.T) {
	ctx := t.Context()

	sourceEnv := NewMockEnvironment(t)
	targetEnv := NewMockEnvironment(t)

	sourceEnv.EXPECT().
		View(ctx, NewResourceType("flipt.core", "Flag"), mock.Anything).
		RunAndReturn(func(ctx context.Context, _ ResourceType, fn ViewFunc) error {
			store := newTestResourceStore()
			store.resources[testResourceKey("default", "flag-same")] = &rpcenvironments.Resource{
				NamespaceKey: "default",
				Key:          "flag-same",
				Payload:      &anypb.Any{TypeUrl: "flipt.core.Flag", Value: []byte("same")},
			}
			store.resources[testResourceKey("default", "flag-diff")] = &rpcenvironments.Resource{
				NamespaceKey: "default",
				Key:          "flag-diff",
				Payload:      &anypb.Any{TypeUrl: "flipt.core.Flag", Value: []byte("source")},
			}
			store.resources[testResourceKey("default", "flag-source")] = &rpcenvironments.Resource{
				NamespaceKey: "default",
				Key:          "flag-source",
				Payload:      &anypb.Any{TypeUrl: "flipt.core.Flag", Value: []byte("source-only")},
			}
			return fn(ctx, store)
		}).
		Once()

	targetEnv.EXPECT().
		View(ctx, NewResourceType("flipt.core", "Flag"), mock.Anything).
		RunAndReturn(func(ctx context.Context, _ ResourceType, fn ViewFunc) error {
			store := newTestResourceStore()
			store.resources[testResourceKey("default", "flag-same")] = &rpcenvironments.Resource{
				NamespaceKey: "default",
				Key:          "flag-same",
				Payload:      &anypb.Any{TypeUrl: "flipt.core.Flag", Value: []byte("same")},
			}
			store.resources[testResourceKey("default", "flag-diff")] = &rpcenvironments.Resource{
				NamespaceKey: "default",
				Key:          "flag-diff",
				Payload:      &anypb.Any{TypeUrl: "flipt.core.Flag", Value: []byte("target")},
			}
			store.resources[testResourceKey("default", "flag-target")] = &rpcenvironments.Resource{
				NamespaceKey: "default",
				Key:          "flag-target",
				Payload:      &anypb.Any{TypeUrl: "flipt.core.Flag", Value: []byte("target-only")},
			}
			return fn(ctx, store)
		}).
		Once()

	s := &Server{
		logger:         zap.NewNop(),
		licenseManager: license.NewMockManager(t),
		envs: &EnvironmentStore{
			byKey: map[string]Environment{
				"source": sourceEnv,
				"target": targetEnv,
			},
		},
	}

	resp, err := s.CompareEnvironments(ctx, &rpcenvironments.CompareEnvironmentsRequest{
		EnvironmentKey:       "source",
		NamespaceKey:         "default",
		TargetEnvironmentKey: "target",
		TargetNamespaceKey:   "default",
		TypeUrls:             []string{"flipt.core.Flag"},
	})

	require.NoError(t, err)
	require.Len(t, resp.Results, 4)

	results := map[string]rpcenvironments.CompareStatus{}
	for _, result := range resp.Results {
		results[result.Key] = result.Status
	}

	assert.Equal(t, rpcenvironments.CompareStatus_COMPARE_STATUS_IDENTICAL, results["flag-same"])
	assert.Equal(t, rpcenvironments.CompareStatus_COMPARE_STATUS_DIFFERENT, results["flag-diff"])
	assert.Equal(t, rpcenvironments.CompareStatus_COMPARE_STATUS_SOURCE_ONLY, results["flag-source"])
	assert.Equal(t, rpcenvironments.CompareStatus_COMPARE_STATUS_TARGET_ONLY, results["flag-target"])
}
