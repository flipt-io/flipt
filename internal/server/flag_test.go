//nolint:goconst
package server

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap/zaptest"
)

func TestGetFlag(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.GetFlagRequest{Key: "foo"}
	)

	store.On("GetFlag", mock.Anything, storage.NewResource("", "foo")).Return(&flipt.Flag{
		Key:     req.Key,
		Enabled: true,
	}, nil)

	got, err := s.GetFlag(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
	assert.Equal(t, "foo", got.Key)
	assert.True(t, got.Enabled)
}

func TestListFlags_PaginationOffset(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	defer store.AssertExpectations(t)

	store.On("ListFlags", mock.Anything, storage.ListWithOptions(storage.NewNamespace(""),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](
			storage.WithOffset(10),
		),
	)).Return(
		storage.ResultSet[*flipt.Flag]{
			Results: []*flipt.Flag{
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
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
	)

	defer store.AssertExpectations(t)

	store.On("ListFlags", mock.Anything, storage.ListWithOptions(storage.NewNamespace(""),
		storage.ListWithQueryParamOptions[storage.NamespaceRequest](
			storage.WithOffset(10),
			storage.WithPageToken("Zm9v"),
		),
	)).Return(
		storage.ResultSet[*flipt.Flag]{
			Results: []*flipt.Flag{
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

func TestCreateFlag(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.CreateFlagRequest{
			Key:         "key",
			Name:        "name",
			Description: "desc",
			Enabled:     true,
		}
	)

	store.On("CreateFlag", mock.Anything, req).Return(&flipt.Flag{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
	}, nil)

	got, err := s.CreateFlag(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestUpdateFlag(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.UpdateFlagRequest{
			Key:         "key",
			Name:        "name",
			Description: "desc",
			Enabled:     true,
		}
	)

	store.On("UpdateFlag", mock.Anything, req).Return(&flipt.Flag{
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Enabled:     req.Enabled,
	}, nil)

	got, err := s.UpdateFlag(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestDeleteFlag(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.DeleteFlagRequest{
			Key: "key",
		}
	)

	store.On("DeleteFlag", mock.Anything, req).Return(nil)

	got, err := s.DeleteFlag(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestCreateVariant(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.CreateVariantRequest{
			FlagKey:     "flagKey",
			Key:         "key",
			Name:        "name",
			Description: "desc",
		}
	)

	store.On("CreateVariant", mock.Anything, req).Return(&flipt.Variant{
		Id:          "1",
		FlagKey:     req.FlagKey,
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Attachment:  req.Attachment,
	}, nil)

	got, err := s.CreateVariant(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestUpdateVariant(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.UpdateVariantRequest{
			Id:          "1",
			FlagKey:     "flagKey",
			Key:         "key",
			Name:        "name",
			Description: "desc",
		}
	)

	store.On("UpdateVariant", mock.Anything, req).Return(&flipt.Variant{
		Id:          req.Id,
		FlagKey:     req.FlagKey,
		Key:         req.Key,
		Name:        req.Name,
		Description: req.Description,
		Attachment:  req.Attachment,
	}, nil)

	got, err := s.UpdateVariant(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestDeleteVariant(t *testing.T) {
	var (
		store  = &common.StoreMock{}
		logger = zaptest.NewLogger(t)
		s      = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.DeleteVariantRequest{
			Id: "1",
		}
	)

	store.On("DeleteVariant", mock.Anything, req).Return(nil)

	got, err := s.DeleteVariant(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}
