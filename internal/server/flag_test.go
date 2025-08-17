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

	// The key assertion: verify that the context passed to ListFlags contains the header environment
	store.On("ListFlags", mock.MatchedBy(func(ctx context.Context) bool {
		env, ok := common.FliptEnvironmentFromContext(ctx)
		return ok && env == headerEnvironment // This should be the header environment, not the request environment
	}), storage.ListWithOptions(storage.NewNamespace(namespaceKey))).Return(
		storage.ResultSet[*core.Flag]{
			Results: []*core.Flag{
				{
					Key:         "test-flag",
					Name:        "Test Flag",
					Description: "Test flag description",
				},
			},
		}, nil)

	store.On("CountFlags", mock.MatchedBy(func(ctx context.Context) bool {
		env, ok := common.FliptEnvironmentFromContext(ctx)
		return ok && env == headerEnvironment // This should be the header environment, not the request environment
	}), mock.Anything).Return(uint64(1), nil)

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
	store.AssertExpectations(t)
}

func TestListFlags_WithoutXEnvironmentHeader(t *testing.T) {
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

	// Test when X-Environment header is NOT present, it should use the request environment
	requestEnvironment := "request-environment"
	namespaceKey := "test-namespace"

	// Context without X-Environment header
	ctx := context.Background()

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	// Verify that the context passed to ListFlags contains the request environment
	store.On("ListFlags", mock.MatchedBy(func(ctx context.Context) bool {
		env, ok := common.FliptEnvironmentFromContext(ctx)
		return ok && env == requestEnvironment // This should be the request environment
	}), storage.ListWithOptions(storage.NewNamespace(namespaceKey))).Return(
		storage.ResultSet[*core.Flag]{
			Results: []*core.Flag{
				{
					Key:         "test-flag-2",
					Name:        "Test Flag 2",
					Description: "Test flag description 2",
				},
			},
		}, nil)

	store.On("CountFlags", mock.MatchedBy(func(ctx context.Context) bool {
		env, ok := common.FliptEnvironmentFromContext(ctx)
		return ok && env == requestEnvironment // This should be the request environment
	}), mock.Anything).Return(uint64(1), nil)

	// Create request
	req := &flipt.ListFlagRequest{
		EnvironmentKey: requestEnvironment,
		NamespaceKey:   namespaceKey,
	}

	// Call ListFlags
	result, err := s.ListFlags(ctx, req)
	require.NoError(t, err)

	// Verify the result
	assert.Len(t, result.Flags, 1)
	assert.Equal(t, "test-flag-2", result.Flags[0].Key)
	assert.Equal(t, "Test Flag 2", result.Flags[0].Name)
	assert.Equal(t, int32(1), result.TotalCount)

	store.AssertExpectations(t)
}

func TestListFlags_WithTypes(t *testing.T) {
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

	// Test when X-Environment header is NOT present, it should use the request environment
	requestEnvironment := "request-environment"
	namespaceKey := "test-namespace"

	// Context without X-Environment header
	ctx := context.Background()

	envStore.On("GetFromContext", mock.Anything).Return(environment)
	environment.On("EvaluationStore").Return(store, nil)

	// Verify that the context passed to ListFlags contains the request environment
	store.On("ListFlags", mock.MatchedBy(func(ctx context.Context) bool {
		env, ok := common.FliptEnvironmentFromContext(ctx)
		return ok && env == requestEnvironment // This should be the request environment
	}), storage.ListWithOptions(storage.NewNamespace(namespaceKey))).Return(
		storage.ResultSet[*core.Flag]{
			Results: []*core.Flag{
				{
					Key:     "test-flag-1",
					Name:    "Test Flag 1",
					Type:    core.FlagType_VARIANT_FLAG_TYPE,
					Enabled: false,
				},
				{
					Key:     "test-flag-2",
					Name:    "Test Flag 2",
					Type:    core.FlagType_BOOLEAN_FLAG_TYPE,
					Enabled: true,
				},
			},
		}, nil)

	store.On("CountFlags", mock.MatchedBy(func(ctx context.Context) bool {
		env, ok := common.FliptEnvironmentFromContext(ctx)
		return ok && env == requestEnvironment // This should be the request environment
	}), mock.Anything).Return(uint64(2), nil)

	// Create request
	req := &flipt.ListFlagRequest{
		EnvironmentKey: requestEnvironment,
		NamespaceKey:   namespaceKey,
	}

	// Call ListFlags
	result, err := s.ListFlags(ctx, req)
	require.NoError(t, err)

	flags := result.Flags
	// Verify the result
	assert.Len(t, flags, 2)
	assert.Equal(t, flipt.FlagType_VARIANT_FLAG_TYPE, flags[0].Type)
	assert.False(t, flags[0].Enabled)
	assert.Equal(t, flipt.FlagType_BOOLEAN_FLAG_TYPE, flags[1].Type)
	assert.True(t, flags[1].Enabled)
	assert.Equal(t, int32(2), result.TotalCount)

	store.AssertExpectations(t)
}
