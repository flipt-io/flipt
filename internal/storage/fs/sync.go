package fs

import (
	"context"
	"sync"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
)

type syncedStore struct {
	*storeSnapshot

	mu sync.RWMutex
}

func (l *syncedStore) GetFlag(ctx context.Context, namespaceKey string, key string) (*flipt.Flag, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.GetFlag(ctx, namespaceKey, key)
}

func (l *syncedStore) ListFlags(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Flag], error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.ListFlags(ctx, namespaceKey, opts...)
}

func (l *syncedStore) CountFlags(ctx context.Context, namespaceKey string) (uint64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.CountFlags(ctx, namespaceKey)
}

func (l *syncedStore) GetRule(ctx context.Context, namespaceKey string, id string) (*flipt.Rule, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.GetRule(ctx, namespaceKey, id)
}

func (l *syncedStore) ListRules(ctx context.Context, namespaceKey string, flagKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Rule], error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.ListRules(ctx, namespaceKey, flagKey, opts...)
}

func (l *syncedStore) CountRules(ctx context.Context, namespaceKey string) (uint64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.CountRules(ctx, namespaceKey)
}

func (l *syncedStore) GetSegment(ctx context.Context, namespaceKey string, key string) (*flipt.Segment, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.GetSegment(ctx, namespaceKey, key)
}

func (l *syncedStore) ListSegments(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Segment], error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.ListSegments(ctx, namespaceKey, opts...)
}

func (l *syncedStore) CountSegments(ctx context.Context, namespaceKey string) (uint64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.CountSegments(ctx, namespaceKey)
}

// GetEvaluationRules returns rules applicable to flagKey provided
// Note: Rules MUST be returned in order by Rank
func (l *syncedStore) GetEvaluationRules(ctx context.Context, namespaceKey string, flagKey string) ([]*storage.EvaluationRule, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.GetEvaluationRules(ctx, namespaceKey, flagKey)
}

func (l *syncedStore) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.GetEvaluationDistributions(ctx, ruleID)
}

func (l *syncedStore) GetNamespace(ctx context.Context, key string) (*flipt.Namespace, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.GetNamespace(ctx, key)
}

func (l *syncedStore) ListNamespaces(ctx context.Context, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Namespace], error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.ListNamespaces(ctx, opts...)
}

func (l *syncedStore) CountNamespaces(ctx context.Context) (uint64, error) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	return l.storeSnapshot.CountNamespaces(ctx)
}
