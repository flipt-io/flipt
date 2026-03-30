package environments

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/server/authz"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
)

func TestListEnvironmentBranchesAndChanges(t *testing.T) {
	ctx := t.Context()

	env := NewMockEnvironment(t)
	branchEnv := NewMockEnvironment(t)

	env.EXPECT().
		ListBranches(ctx).
		Return(&rpcenvironments.ListEnvironmentBranchesResponse{
			Branches: []*rpcenvironments.BranchEnvironment{
				{EnvironmentKey: "env-a", Key: "feature-1"},
			},
		}, nil).
		Once()

	env.EXPECT().
		Branch(ctx, "feature-1").
		Return(branchEnv, nil).
		Twice()

	env.EXPECT().
		ListBranchedChanges(ctx, branchEnv).
		Return(&rpcenvironments.ListBranchedEnvironmentChangesResponse{
			Changes: []*rpcenvironments.Change{{Revision: "abc123"}},
		}, nil).
		Once()

	env.EXPECT().
		Propose(ctx, branchEnv, ProposalOptions{
			Title: "t",
			Body:  "b",
			Draft: true,
		}).
		Return(&rpcenvironments.EnvironmentProposalDetails{
			Url: "http://example.local",
		}, nil).
		Once()

	s := &Server{
		logger: zap.NewNop(),
		envs: &EnvironmentStore{
			byKey: map[string]Environment{
				"env-a": env,
			},
		},
	}

	branches, err := s.ListEnvironmentBranches(ctx, &rpcenvironments.ListEnvironmentBranchesRequest{
		EnvironmentKey: "env-a",
	})
	require.NoError(t, err)
	require.Len(t, branches.Branches, 1)

	changes, err := s.ListBranchedEnvironmentChanges(ctx, &rpcenvironments.ListBranchedEnvironmentChangesRequest{
		EnvironmentKey: "env-a",
		Key:            "feature-1",
	})
	require.NoError(t, err)
	require.Len(t, changes.Changes, 1)

	proposal, err := s.ProposeEnvironment(ctx, &rpcenvironments.ProposeEnvironmentRequest{
		EnvironmentKey: "env-a",
		Key:            "feature-1",
		Title:          ptr("t"),
		Body:           ptr("b"),
		Draft:          ptr(true),
	})
	require.NoError(t, err)
	assert.Equal(t, "http://example.local", proposal.Url)
}

func TestListEnvironmentsAndBranchLifecycle(t *testing.T) {
	ctx := t.Context()

	baseEnv := NewMockEnvironment(t)
	otherEnv := NewMockEnvironment(t)
	branchEnv := NewMockEnvironment(t)

	baseEnv.EXPECT().Key().Return("env-a").Maybe()
	baseEnv.EXPECT().Default().Return(true).Maybe()
	baseEnv.EXPECT().Configuration().Return(nil).Maybe()
	baseEnv.EXPECT().Branch(ctx, "feature-1").Return(branchEnv, nil).Once()
	baseEnv.EXPECT().DeleteBranch(ctx, "feature-1").Return(nil).Once()

	otherEnv.EXPECT().Key().Return("env-b").Maybe()
	otherEnv.EXPECT().Default().Return(false).Maybe()
	otherEnv.EXPECT().Configuration().Return(nil).Maybe()

	branchEnv.EXPECT().Key().Return("feature-1").Maybe()
	branchEnv.EXPECT().Default().Return(false).Maybe()
	branchEnv.EXPECT().Configuration().Return(nil).Maybe()

	store := &EnvironmentStore{
		byKey: map[string]Environment{
			"env-a": baseEnv,
			"env-b": otherEnv,
		},
		defaultEnv: baseEnv,
	}

	s := &Server{
		logger: zap.NewNop(),
		envs:   store,
	}

	filteredCtx := contextWithEnvironments(ctx, []string{"env-a"})
	filtered, err := s.ListEnvironments(filteredCtx, &rpcenvironments.ListEnvironmentsRequest{})
	require.NoError(t, err)
	require.Len(t, filtered.Environments, 1)
	assert.Equal(t, "env-a", filtered.Environments[0].Key)

	allCtx := contextWithEnvironments(ctx, []string{"*"})
	all, err := s.ListEnvironments(allCtx, &rpcenvironments.ListEnvironmentsRequest{})
	require.NoError(t, err)
	assert.Len(t, all.Environments, 2)

	branchResp, err := s.BranchEnvironment(ctx, &rpcenvironments.BranchEnvironmentRequest{
		EnvironmentKey: "env-a",
		Key:            "feature-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "feature-1", branchResp.Key)

	_, err = s.DeleteBranchEnvironment(ctx, &rpcenvironments.DeleteBranchEnvironmentRequest{
		EnvironmentKey: "env-a",
		Key:            "feature-1",
	})
	require.NoError(t, err)
}

func TestListNamespaces_RespectsAuthFilter(t *testing.T) {
	ctx := t.Context()
	env := NewMockEnvironment(t)

	env.EXPECT().
		ListNamespaces(mock.Anything).
		RunAndReturn(func(context.Context) (*rpcenvironments.ListNamespacesResponse, error) {
			return &rpcenvironments.ListNamespacesResponse{
				Items: []*rpcenvironments.Namespace{
					{Key: "ns-a"},
					{Key: "ns-b"},
				},
			}, nil
		}).
		Twice()

	s := &Server{
		logger: zap.NewNop(),
		envs: &EnvironmentStore{
			byKey: map[string]Environment{
				"env-a": env,
			},
		},
	}

	filteredCtx := contextWithNamespaces(ctx, []string{"ns-a"})
	filtered, err := s.ListNamespaces(filteredCtx, &rpcenvironments.ListNamespacesRequest{
		EnvironmentKey: "env-a",
	})
	require.NoError(t, err)
	require.Len(t, filtered.Items, 1)
	assert.Equal(t, "ns-a", filtered.Items[0].Key)

	allCtx := contextWithNamespaces(ctx, []string{"*"})
	unfiltered, err := s.ListNamespaces(allCtx, &rpcenvironments.ListNamespacesRequest{
		EnvironmentKey: "env-a",
	})
	require.NoError(t, err)
	require.Len(t, unfiltered.Items, 2)
}

func contextWithNamespaces(ctx context.Context, namespaces []string) context.Context {
	return context.WithValue(ctx, authz.NamespacesKey, namespaces)
}

func contextWithEnvironments(ctx context.Context, environments []string) context.Context {
	return context.WithValue(ctx, authz.EnvironmentsKey, environments)
}

func ptr[T any](v T) *T {
	return &v
}
