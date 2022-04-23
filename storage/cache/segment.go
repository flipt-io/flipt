package cache

import (
	"context"
	"fmt"

	flipt "github.com/markphelps/flipt/rpc/flipt"
	"github.com/markphelps/flipt/storage"
)

const segmentCachePrefix = "segment:"

// GetSegment returns the segment from the cache if it exists; otherwise it delegates to the underlying store
// caching the result if no error
func (c *Store) GetSegment(ctx context.Context, k string) (*flipt.Segment, error) {
	key := segmentCachePrefix + k

	// check if segment exists in cache
	data, ok, err := c.cache.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("getting segment from cache: %w", err)
	}

	if ok {
		c.logger.Debugf("cache hit: %q", key)

		segment, ok := data.(*flipt.Segment)
		if !ok {
			// not segment, bad cache
			// TODO: should we just invalidate the cache instead of returning an error?
			return nil, ErrCorrupt
		}

		return segment, nil
	}

	// segment not in cache, delegate to underlying store
	segment, err := c.store.GetSegment(ctx, k)
	if err != nil {
		return segment, err
	}

	if err := c.cache.Set(ctx, key, segment); err != nil {
		return segment, err
	}

	c.logger.Debugf("cache miss; added: %q", key)
	return segment, nil
}

// ListSegments delegates to the underlying store
func (c *Store) ListSegments(ctx context.Context, opts ...storage.QueryOption) ([]*flipt.Segment, error) {
	return c.store.ListSegments(ctx, opts...)
}

// CreateSegment delegates to the underlying store, flushing the cache in the process
func (c *Store) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	c.cache.Flush(ctx)
	c.logger.Debug("flushed cache")
	return c.store.CreateSegment(ctx, r)
}

// UpdateSegment delegates to the underlying store, flushing the cache in the process
func (c *Store) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	c.cache.Flush(ctx)
	c.logger.Debug("flushed cache")
	return c.store.UpdateSegment(ctx, r)
}

// DeleteSegment delegates to the underlying store, flushing the cache in the process
func (c *Store) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	c.cache.Flush(ctx)
	c.logger.Debug("flushed cache")
	return c.store.DeleteSegment(ctx, r)
}

// CreateConstraint delegates to the underlying store, flushing the cache in the process
func (c *Store) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	c.cache.Flush(ctx)
	c.logger.Debug("flushed cache")
	return c.store.CreateConstraint(ctx, r)
}

// UpdateConstraint delegates to the underlying store, flushing the cache in the process
func (c *Store) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	c.cache.Flush(ctx)
	c.logger.Debug("flushed cache")
	return c.store.UpdateConstraint(ctx, r)
}

// DeleteConstraint delegates to the underlying store, flushing the cache in the process
func (c *Store) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	c.cache.Flush(ctx)
	c.logger.Debug("flushed cache")
	return c.store.DeleteConstraint(ctx, r)
}
