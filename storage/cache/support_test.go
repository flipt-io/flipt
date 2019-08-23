package cache

import (
	"context"

	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
)

type cacherSpy struct {
	addCalled, removeCalled, getCalled int
	cache                              Cacher
}

func (c *cacherSpy) Get(key interface{}) (interface{}, bool) {
	c.getCalled++
	return c.cache.Get(key)
}

func (c *cacherSpy) Add(key interface{}, value interface{}) bool {
	c.addCalled++
	return c.cache.Add(key, value)
}

func (c *cacherSpy) Remove(key interface{}) bool {
	c.removeCalled++
	return c.cache.Remove(key)
}

var _ storage.FlagStore = &flagStoreMock{}

type flagStoreMock struct {
	getFlagFn       func(context.Context, *flipt.GetFlagRequest) (*flipt.Flag, error)
	listFlagsFn     func(context.Context, *flipt.ListFlagRequest) ([]*flipt.Flag, error)
	createFlagFn    func(context.Context, *flipt.CreateFlagRequest) (*flipt.Flag, error)
	updateFlagFn    func(context.Context, *flipt.UpdateFlagRequest) (*flipt.Flag, error)
	deleteFlagFn    func(context.Context, *flipt.DeleteFlagRequest) error
	createVariantFn func(context.Context, *flipt.CreateVariantRequest) (*flipt.Variant, error)
	updateVariantFn func(context.Context, *flipt.UpdateVariantRequest) (*flipt.Variant, error)
	deleteVariantFn func(context.Context, *flipt.DeleteVariantRequest) error
}

func (m *flagStoreMock) GetFlag(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
	return m.getFlagFn(ctx, r)
}

func (m *flagStoreMock) ListFlags(ctx context.Context, r *flipt.ListFlagRequest) ([]*flipt.Flag, error) {
	return m.listFlagsFn(ctx, r)
}

func (m *flagStoreMock) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	return m.createFlagFn(ctx, r)
}

func (m *flagStoreMock) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	return m.updateFlagFn(ctx, r)
}

func (m *flagStoreMock) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	return m.deleteFlagFn(ctx, r)
}

func (m *flagStoreMock) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	return m.createVariantFn(ctx, r)
}

func (m *flagStoreMock) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	return m.updateVariantFn(ctx, r)
}

func (m *flagStoreMock) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	return m.deleteVariantFn(ctx, r)
}
