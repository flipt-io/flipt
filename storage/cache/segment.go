package cache

import (
	"context"

	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
)

const segmentCachePrefix = "segment:"

// GetSegment returns the segment from the cache if it exists; otherwise it delegates to the underlying store
// caching the result if no error
func (c *CacheStore) GetSegment(ctx context.Context, k string) (*flipt.Segment, error) {
	key := segmentCachePrefix + k

	// check if segment exists in cache
	if data, ok := c.cache.Get(key); ok {
		c.logger.Debugf("cache hit: %q", key)
		cacheHitTotal.WithLabelValues("segment", "memory").Inc()

		segment, ok := data.(*flipt.Segment)
		if !ok {
			// not segment, bad cache
			return nil, ErrCacheCorrupt
		}

		return segment, nil
	}

	// else, get it and add to cache
	segment, err := c.store.GetSegment(ctx, k)
	if err != nil {
		return segment, err
	}

	c.cache.Set(key, segment)
	c.logger.Debugf("cache miss; added: %q", key)
	cacheMissTotal.WithLabelValues("segment", "memory").Inc()

	return segment, nil
}

// ListSegments delegates to the underlying store
func (c *CacheStore) ListSegments(ctx context.Context, opts ...storage.QueryOption) ([]*flipt.Segment, error) {
	return c.store.ListSegments(ctx, opts...)
}

// CreateSegment delegates to the underlying store, flushing the cache in the process
func (c *CacheStore) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.CreateSegment(ctx, r)
}

// UpdateSegment delegates to the underlying store, flushing the cache in the process
func (c *CacheStore) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.UpdateSegment(ctx, r)
}

// DeleteSegment delegates to the underlying store, flushing the cache in the process
func (c *CacheStore) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.DeleteSegment(ctx, r)
}

// CreateConstraint delegates to the underlying store, flushing the cache in the process
func (c *CacheStore) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.CreateConstraint(ctx, r)
}

// UpdateConstraint delegates to the underlying store, flushing the cache in the process
func (c *CacheStore) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.UpdateConstraint(ctx, r)
}

// DeleteConstraint delegates to the underlying store, flushing the cache in the process
func (c *CacheStore) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.DeleteConstraint(ctx, r)
}
