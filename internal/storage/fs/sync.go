package fs

import (
	"context"
	"sync"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
)

var _ storage.Store = (*syncedStore)(nil)

// syncedStore embeds a storeSnapshot and wraps the Store methods with a read-write mutex
// to synchronize reads with swapping out the storeSnapshot.
type syncedStore struct {
	storage.Store

	mu sync.RWMutex
}

func (s *syncedStore) GetFlag(ctx context.Context, namespaceKey string, key string) (*flipt.Flag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return s.Store.GetFlag(ctx, namespaceKey, key)
}

func (s *syncedStore) ListFlags(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Flag], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return s.Store.ListFlags(ctx, namespaceKey, opts...)
}

func (s *syncedStore) CountFlags(ctx context.Context, namespaceKey string) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return s.Store.CountFlags(ctx, namespaceKey)
}

func (s *syncedStore) GetRule(ctx context.Context, namespaceKey string, id string) (*flipt.Rule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return s.Store.GetRule(ctx, namespaceKey, id)
}

func (s *syncedStore) ListRules(ctx context.Context, namespaceKey string, flagKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Rule], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return s.Store.ListRules(ctx, namespaceKey, flagKey, opts...)
}

func (s *syncedStore) CountRules(ctx context.Context, namespaceKey, flagKey string) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return s.Store.CountRules(ctx, namespaceKey, flagKey)
}

func (s *syncedStore) GetSegment(ctx context.Context, namespaceKey string, key string) (*flipt.Segment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return s.Store.GetSegment(ctx, namespaceKey, key)
}

func (s *syncedStore) ListSegments(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Segment], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return s.Store.ListSegments(ctx, namespaceKey, opts...)
}

func (s *syncedStore) CountSegments(ctx context.Context, namespaceKey string) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return s.Store.CountSegments(ctx, namespaceKey)
}

func (s *syncedStore) GetEvaluationRules(ctx context.Context, namespaceKey string, flagKey string) ([]*storage.EvaluationRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return s.Store.GetEvaluationRules(ctx, namespaceKey, flagKey)
}

func (s *syncedStore) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Store.GetEvaluationDistributions(ctx, ruleID)
}

func (s *syncedStore) GetEvaluationRollouts(ctx context.Context, namespaceKey, flagKey string) ([]*storage.EvaluationRollout, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return s.Store.GetEvaluationRollouts(ctx, namespaceKey, flagKey)
}

func (s *syncedStore) GetNamespace(ctx context.Context, key string) (*flipt.Namespace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if key == "" {
		key = flipt.DefaultNamespace
	}

	return s.Store.GetNamespace(ctx, key)
}

func (s *syncedStore) ListNamespaces(ctx context.Context, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Namespace], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Store.ListNamespaces(ctx, opts...)
}

func (s *syncedStore) CountNamespaces(ctx context.Context) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Store.CountNamespaces(ctx)
}

func (s *syncedStore) GetRollout(ctx context.Context, namespaceKey, id string) (*flipt.Rollout, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return s.Store.GetRollout(ctx, namespaceKey, id)
}

func (s *syncedStore) ListRollouts(ctx context.Context, namespaceKey, flagKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Rollout], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return s.Store.ListRollouts(ctx, namespaceKey, flagKey, opts...)
}

func (s *syncedStore) CountRollouts(ctx context.Context, namespaceKey, flagKey string) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return s.Store.CountRollouts(ctx, namespaceKey, flagKey)
}

func (s *syncedStore) CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (*flipt.Rollout, error) {
	return nil, ErrNotImplemented
}

func (s *syncedStore) UpdateRollout(ctx context.Context, r *flipt.UpdateRolloutRequest) (*flipt.Rollout, error) {
	return nil, ErrNotImplemented
}

func (s *syncedStore) DeleteRollout(ctx context.Context, r *flipt.DeleteRolloutRequest) error {
	return ErrNotImplemented
}

func (s *syncedStore) OrderRollouts(ctx context.Context, r *flipt.OrderRolloutsRequest) error {
	return ErrNotImplemented
}
