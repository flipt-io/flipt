package fs

import (
	"context"
	"sync"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
)

// syncedStore embeds a Snapshot and wraps the Store methods with a read-write mutex
// to synchronize reads with swapping out the Snapshot.
type syncedStore struct {
	*Snapshot

	mu sync.RWMutex
}

func (s *syncedStore) GetFlag(ctx context.Context, namespaceKey string, key string) (*flipt.Flag, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.GetFlag(ctx, namespaceKey, key)
}

func (s *syncedStore) ListFlags(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Flag], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.ListFlags(ctx, namespaceKey, opts...)
}

func (s *syncedStore) CountFlags(ctx context.Context, namespaceKey string) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.CountFlags(ctx, namespaceKey)
}

func (s *syncedStore) GetRule(ctx context.Context, namespaceKey string, id string) (*flipt.Rule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.GetRule(ctx, namespaceKey, id)
}

func (s *syncedStore) ListRules(ctx context.Context, namespaceKey string, flagKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Rule], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.ListRules(ctx, namespaceKey, flagKey, opts...)
}

func (s *syncedStore) CountRules(ctx context.Context, namespaceKey, flagKey string) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.CountRules(ctx, namespaceKey, flagKey)
}

func (s *syncedStore) GetSegment(ctx context.Context, namespaceKey string, key string) (*flipt.Segment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.GetSegment(ctx, namespaceKey, key)
}

func (s *syncedStore) ListSegments(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Segment], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.ListSegments(ctx, namespaceKey, opts...)
}

func (s *syncedStore) CountSegments(ctx context.Context, namespaceKey string) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.CountSegments(ctx, namespaceKey)
}

func (s *syncedStore) GetEvaluationRules(ctx context.Context, namespaceKey string, flagKey string) ([]*storage.EvaluationRule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.GetEvaluationRules(ctx, namespaceKey, flagKey)
}

func (s *syncedStore) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.GetEvaluationDistributions(ctx, ruleID)
}

func (s *syncedStore) GetNamespace(ctx context.Context, key string) (*flipt.Namespace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.GetNamespace(ctx, key)
}

func (s *syncedStore) ListNamespaces(ctx context.Context, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Namespace], error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.ListNamespaces(ctx, opts...)
}

func (s *syncedStore) CountNamespaces(ctx context.Context) (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Snapshot.CountNamespaces(ctx)
}
