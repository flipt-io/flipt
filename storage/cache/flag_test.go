package cache

import (
	"context"
	"testing"

	"github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetFlag(t *testing.T) {
	var (
		store   = &flagStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewFlagCache(logger, cacher, store)
	)

	store.On("GetFlag", mock.Anything, mock.Anything).Return(&flipt.Flag{Key: "foo"}, nil)
	cacher.On("Get", mock.Anything).Return(&flipt.Flag{}, false).Once()
	cacher.On("Set", mock.Anything, mock.Anything)

	got, err := subject.GetFlag(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// shouldnt exist in the cache so it should be added
	cacher.AssertCalled(t, "Set", "flag:foo", mock.Anything)
	cacher.AssertCalled(t, "Get", "flag:foo")

	cacher.On("Get", mock.Anything).Return(&flipt.Flag{Key: "foo"}, true)

	got, err = subject.GetFlag(context.TODO(), "foo")
	require.NoError(t, err)
	assert.NotNil(t, got)

	// should already exist in the cache so it should NOT be added
	cacher.AssertNumberOfCalls(t, "Set", 1)
	cacher.AssertNumberOfCalls(t, "Get", 2)
}

func TestGetFlagNotFound(t *testing.T) {
	var (
		store   = &flagStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewFlagCache(logger, cacher, store)
	)

	store.On("GetFlag", mock.Anything, mock.Anything).Return(&flipt.Flag{}, errors.ErrNotFound("foo"))
	cacher.On("Get", mock.Anything).Return(&flipt.Flag{}, false).Once()

	_, err := subject.GetFlag(context.TODO(), "foo")
	require.Error(t, err)

	// doesnt exists so it should not be added
	cacher.AssertNotCalled(t, "Set")
	cacher.AssertCalled(t, "Get", "flag:foo")
}

func TestListFlags(t *testing.T) {
	var (
		store   = &flagStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewFlagCache(logger, cacher, store)
	)

	ret := []*flipt.Flag{
		{Key: "foo"},
		{Key: "bar"},
	}

	store.On("ListFlags", mock.Anything, uint64(0), uint64(0)).Return(ret, nil)

	got, err := subject.ListFlags(context.TODO(), 0, 0)
	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.Len(t, got, 2)

	cacher.AssertNotCalled(t, "Set")
	cacher.AssertNotCalled(t, "Get")
}

func TestCreateFlag(t *testing.T) {
	var (
		store   = &flagStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewFlagCache(logger, cacher, store)
	)

	store.On("CreateFlag", mock.Anything, mock.Anything).Return(&flipt.Flag{Key: "foo"}, nil)
	cacher.On("Flush")

	flag, err := subject.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{Key: "foo"})
	require.NoError(t, err)
	assert.NotNil(t, flag)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}

func TestUpdateFlag(t *testing.T) {
	var (
		store   = &flagStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewFlagCache(logger, cacher, store)
	)

	store.On("UpdateFlag", mock.Anything, mock.Anything).Return(&flipt.Flag{Key: "foo"}, nil)
	cacher.On("Flush")

	_, err := subject.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{Key: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}

func TestDeleteFlag(t *testing.T) {
	var (
		store   = &flagStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewFlagCache(logger, cacher, store)
	)

	store.On("DeleteFlag", mock.Anything, mock.Anything).Return(nil)
	cacher.On("Flush")

	err := subject.DeleteFlag(context.TODO(), &flipt.DeleteFlagRequest{Key: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}

func TestCreateVariant(t *testing.T) {
	var (
		store   = &flagStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewFlagCache(logger, cacher, store)
	)

	store.On("CreateVariant", mock.Anything, mock.Anything).Return(&flipt.Variant{FlagKey: "foo"}, nil)
	cacher.On("Flush")

	_, err := subject.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{FlagKey: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}

func TestUpdateVariant(t *testing.T) {
	var (
		store   = &flagStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewFlagCache(logger, cacher, store)
	)

	store.On("UpdateVariant", mock.Anything, mock.Anything).Return(&flipt.Variant{FlagKey: "foo"}, nil)
	cacher.On("Flush")

	_, err := subject.UpdateVariant(context.TODO(), &flipt.UpdateVariantRequest{FlagKey: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}

func TestDeleteVariant(t *testing.T) {
	var (
		store   = &flagStoreMock{}
		cacher  = &cacherSpy{}
		subject = NewFlagCache(logger, cacher, store)
	)

	store.On("DeleteVariant", mock.Anything, mock.Anything).Return(nil)
	cacher.On("Flush")

	err := subject.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{FlagKey: "foo"})
	require.NoError(t, err)

	// should not be added
	cacher.AssertNotCalled(t, "Set")
	// should flush cache
	cacher.AssertCalled(t, "Flush")
}
