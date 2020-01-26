package cache

import (
	"context"
	"sync"

	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
)

// cacherSpy is a simple in memory map that acts as a cache
// and records interactions for tests
type cacherSpy struct {
	getCalled, setCalled, deleteCalled, flushCalled int
	cache                                           map[string]interface{}
	mu                                              sync.Mutex
}

func newCacherSpy() *cacherSpy {
	return &cacherSpy{
		cache: make(map[string]interface{}),
	}
}

func (c *cacherSpy) Get(key string) (interface{}, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.getCalled++
	if v, ok := c.cache[key]; ok {
		return v, true
	}

	return nil, false
}

func (c *cacherSpy) Set(key string, value interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.setCalled++
	c.cache[key] = value
}

func (c *cacherSpy) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.deleteCalled++
	delete(c.cache, key)
}

func (c *cacherSpy) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.flushCalled++
	c.cache = make(map[string]interface{})
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
