package environments

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	errs "go.flipt.io/flipt/errors"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
)

func TestNamespaceCRUDMethods(t *testing.T) {
	ctx := t.Context()
	env := NewMockEnvironment(t)

	s := &Server{
		logger: zap.NewNop(),
		envs: &EnvironmentStore{
			byKey: map[string]Environment{"env-a": env},
		},
	}

	env.EXPECT().
		GetNamespace(ctx, "default").
		Return(nil, errs.ErrNotFoundf("namespace")).
		Once()
	env.EXPECT().
		CreateNamespace(ctx, "rev-1", mock.MatchedBy(func(ns *rpcenvironments.Namespace) bool {
			return ns.Key == "default"
		})).
		Return("rev-2", nil).
		Once()

	createResp, err := s.CreateNamespace(ctx, &rpcenvironments.UpdateNamespaceRequest{
		EnvironmentKey: "env-a",
		Key:            "default",
		Name:           "Default",
		Revision:       "rev-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "rev-2", createResp.Revision)

	env.EXPECT().
		GetNamespace(ctx, "default").
		Return(&rpcenvironments.NamespaceResponse{}, nil).
		Once()
	env.EXPECT().
		UpdateNamespace(ctx, "rev-2", mock.MatchedBy(func(ns *rpcenvironments.Namespace) bool {
			return ns.Key == "default"
		})).
		Return("rev-3", nil).
		Once()

	updateResp, err := s.UpdateNamespace(ctx, &rpcenvironments.UpdateNamespaceRequest{
		EnvironmentKey: "env-a",
		Key:            "default",
		Name:           "Default",
		Revision:       "rev-2",
	})
	require.NoError(t, err)
	assert.Equal(t, "rev-3", updateResp.Revision)

	env.EXPECT().
		DeleteNamespace(ctx, "rev-3", "default").
		Return("rev-4", nil).
		Once()

	deleteResp, err := s.DeleteNamespace(ctx, &rpcenvironments.DeleteNamespaceRequest{
		EnvironmentKey: "env-a",
		Key:            "default",
		Revision:       "rev-3",
	})
	require.NoError(t, err)
	assert.Equal(t, "rev-4", deleteResp.Revision)
}

func TestResourceCRUDMethods(t *testing.T) {
	ctx := t.Context()
	env := NewMockEnvironment(t)
	typ := NewResourceType("flipt.core", "Flag")

	s := &Server{
		logger: zap.NewNop(),
		envs: &EnvironmentStore{
			byKey: map[string]Environment{"env-a": env},
		},
	}

	env.EXPECT().
		View(ctx, typ, mock.Anything).
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

	getResp, err := s.GetResource(ctx, &rpcenvironments.GetResourceRequest{
		EnvironmentKey: "env-a",
		NamespaceKey:   "default",
		TypeUrl:        "flipt.core.Flag",
		Key:            "flag-a",
	})
	require.NoError(t, err)
	assert.Equal(t, "flag-a", getResp.Resource.Key)

	env.EXPECT().
		View(ctx, typ, mock.Anything).
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

	listResp, err := s.ListResources(ctx, &rpcenvironments.ListResourcesRequest{
		EnvironmentKey: "env-a",
		NamespaceKey:   "default",
		TypeUrl:        "flipt.core.Flag",
	})
	require.NoError(t, err)
	require.Len(t, listResp.Resources, 1)

	env.EXPECT().
		Update(ctx, "rev-1", typ, mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, _ ResourceType, fn UpdateFunc) (string, error) {
			store := newTestResourceStore()
			require.NoError(t, fn(ctx, store))
			assert.Equal(t, 1, store.createCalls)
			return "rev-2", nil
		}).
		Once()

	createResp, err := s.CreateResource(ctx, &rpcenvironments.UpdateResourceRequest{
		EnvironmentKey: "env-a",
		NamespaceKey:   "default",
		Key:            "flag-new",
		Payload:        &anypb.Any{TypeUrl: "flipt.core.Flag"},
		Revision:       "rev-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "rev-2", createResp.Revision)

	env.EXPECT().
		Update(ctx, "rev-2", typ, mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, _ ResourceType, fn UpdateFunc) (string, error) {
			store := newTestResourceStore()
			store.resources[testResourceKey("default", "flag-a")] = &rpcenvironments.Resource{
				NamespaceKey: "default",
				Key:          "flag-a",
				Payload:      &anypb.Any{TypeUrl: "flipt.core.Flag"},
			}
			require.NoError(t, fn(ctx, store))
			assert.Equal(t, 1, store.updateCalls)
			return "rev-3", nil
		}).
		Once()

	updateResp, err := s.UpdateResource(ctx, &rpcenvironments.UpdateResourceRequest{
		EnvironmentKey: "env-a",
		NamespaceKey:   "default",
		Key:            "flag-a",
		Payload:        &anypb.Any{TypeUrl: "flipt.core.Flag"},
		Revision:       "rev-2",
	})
	require.NoError(t, err)
	assert.Equal(t, "rev-3", updateResp.Revision)

	env.EXPECT().
		Update(ctx, "rev-3", typ, mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, _ ResourceType, fn UpdateFunc) (string, error) {
			store := newTestResourceStore()
			store.resources[testResourceKey("default", "flag-a")] = &rpcenvironments.Resource{
				NamespaceKey: "default",
				Key:          "flag-a",
				Payload:      &anypb.Any{TypeUrl: "flipt.core.Flag"},
			}
			require.NoError(t, fn(ctx, store))
			_, exists := store.resources[testResourceKey("default", "flag-a")]
			assert.False(t, exists)
			return "rev-4", nil
		}).
		Once()

	deleteResp, err := s.DeleteResource(ctx, &rpcenvironments.DeleteResourceRequest{
		EnvironmentKey: "env-a",
		NamespaceKey:   "default",
		TypeUrl:        "flipt.core.Flag",
		Key:            "flag-a",
		Revision:       "rev-3",
	})
	require.NoError(t, err)
	assert.Equal(t, "rev-4", deleteResp.Revision)
}

func TestNamespaceAndResourceErrorPaths(t *testing.T) {
	ctx := t.Context()
	env := NewMockEnvironment(t)
	typ := NewResourceType("flipt.core", "Flag")

	s := &Server{
		logger: zap.NewNop(),
		envs: &EnvironmentStore{
			byKey: map[string]Environment{"env-a": env},
		},
	}

	env.EXPECT().
		GetNamespace(ctx, "default").
		Return(&rpcenvironments.NamespaceResponse{}, nil).
		Once()

	_, err := s.CreateNamespace(ctx, &rpcenvironments.UpdateNamespaceRequest{
		EnvironmentKey: "env-a",
		Key:            "default",
		Name:           "Default",
		Revision:       "rev-1",
	})
	require.Error(t, err)
	assert.True(t, errs.AsMatch[errs.ErrAlreadyExists](err))

	env.EXPECT().
		GetNamespace(ctx, "missing").
		Return(nil, errs.ErrNotFoundf("namespace")).
		Once()

	_, err = s.UpdateNamespace(ctx, &rpcenvironments.UpdateNamespaceRequest{
		EnvironmentKey: "env-a",
		Key:            "missing",
		Name:           "Missing",
		Revision:       "rev-1",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update namespace")

	env.EXPECT().
		Update(ctx, "rev-1", typ, mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, _ ResourceType, fn UpdateFunc) (string, error) {
			store := newTestResourceStore()
			store.resources[testResourceKey("default", "flag-a")] = &rpcenvironments.Resource{
				NamespaceKey: "default",
				Key:          "flag-a",
				Payload:      &anypb.Any{TypeUrl: "flipt.core.Flag"},
			}
			return "", fn(ctx, store)
		}).
		Once()

	_, err = s.CreateResource(ctx, &rpcenvironments.UpdateResourceRequest{
		EnvironmentKey: "env-a",
		NamespaceKey:   "default",
		Key:            "flag-a",
		Payload:        &anypb.Any{TypeUrl: "flipt.core.Flag"},
		Revision:       "rev-1",
	})
	require.Error(t, err)
	assert.True(t, errs.AsMatch[errs.ErrAlreadyExists](err))

	env.EXPECT().
		Update(ctx, "rev-2", typ, mock.Anything).
		RunAndReturn(func(ctx context.Context, _ string, _ ResourceType, fn UpdateFunc) (string, error) {
			store := newTestResourceStore()
			return "", fn(ctx, store)
		}).
		Once()

	_, err = s.UpdateResource(ctx, &rpcenvironments.UpdateResourceRequest{
		EnvironmentKey: "env-a",
		NamespaceKey:   "default",
		Key:            "missing",
		Payload:        &anypb.Any{TypeUrl: "flipt.core.Flag"},
		Revision:       "rev-2",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update resource")

	env.EXPECT().
		Update(ctx, "rev-3", typ, mock.Anything).
		Return("", errs.ErrInvalid("bad revision")).
		Once()

	_, err = s.DeleteResource(ctx, &rpcenvironments.DeleteResourceRequest{
		EnvironmentKey: "env-a",
		NamespaceKey:   "default",
		TypeUrl:        "flipt.core.Flag",
		Key:            "flag-a",
		Revision:       "rev-3",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "bad revision")
}
