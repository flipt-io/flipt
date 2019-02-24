package cache

import (
	"context"
	"errors"
	"testing"

	lru "github.com/hashicorp/golang-lru"
	flipt "github.com/markphelps/flipt/proto"
	"github.com/markphelps/flipt/storage"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetFlag(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &flagStoreMock{
			getFlagFn: func(_ context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{Key: r.Key}, nil
			},
		}
		cache, _ = lru.New(1)
		spy      = &cacherSpy{cache: cache}
		subject  = NewFlagCache(logger, spy, store)
	)

	got, err := subject.GetFlag(context.TODO(), &flipt.GetFlagRequest{Key: "foo"})
	require.NoError(t, err)
	assert.NotNil(t, got)

	// shouldnt exist in the cache so it should be added
	assert.Equal(t, 1, spy.addCalled)
	assert.Equal(t, 1, spy.getCalled)

	got, err = subject.GetFlag(context.TODO(), &flipt.GetFlagRequest{Key: "foo"})
	require.NoError(t, err)
	assert.NotNil(t, got)

	// should already exist in the cache so it should NOT be added
	assert.Equal(t, 1, spy.addCalled)
	assert.Equal(t, 2, spy.getCalled)
}

func TestGetFlagNotFound(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &flagStoreMock{
			getFlagFn: func(context.Context, *flipt.GetFlagRequest) (*flipt.Flag, error) {
				return nil, storage.ErrNotFound("foo")
			},
		}
		cache, _ = lru.New(1)
		spy      = &cacherSpy{cache: cache}
		subject  = NewFlagCache(logger, spy, store)
	)

	_, err := subject.GetFlag(context.TODO(), &flipt.GetFlagRequest{Key: "foo"})
	require.Error(t, err)

	// doesnt exists so it should not be added
	assert.Equal(t, 0, spy.addCalled)
	assert.Equal(t, 1, spy.getCalled)
}

func TestListFlags(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &flagStoreMock{
			listFlagsFn: func(context.Context, *flipt.ListFlagRequest) ([]*flipt.Flag, error) {
				return []*flipt.Flag{
					{Key: "foo"},
					{Key: "bar"},
				}, nil
			},
		}
		cache, _ = lru.New(1)
		spy      = &cacherSpy{cache: cache}
		subject  = NewFlagCache(logger, spy, store)
	)

	got, err := subject.ListFlags(context.TODO(), &flipt.ListFlagRequest{})
	require.NoError(t, err)
	assert.NotEmpty(t, got)
	assert.Len(t, got, 2)

	// doesnt read from the cache
	assert.Equal(t, 0, spy.getCalled)
}

func TestCreateFlag(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &flagStoreMock{
			createFlagFn: func(_ context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{
					Key: r.Key,
				}, nil
			},
		}
		cache, _ = lru.New(1)
		spy      = &cacherSpy{cache: cache}
		subject  = NewFlagCache(logger, spy, store)
	)

	flag, err := subject.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{Key: "foo"})
	require.NoError(t, err)
	assert.NotNil(t, flag)

	// should be added
	assert.Equal(t, 1, spy.addCalled)
}

func TestCreateFlag_Error(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &flagStoreMock{
			createFlagFn: func(context.Context, *flipt.CreateFlagRequest) (*flipt.Flag, error) {
				return nil, errors.New("error")
			},
		}
		cache, _ = lru.New(1)
		spy      = &cacherSpy{cache: cache}
		subject  = NewFlagCache(logger, spy, store)
	)

	_, err := subject.CreateFlag(context.TODO(), &flipt.CreateFlagRequest{Key: "foo"})
	require.Error(t, err)

	// should NOT be added
	assert.Equal(t, 0, spy.addCalled)
}

func TestUpdateFlag(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &flagStoreMock{
			updateFlagFn: func(_ context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
				return &flipt.Flag{Key: r.Key}, nil
			},
		}
		cache, _ = lru.New(1)
		spy      = &cacherSpy{cache: cache}
		subject  = NewFlagCache(logger, spy, store)
	)

	_, err := subject.UpdateFlag(context.TODO(), &flipt.UpdateFlagRequest{Key: "foo"})
	require.NoError(t, err)

	// it should be removed
	assert.Equal(t, 1, spy.removeCalled)
	assert.Equal(t, 0, spy.addCalled)
	assert.Equal(t, 0, spy.getCalled)
}

func TestDeleteFlag(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &flagStoreMock{
			deleteFlagFn: func(_ context.Context, r *flipt.DeleteFlagRequest) error {
				return nil
			},
		}
		cache, _ = lru.New(1)
		spy      = &cacherSpy{cache: cache}
		subject  = NewFlagCache(logger, spy, store)
	)

	err := subject.DeleteFlag(context.TODO(), &flipt.DeleteFlagRequest{Key: "foo"})
	require.NoError(t, err)

	// it should be removed
	assert.Equal(t, 1, spy.removeCalled)
	assert.Equal(t, 0, spy.addCalled)
	assert.Equal(t, 0, spy.getCalled)
}

func TestCreateVariant(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &flagStoreMock{
			createVariantFn: func(_ context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
				return &flipt.Variant{FlagKey: r.FlagKey}, nil
			},
		}
		cache, _ = lru.New(1)
		spy      = &cacherSpy{cache: cache}
		subject  = NewFlagCache(logger, spy, store)
	)

	_, err := subject.CreateVariant(context.TODO(), &flipt.CreateVariantRequest{FlagKey: "foo"})
	require.NoError(t, err)

	// it should be removed
	assert.Equal(t, 1, spy.removeCalled)
	assert.Equal(t, 0, spy.addCalled)
	assert.Equal(t, 0, spy.getCalled)
}

func TestUpdateVariant(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &flagStoreMock{
			updateVariantFn: func(_ context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
				return &flipt.Variant{FlagKey: r.FlagKey}, nil
			},
		}
		cache, _ = lru.New(1)
		spy      = &cacherSpy{cache: cache}
		subject  = NewFlagCache(logger, spy, store)
	)

	_, err := subject.UpdateVariant(context.TODO(), &flipt.UpdateVariantRequest{FlagKey: "foo"})
	require.NoError(t, err)

	// it should be removed
	assert.Equal(t, 1, spy.removeCalled)
	assert.Equal(t, 0, spy.addCalled)
	assert.Equal(t, 0, spy.getCalled)
}

func TestDeleteVariant(t *testing.T) {
	var (
		logger, _ = test.NewNullLogger()
		store     = &flagStoreMock{
			deleteVariantFn: func(context.Context, *flipt.DeleteVariantRequest) error {
				return nil
			},
		}
		cache, _ = lru.New(1)
		spy      = &cacherSpy{cache: cache}
		subject  = NewFlagCache(logger, spy, store)
	)

	err := subject.DeleteVariant(context.TODO(), &flipt.DeleteVariantRequest{FlagKey: "foo"})
	require.NoError(t, err)

	// it should be removed
	assert.Equal(t, 1, spy.removeCalled)
	assert.Equal(t, 0, spy.addCalled)
	assert.Equal(t, 0, spy.getCalled)
}
