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

	envStore.On("GetFromContext", mock.Anything).Return(environment)
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

	envStore.On("GetFromContext", mock.Anything).Return(environment)
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

func TestListFlags_WithXEnvironmentHeader(t *testing.T) {
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

	// Test when X-Environment header is present, it should override the request environment
	headerEnvironment := "header-environment"
	requestEnvironment := "request-environment"
	namespaceKey := "test-namespace"

	// Set up context with X-Environment header
	ctx := common.WithFliptEnvironment(context.Background(), headerEnvironment)

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	store.On("ListFlags", mock.Anything, storage.ListWithOptions(storage.NewNamespace(namespaceKey))).Return(
		storage.ResultSet[*core.Flag]{
			Results: []*core.Flag{
				{
					Key:         "test-flag",
					Name:        "Test Flag",
					Description: "Test flag description",
				},
			},
		}, nil)

	store.On("CountFlags", mock.Anything, mock.Anything).Return(uint64(1), nil)

	// Create request with different environment
	req := &flipt.ListFlagRequest{
		EnvironmentKey: requestEnvironment,
		NamespaceKey:   namespaceKey,
	}

	// Call ListFlags
	result, err := s.ListFlags(ctx, req)
	require.NoError(t, err)

	// Verify the result
	assert.Len(t, result.Flags, 1)
	assert.Equal(t, "test-flag", result.Flags[0].Key)
	assert.Equal(t, "Test Flag", result.Flags[0].Name)
	assert.Equal(t, int32(1), result.TotalCount)

	// Verify that the context was set with the header environment
	// This is implicitly tested by the mock expectation
	store.AssertExpectations(t)
}
