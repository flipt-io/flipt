package cache

import (
	"context"
	"testing"

	"github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc/flipt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetFlag(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("GetFlag", mock.Anything, mock.Anything).Return(&flipt.Flag{Key: "foo"}, nil)
	cacher.On("Get", mock.Anything, mock.Anything).Return(&flipt.Flag{}, false, nil).Once()
	cacher.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)

	got, err := subject.GetFlag(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// shouldnt exist in the cache so it should be added
	cacher.AssertCalled(t, "Set", mock.Anything, "flag:foo", mock.Anything)
	cacher.AssertCalled(t, "Get", mock.Anything, "flag:foo")

	cacher.On("Get", mock.Anything, mock.Anything).Return(&flipt.Flag{Key: "foo"}, true, nil)

	got, err = subject.GetFlag(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// should already exist in the cache so it should NOT be added
	cacher.AssertNumberOfCalls(t, "Set", 1)
	cacher.AssertNumberOfCalls(t, "Get", 2)
}

func TestGetFlagNotFound(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("GetFlag", mock.Anything, mock.Anything).Return(&flipt.Flag{}, errors.ErrNotFound("foo"))
	cacher.On("Get", mock.Anything, mock.Anything).Return(&flipt.Flag{}, false, nil)

	_, err := subject.GetFlag(context.TODO(), "foo")
	require.Error(t, err)

	// doesnt exists so it should not be added
	cacher.AssertNotCalled(t, "Set")
	cacher.AssertCalled(t, "Get", mock.Anything, "flag:foo")
}

func TestListFlags(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	ret := []*flipt.Flag{
		{Key: "foo"},
		{Key: "bar"},
	}

	store.On("ListFlags", mock.Anything, mock.Anything).Return(ret, nil)

	got, err := subject.ListFlags(context.TODO())
	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.Len(t, got, 2)

	cacher.AssertNotCalled(t, "Set")
	cacher.AssertNotCalled(t, "Get")
}

func TestCreateFlag(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("CreateFlag", mock.Anything, mock.Anything).Return(&flipt.Flag{Key: "foo"}, nil)
	cacher.On("Flush", mock.Anything).Return(nil)

	flag, err := subject.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{Key: "foo"})
	require.NoError(t, err)
	assert.NotNil(t, flag)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush", mock.Anything)
}

func TestUpdateFlag(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("UpdateFlag", mock.Anything, mock.Anything).Return(&flipt.Flag{Key: "foo"}, nil)
	cacher.On("Flush", mock.Anything).Return(nil)

	_, err := subject.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{Key: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush", mock.Anything)
}

func TestDeleteFlag(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("DeleteFlag", mock.Anything, mock.Anything).Return(nil)
	cacher.On("Flush", mock.Anything).Return(nil)

	err := subject.DeleteFlag(context.TODO(), &flipt.DeleteFlagRequest{Key: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush", mock.Anything)
}

func TestCreateVariant(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("CreateVariant", mock.Anything, mock.Anything).Return(&flipt.Variant{FlagKey: "foo"}, nil)
	cacher.On("Flush", mock.Anything).Return(nil)

	_, err := subject.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{FlagKey: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush", mock.Anything)
}

func TestUpdateVariant(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("UpdateVariant", mock.Anything, mock.Anything).Return(&flipt.Variant{FlagKey: "foo"}, nil)
	cacher.On("Flush", mock.Anything).Return(nil)

	_, err := subject.UpdateVariant(context.TODO(), &flipt.UpdateVariantRequest{FlagKey: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush", mock.Anything)
}

func TestDeleteVariant(t *testing.T) {
	var (
		store   = &storeMock{}
		cacher  = &cacherSpy{}
		subject = NewStore(logger, cacher, store)
	)

	store.On("DeleteVariant", mock.Anything, mock.Anything).Return(nil)
	cacher.On("Flush", mock.Anything).Return(nil)

	err := subject.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{FlagKey: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush", mock.Anything)
}
