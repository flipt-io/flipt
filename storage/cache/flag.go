package cache

import (
	"context"

	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
	"github.com/sirupsen/logrus"
)

const flagCachePrefix = "flag:"

var _ storage.FlagStore = &FlagCache{}

// FlagCache wraps a FlagStore and provides caching
type FlagCache struct {
	logger logrus.FieldLogger
	cache  Cacher
	store  storage.FlagStore
}

// NewFlagCache creates a FlagCache by wrapping a storage.FlagStore
func NewFlagCache(logger logrus.FieldLogger, cacher Cacher, store storage.FlagStore) *FlagCache {
	return &FlagCache{
		logger: logger.WithField("cache", "memory"),
		cache:  cacher,
		store:  store,
	}
}

// GetFlag returns the flag from the cache if it exists; otherwise it delegates to the underlying store
// caching the result if no error
func (f *FlagCache) GetFlag(ctx context.Context, k string) (*flipt.Flag, error) {
	key := flagCachePrefix + k

	// check if flag exists in cache
	if data, ok := f.cache.Get(key); ok {
		f.logger.Debugf("cache hit: %q", key)
		cacheHitTotal.WithLabelValues("flag", "memory").Inc()

		flag, ok := data.(*flipt.Flag)
		if !ok {
			// not flag, bad cache
			return nil, ErrCacheCorrupt
		}

		return flag, nil
	}

	// else, get it and add to cache
	flag, err := f.store.GetFlag(ctx, k)
	if err != nil {
		return flag, err
	}

	f.cache.Set(key, flag)
	f.logger.Debugf("cache miss; added: %q", key)
	cacheMissTotal.WithLabelValues("flag", "memory").Inc()

	return flag, nil
}

// ListFlags delegates to the underlying store
func (f *FlagCache) ListFlags(ctx context.Context, opts ...storage.QueryOption) ([]*flipt.Flag, error) {
	return f.store.ListFlags(ctx, opts...)
}

// CreateFlag delegates to the underlying store, flushing the cache in the process
func (f *FlagCache) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	f.cache.Flush()
	f.logger.Debug("flushed cache")
	return f.store.CreateFlag(ctx, r)
}

// UpdateFlag delegates to the underlying store, flushing the cache in the process
func (f *FlagCache) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	f.cache.Flush()
	f.logger.Debug("flushed cache")
	return f.store.UpdateFlag(ctx, r)
}

// DeleteFlag delegates to the underlying store, flushing the cache in the process
func (f *FlagCache) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	f.cache.Flush()
	f.logger.Debug("flushed cache")
	return f.store.DeleteFlag(ctx, r)
}

// CreateVariant delegates to the underlying store, flushing the cache in the process
func (f *FlagCache) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	f.cache.Flush()
	f.logger.Debug("flushed cache")
	return f.store.CreateVariant(ctx, r)
}

// UpdateVariant delegates to the underlying store, flushing the cache in the process
func (f *FlagCache) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	f.cache.Flush()
	f.logger.Debug("flushed cache")
	return f.store.UpdateVariant(ctx, r)
}

// DeleteVariant delegates to the underlying store, flushing the cache in the process
func (f *FlagCache) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	f.cache.Flush()
	f.logger.Debug("flushed cache")
	return f.store.DeleteVariant(ctx, r)
}
