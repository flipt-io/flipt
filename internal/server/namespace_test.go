//nolint:goconst
package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/server/authz"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap/zaptest"
)

func TestGetNamespace(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.GetNamespaceRequest{Key: "foo"}
	)

	store.On("GetNamespace", mock.Anything, storage.NewNamespace("foo")).Return(&flipt.Namespace{
		Key: req.Key,
	}, nil)

	got, err := s.GetNamespace(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
	assert.Equal(t, "foo", got.Key)
}

func TestListNamespaces_PaginationOffset(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	defer store.AssertExpectations(t)

	store.On("ListNamespaces", mock.Anything, storage.ListWithOptions(storage.ReferenceRequest{},
		storage.ListWithQueryParamOptions[storage.ReferenceRequest](
			storage.WithLimit(0),
			storage.WithOffset(10),
		),
	)).Return(
		storage.ResultSet[*flipt.Namespace]{
			Results: []*flipt.Namespace{
				{
					Key: "foo",
				},
			},
			NextPageToken: "YmFy",
		}, nil)

	store.On("CountNamespaces", mock.Anything, storage.ReferenceRequest{}).Return(uint64(1), nil)

	got, err := s.ListNamespaces(context.TODO(), &flipt.ListNamespaceRequest{
		Offset: 10,
	})

	require.NoError(t, err)

	assert.NotEmpty(t, got.Namespaces)
	assert.Equal(t, "YmFy", got.NextPageToken)
	assert.Equal(t, int32(1), got.TotalCount)
}

func TestListNamespaces_WithAuthz(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	defer store.AssertExpectations(t)

	store.On("ListNamespaces", mock.Anything, storage.ListWithOptions(storage.ReferenceRequest{},
		storage.ListWithQueryParamOptions[storage.ReferenceRequest](
			storage.WithLimit(0),
			storage.WithOffset(10),
		),
	)).Return(
		storage.ResultSet[*flipt.Namespace]{
			Results: []*flipt.Namespace{
				{
					Key: "foo",
				},
				{
					Key: "bar",
				},
				{
					Key: "sub",
				},
			},
			NextPageToken: "YmFy",
		}, nil)

	store.On("CountNamespaces", mock.Anything, storage.ReferenceRequest{}).Return(uint64(9), nil)

	tests := []struct {
		name     string
		viewable []string
		expected []string
	}{
		{"empty", []string{}, []string{}},
		{"foo", []string{"foo"}, []string{"foo"}},
		{"foo, bar", []string{"foo", "bar"}, []string{"foo", "bar"}},
		{"foo, bar, ext", []string{"foo", "bar", "ext"}, []string{"foo", "bar"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.TODO(), authz.NamespacesKey, tt.viewable)

			got, err := s.ListNamespaces(ctx, &flipt.ListNamespaceRequest{
				Offset: 10,
			})

			require.NoError(t, err)

			out := make([]string, len(got.Namespaces))
			for i, ns := range got.Namespaces {
				out[i] = ns.Key
			}
			assert.Equal(t, tt.expected, out)
			assert.Equal(t, len(tt.expected), int(got.TotalCount))
		})
	}
}

func TestListNamespaces_PaginationPageToken(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	defer store.AssertExpectations(t)

	store.On("ListNamespaces", mock.Anything, storage.ListWithOptions(storage.ReferenceRequest{},
		storage.ListWithQueryParamOptions[storage.ReferenceRequest](
			storage.WithPageToken("Zm9v"),
			storage.WithOffset(10),
		),
	)).Return(
		storage.ResultSet[*flipt.Namespace]{
			Results: []*flipt.Namespace{
				{
					Key: "foo",
				},
			},
			NextPageToken: "YmFy",
		}, nil)

	store.On("CountNamespaces", mock.Anything, storage.ReferenceRequest{}).Return(uint64(1), nil)

	got, err := s.ListNamespaces(context.TODO(), &flipt.ListNamespaceRequest{
		PageToken: "Zm9v",
		Offset:    10,
	})

	require.NoError(t, err)

	assert.NotEmpty(t, got.Namespaces)
	assert.Equal(t, "YmFy", got.NextPageToken)
	assert.Equal(t, int32(1), got.TotalCount)
}

