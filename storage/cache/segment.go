package cache

import (
	"context"

	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
	"github.com/sirupsen/logrus"
)

const (
	segmentCachePrefix = "s:"
	segmentsCacheKey   = "s"
)

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

func (s *SegmentCache) GetSegment(ctx context.Context, r *flipt.GetSegmentRequest) (*flipt.Segment, error) {
	key := segmentCacheKey(r.Key)

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
	segment, err := s.store.GetSegment(ctx, r)
	if err != nil {
		return segment, err
	}

	s.cache.Set(segmentCacheKey(r.Key), segment)
	s.logger.Debugf("cache miss; added: %q", key)
	cacheMissTotal.WithLabelValues("segment", "memory").Inc()

	return segment, nil
}

func (s *SegmentCache) ListSegments(ctx context.Context, r *flipt.ListSegmentRequest) ([]*flipt.Segment, error) {
	// check if segments exists in cache
	if data, ok := s.cache.Get(segmentsCacheKey); ok {
		s.logger.Debug("cache hit: segments")
		cacheHitTotal.WithLabelValues("segments", "memory").Inc()

		segments, ok := data.([]*flipt.Segment)
		if !ok {
			// not flags slice, bad cache
			return nil, ErrCacheCorrupt
		}

		return segments, nil
	}

	// else, get them and add to cache
	segments, err := s.store.ListSegments(ctx, r)
	if err != nil {
		return segments, err
	}

	s.cache.Set(segmentsCacheKey, segments)
	s.logger.Debug("cache miss; added segments")
	cacheMissTotal.WithLabelValues("segments", "memory").Inc()

	return segments, nil
}

func (s *SegmentCache) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	s.cache.Flush()
	s.logger.Debug("flushed cache")
	return s.store.CreateSegment(ctx, r)
}

func (s *SegmentCache) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	s.cache.Flush()
	s.logger.Debug("flushed cache")
	return s.store.UpdateSegment(ctx, r)
}

func (s *SegmentCache) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	s.cache.Flush()
	s.logger.Debug("flushed cache")
	return s.store.DeleteSegment(ctx, r)
}

func (s *SegmentCache) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	s.cache.Flush()
	s.logger.Debug("flushed cache")
	return s.store.CreateConstraint(ctx, r)
}

func (s *SegmentCache) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	s.cache.Flush()
	s.logger.Debug("flushed cache")
	return s.store.UpdateConstraint(ctx, r)
}

func (s *SegmentCache) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	s.cache.Flush()
	s.logger.Debug("flushed cache")
	return s.store.DeleteConstraint(ctx, r)
}

func segmentCacheKey(k string) string {
	return segmentCachePrefix + k
}
