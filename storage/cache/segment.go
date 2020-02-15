package cache

import (
	"context"

	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
	"github.com/sirupsen/logrus"
)

const segmentCachePrefix = "segment:"

var _ storage.SegmentStore = &SegmentCache{}

// SegmentCache wraps a SegmentStore and provides caching
type SegmentCache struct {
	logger logrus.FieldLogger
	cache  Cacher
	store  storage.SegmentStore
}

// NewSegmentCache creates a SegmentCache by wrapping a storage.SegmentStore
func NewSegmentCache(logger logrus.FieldLogger, cacher Cacher, store storage.SegmentStore) *SegmentCache {
	return &SegmentCache{
		logger: logger.WithField("cache", "memory"),
		cache:  cacher,
		store:  store,
	}
}

// GetSegment returns the segment from the cache if it exists; otherwise it delegates to the underlying store
// caching the result if no error
func (s *SegmentCache) GetSegment(ctx context.Context, k string) (*flipt.Segment, error) {
	key := segmentCachePrefix + k

	// check if segment exists in cache
	if data, ok := s.cache.Get(key); ok {
		s.logger.Debugf("cache hit: %q", key)
		cacheHitTotal.WithLabelValues("segment", "memory").Inc()

		segment, ok := data.(*flipt.Segment)
		if !ok {
			// not segment, bad cache
			return nil, ErrCacheCorrupt
		}

		return segment, nil
	}

	// else, get it and add to cache
	segment, err := s.store.GetSegment(ctx, k)
	if err != nil {
		return segment, err
	}

	s.cache.Set(key, segment)
	s.logger.Debugf("cache miss; added: %q", key)
	cacheMissTotal.WithLabelValues("segment", "memory").Inc()

	return segment, nil
}

// ListSegments delegates to the underlying store
func (s *SegmentCache) ListSegments(ctx context.Context, opts ...storage.QueryOption) ([]*flipt.Segment, error) {
	return s.store.ListSegments(ctx, opts...)
}

// CreateSegment delegates to the underlying store, flushing the cache in the process
func (s *SegmentCache) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	s.cache.Flush()
	s.logger.Debug("flushed cache")
	return s.store.CreateSegment(ctx, r)
}

// UpdateSegment delegates to the underlying store, flushing the cache in the process
func (s *SegmentCache) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	s.cache.Flush()
	s.logger.Debug("flushed cache")
	return s.store.UpdateSegment(ctx, r)
}

// DeleteSegment delegates to the underlying store, flushing the cache in the process
func (s *SegmentCache) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	s.cache.Flush()
	s.logger.Debug("flushed cache")
	return s.store.DeleteSegment(ctx, r)
}

// CreateConstraint delegates to the underlying store, flushing the cache in the process
func (s *SegmentCache) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	s.cache.Flush()
	s.logger.Debug("flushed cache")
	return s.store.CreateConstraint(ctx, r)
}

// UpdateConstraint delegates to the underlying store, flushing the cache in the process
func (s *SegmentCache) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	s.cache.Flush()
	s.logger.Debug("flushed cache")
	return s.store.UpdateConstraint(ctx, r)
}

// DeleteConstraint delegates to the underlying store, flushing the cache in the process
func (s *SegmentCache) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	s.cache.Flush()
	s.logger.Debug("flushed cache")
	return s.store.DeleteConstraint(ctx, r)
}
