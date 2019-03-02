package cache

import (
	"context"

	"github.com/golang/protobuf/proto"
	flipt "github.com/markphelps/flipt/proto"
	"github.com/markphelps/flipt/storage"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const flagCachePrefix = "flag:"

// FlagCache wraps a FlagStore and provides caching
type FlagCache struct {
	logger logrus.FieldLogger
	cache  Cacher
	store  storage.FlagStore
}

// NewFlagCache creates a FlagCache by wrapping a storage.FlagStore
func NewFlagCache(logger logrus.FieldLogger, cacher Cacher, store storage.FlagStore) *FlagCache {
	return &FlagCache{
		logger: logger.WithField("cache", "flag"),
		cache:  cacher,
		store:  store,
	}
}

// GetFlag returns the flag from the cache if it exists; otherwise it delegates to the underlying store
// caching the result if no error
func (f *FlagCache) GetFlag(ctx context.Context, r *flipt.GetFlagRequest) (*flipt.Flag, error) {
	key := flagCacheKey(r.Key)

	// check if flag exists in cache
	if data, ok := f.cache.Get(key); ok {
		f.logger.Debugf("cache hit: %q", key)

		bytes, bok := data.([]byte)
		if !bok {
			// not bytes, bad cache
			return nil, ErrCacheCorrupt
		}

		flag := &flipt.Flag{}

		if err := proto.Unmarshal(bytes, flag); err != nil {
			return nil, errors.Wrap(err, "getting from cache")
		}

		return flag, nil
	}

	// else, get it and add to cache if it exists in the store
	flag, err := f.store.GetFlag(ctx, r)
	if err != nil {
		return flag, err
	}

	data, err := proto.Marshal(flag)
	if err != nil {
		return flag, errors.Wrap(err, "adding to cache")
	}

	_ = f.cache.Add(flagCacheKey(r.Key), data)
	f.logger.Debugf("cache miss; added: %q", key)
	return flag, nil
}

// ListFlags delegates to the underlying store
func (f *FlagCache) ListFlags(ctx context.Context, r *flipt.ListFlagRequest) ([]*flipt.Flag, error) {
	return f.store.ListFlags(ctx, r)
}

// CreateFlag delegates to the underlying store, caching the result if no error
func (f *FlagCache) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	flag, err := f.store.CreateFlag(ctx, r)
	if err != nil {
		return flag, err
	}

	data, err := proto.Marshal(flag)
	if err != nil {
		return flag, errors.Wrap(err, "adding to cache")
	}

	key := flagCacheKey(r.Key)

	f.logger.Debugf("added: %q", key)
	_ = f.cache.Add(key, data)

	return flag, err
}

// UpdateFlag invalidates the cache key and delegates to the underlying store
func (f *FlagCache) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	key := flagCacheKey(r.Key)
	f.cache.Remove(key)
	f.logger.Debugf("removed: %q", key)
	return f.store.UpdateFlag(ctx, r)
}

// DeleteFlag invalidates the cache key and delegates to the underlying store
func (f *FlagCache) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	key := flagCacheKey(r.Key)
	f.cache.Remove(key)
	f.logger.Debugf("removed: %q", key)
	return f.store.DeleteFlag(ctx, r)
}

// CreateVariant invalidates the cache key and delegates to the underlying store
func (f *FlagCache) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	key := flagCacheKey(r.FlagKey)
	f.cache.Remove(key)
	f.logger.Debugf("removed: %q", key)
	return f.store.CreateVariant(ctx, r)
}

// UpdateVariant invalidates the cache key and delegates to the underlying store
func (f *FlagCache) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	key := flagCacheKey(r.FlagKey)
	f.cache.Remove(key)
	f.logger.Debugf("removed: %q", key)
	return f.store.UpdateVariant(ctx, r)
}

// DeleteVariant invalidates the cache key and delegates to the underlying store
func (f *FlagCache) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	key := flagCacheKey(r.FlagKey)
	f.cache.Remove(key)
	f.logger.Debugf("removed: %q", key)
	return f.store.DeleteVariant(ctx, r)
}

func flagCacheKey(k string) string {
	return flagCachePrefix + k
}
