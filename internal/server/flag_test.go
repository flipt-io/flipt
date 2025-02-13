//nolint:goconst
package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/server/evaluation"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/core"
	"go.uber.org/zap/zaptest"
)

func TestListFlags_PaginationOffset(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		environment = &environments.MockEnvironment{}
		envStore    = &evaluation.MockEnvironmentStore{}
		logger      = zaptest.NewLogger(t)
		s           = &Server{
			logger: logger,
			store:  envStore,
		}
	)

	defer store.AssertExpectations(t)

	envStore.On("Get", mock.Anything, "default").Return(environment, nil)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("ListFlags", mock.Anything, storage.ListWithOptions(storage.NewNamespace(""),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](
			storage.WithOffset(10),
		),
	)).Return(
		storage.ResultSet[*core.Flag]{
			Results: []*core.Flag{
				{
					Key: "foo",
				},
			},
			NextPageToken: "YmFy",
		}, nil)

	store.On("CountFlags", mock.Anything, mock.Anything).Return(uint64(1), nil)

	got, err := s.ListFlags(context.TODO(), &flipt.ListFlagRequest{
		Offset: 10,
	})

	require.NoError(t, err)

	assert.NotEmpty(t, got.Flags)
	assert.Equal(t, "YmFy", got.NextPageToken)
	assert.Equal(t, int32(1), got.TotalCount)
}

func TestListFlags_PaginationPageToken(t *testing.T) {
	var (
		store       = &common.StoreMock{}
		environment = &environments.MockEnvironment{}
		envStore    = &evaluation.MockEnvironmentStore{}
		logger      = zaptest.NewLogger(t)
		s           = &Server{
			logger: logger,
			store:  envStore,
		}
	)

	defer store.AssertExpectations(t)

	envStore.On("Get", mock.Anything, "default").Return(environment, nil)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("ListFlags", mock.Anything, storage.ListWithOptions(storage.NewNamespace(""),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](
			storage.WithOffset(10),
			storage.WithPageToken("Zm9v"),
		),
	)).Return(
		storage.ResultSet[*core.Flag]{
			Results: []*core.Flag{
				{
					Key: "foo",
				},
			},
			NextPageToken: "YmFy",
		}, nil)

	store.On("CountFlags", mock.Anything, mock.Anything).Return(uint64(1), nil)

	got, err := s.ListFlags(context.TODO(), &flipt.ListFlagRequest{
		PageToken: "Zm9v",
		Offset:    10,
	})

	require.NoError(t, err)

	assert.NotEmpty(t, got.Flags)
	assert.Equal(t, "YmFy", got.NextPageToken)
	assert.Equal(t, int32(1), got.TotalCount)
}
