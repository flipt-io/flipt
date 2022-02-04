package server

import (
	"context"
	"testing"

	flipt "github.com/markphelps/flipt/rpc/flipt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetFlag(t *testing.T) {
	var (
		store = &storeMock{}
		s     = &Server{
			logger: logger,
			store:  store,
		}
		req = &flipt.GetFlagRequest{Key: "foo"}
	)

	store.On("GetFlag", mock.Anything, "foo").Return(&flipt.Flag{
		Key: req.Key,
	}, nil)

	got, err := s.GetFlag(context.TODO(), req)
	require.NoError(t, err)

	assert.NotNil(t, got)
}

func TestListFlags(t *testing.T) {
	var (
		store = &storeMock{}
		s     = &Server{
			logger: logger,
			store:  store,
		}
	)

	store.On("ListFlags", mock.Anything, mock.Anything).Return(
		[]*flipt.Flag{
			{
				Key: "foo",
			},
		}, nil)

	got, err := s.ListFlags(context.TODO(), &flipt.ListFlagRequest{})
	require.NoError(t, err)

	assert.NotEmpty(t, got.Flags)
}

func TestCreateFlag(t *testing.T) {
	var (
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
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
		store = &storeMock{}
		s     = &Server{
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