func TestCreateNamespace(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.CreateNamespaceRequest{
			Key:         "foo",
			Name:        "name",
			Description: "desc",
		}
	)

	store.On("CreateNamespace", mock.Anything, req).Return(&flipt.Namespace{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
	}, nil)

	got, err := s.CreateNamespace(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestUpdateNamespace(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.UpdateNamespaceRequest{
			Key:         "foo",
			Name:        "name",
			Description: "desc",
		}
	)

	store.On("UpdateNamespace", mock.Anything, req).Return(&flipt.Namespace{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
	}, nil)

	got, err := s.UpdateNamespace(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestDeleteNamespace(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.DeleteNamespaceRequest{
			Key: "foo",
		}
	)

	store.On("GetNamespace", mock.Anything, storage.NewNamespace("foo")).Return(&flipt.Namespace{
		Key: req.Key,
	}, nil)

	store.On("CountFlags", mock.Anything, storage.NewNamespace("foo")).Return(uint64(0), nil)

	store.On("DeleteNamespace", mock.Anything, req).Return(nil)

	got, err := s.DeleteNamespace(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestDeleteNamespace_NonExistent(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.DeleteNamespaceRequest{
			Key: "foo",
		}
	)

	var ns *flipt.Namespace
	store.On("GetNamespace", mock.Anything, storage.NewNamespace("foo")).Return(ns, nil) // mock library does not like nil

	store.AssertNotCalled(t, "CountFlags")
	store.AssertNotCalled(t, "DeleteNamespace")

	got, err := s.DeleteNamespace(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestDeleteNamespace_Protected(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.DeleteNamespaceRequest{
			Key: "foo",
		}
	)

	store.On("GetNamespace", mock.Anything, storage.NewNamespace("foo")).Return(&flipt.Namespace{
		Key:       req.Key,
		Protected: true,
	}, nil)

	store.On("CountFlags", mock.Anything, storage.NewNamespace("foo")).Return(uint64(0), nil)

	store.AssertNotCalled(t, "DeleteNamespace")

	got, err := s.DeleteNamespace(context.TODO(), req)
	require.EqualError(t, err, "namespace \"foo\" is protected")
	assert.Nil(t, got)
}

func TestDeleteNamespace_HasFlags(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.DeleteNamespaceRequest{
			Key: "foo",
		}
	)

	store.On("GetNamespace", mock.Anything, storage.NewNamespace("foo")).Return(&flipt.Namespace{
		Key: req.Key,
	}, nil)

	store.On("CountFlags", mock.Anything, storage.NewNamespace("foo")).Return(uint64(1), nil)

	store.AssertNotCalled(t, "DeleteNamespace")

	got, err := s.DeleteNamespace(context.TODO(), req)
	require.EqualError(t, err, "namespace \"foo\" cannot be deleted; flags must be deleted first")
	assert.Nil(t, got)
}

func TestDeleteNamespace_ProtectedWithForce(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.DeleteNamespaceRequest{
			Key:   "foo",
			Force: true,
		}
	)

	store.On("GetNamespace", mock.Anything, storage.NewNamespace("foo")).Return(&flipt.Namespace{
		Key:       req.Key,
		Protected: true,
	}, nil)

	store.AssertNotCalled(t, "CountFlags")

	store.On("DeleteNamespace", mock.Anything, req).Return(nil)

	got, err := s.DeleteNamespace(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestDeleteNamespace_HasFlagsWithForce(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.DeleteNamespaceRequest{
			Key:   "foo",
			Force: true,
		}
	)

	store.On("GetNamespace", mock.Anything, storage.NewNamespace("foo")).Return(&flipt.Namespace{
		Key: req.Key,
	}, nil)

	store.AssertNotCalled(t, "CountFlags")

	store.On("DeleteNamespace", mock.Anything, req).Return(nil)

	got, err := s.DeleteNamespace(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}
