package environments

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	errs "go.flipt.io/flipt/errors"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
)

func TestApplyCreateWithConflict(t *testing.T) {
	ctx := t.Context()

	t.Run("skip when resource exists and strategy is skip", func(t *testing.T) {
		store := newTestResourceStore()
		store.resources[testResourceKey("default", "flag")] = &rpcenvironments.Resource{
			NamespaceKey: "default",
			Key:          "flag",
		}

		status, err := applyCreateWithConflict(ctx, store, &rpcenvironments.Resource{
			NamespaceKey: "default",
			Key:          "flag",
		}, rpcenvironments.ConflictStrategy_CONFLICT_STRATEGY_SKIP, true)

		require.NoError(t, err)
		assert.Equal(t, rpcenvironments.OperationStatus_OPERATION_STATUS_SKIPPED, status)
	})

	t.Run("fails when resource exists and strategy is fail", func(t *testing.T) {
		store := newTestResourceStore()

		status, err := applyCreateWithConflict(ctx, store, &rpcenvironments.Resource{
			NamespaceKey: "default",
			Key:          "flag",
		}, rpcenvironments.ConflictStrategy_CONFLICT_STRATEGY_FAIL, true)

		require.Error(t, err)
		assert.Equal(t, rpcenvironments.OperationStatus_OPERATION_STATUS_FAILED, status)
		assert.True(t, errs.AsMatch[errs.ErrAlreadyExists](err))
	})

	t.Run("overwrites when resource exists and strategy is overwrite", func(t *testing.T) {
		store := newTestResourceStore()
		store.resources[testResourceKey("default", "flag")] = &rpcenvironments.Resource{
			NamespaceKey: "default",
			Key:          "flag",
		}

		status, err := applyCreateWithConflict(ctx, store, &rpcenvironments.Resource{
			NamespaceKey: "default",
			Key:          "flag",
		}, rpcenvironments.ConflictStrategy_CONFLICT_STRATEGY_OVERWRITE, true)

		require.NoError(t, err)
		assert.Equal(t, rpcenvironments.OperationStatus_OPERATION_STATUS_SUCCESS, status)
		assert.Equal(t, 1, store.updateCalls)
		assert.Equal(t, 0, store.createCalls)
	})

	t.Run("creates when resource does not exist", func(t *testing.T) {
		store := newTestResourceStore()

		status, err := applyCreateWithConflict(ctx, store, &rpcenvironments.Resource{
			NamespaceKey: "default",
			Key:          "flag",
		}, rpcenvironments.ConflictStrategy_CONFLICT_STRATEGY_FAIL, false)

		require.NoError(t, err)
		assert.Equal(t, rpcenvironments.OperationStatus_OPERATION_STATUS_SUCCESS, status)
		assert.Equal(t, 1, store.createCalls)
		assert.Equal(t, 0, store.updateCalls)
	})
}

func TestApplyBulkOperation(t *testing.T) {
	ctx := t.Context()

	t.Run("create with skip returns skipped when target exists", func(t *testing.T) {
		store := newTestResourceStore()
		store.resources[testResourceKey("default", "flag")] = &rpcenvironments.Resource{
			NamespaceKey: "default",
			Key:          "flag",
		}

		status, err := applyBulkOperation(ctx, store, "default", &rpcenvironments.BulkApplyResourcesRequest{
			Operation:  rpcenvironments.BulkOperation_BULK_OPERATION_CREATE,
			Key:        "flag",
			OnConflict: rpcenvironments.ConflictStrategy_CONFLICT_STRATEGY_SKIP,
		})

		require.NoError(t, err)
		assert.Equal(t, rpcenvironments.OperationStatus_OPERATION_STATUS_SKIPPED, status)
	})

	t.Run("upsert surfaces get errors that are not not found", func(t *testing.T) {
		store := newTestResourceStore()
		store.getErrors[testResourceKey("default", "flag")] = fmt.Errorf("read failed")

		status, err := applyBulkOperation(ctx, store, "default", &rpcenvironments.BulkApplyResourcesRequest{
			Operation: rpcenvironments.BulkOperation_BULK_OPERATION_UPSERT,
			Key:       "flag",
		})

		require.Error(t, err)
		assert.Equal(t, rpcenvironments.OperationStatus_OPERATION_STATUS_FAILED, status)
		assert.Contains(t, err.Error(), "read failed")
	})
}

type testResourceStore struct {
	resources   map[string]*rpcenvironments.Resource
	getErrors   map[string]error
	createCalls int
	updateCalls int
}

func newTestResourceStore() *testResourceStore {
	return &testResourceStore{
		resources: make(map[string]*rpcenvironments.Resource),
		getErrors: make(map[string]error),
	}
}

func (s *testResourceStore) GetResource(_ context.Context, namespace, key string) (*rpcenvironments.ResourceResponse, error) {
	compositeKey := testResourceKey(namespace, key)
	if err, ok := s.getErrors[compositeKey]; ok {
		return nil, err
	}

	resource, ok := s.resources[compositeKey]
	if !ok {
		return nil, errs.ErrNotFoundf("resource %q/%q", namespace, key)
	}

	return &rpcenvironments.ResourceResponse{Resource: resource}, nil
}

func (s *testResourceStore) ListResources(_ context.Context, namespace string) (*rpcenvironments.ListResourcesResponse, error) {
	var resources []*rpcenvironments.Resource

	for _, resource := range s.resources {
		if resource.NamespaceKey == namespace {
			resources = append(resources, resource)
		}
	}

	return &rpcenvironments.ListResourcesResponse{Resources: resources}, nil
}

func (s *testResourceStore) CreateResource(_ context.Context, resource *rpcenvironments.Resource) error {
	s.createCalls++
	s.resources[testResourceKey(resource.NamespaceKey, resource.Key)] = resource
	return nil
}

func (s *testResourceStore) UpdateResource(_ context.Context, resource *rpcenvironments.Resource) error {
	s.updateCalls++
	s.resources[testResourceKey(resource.NamespaceKey, resource.Key)] = resource
	return nil
}

func (s *testResourceStore) DeleteResource(_ context.Context, namespace, key string) error {
	delete(s.resources, testResourceKey(namespace, key))
	return nil
}

func testResourceKey(namespace, key string) string {
	return namespace + "/" + key
}
