package cache

import (
	"context"

	"github.com/markphelps/flipt/storage"
	"github.com/sirupsen/logrus"
)

const (
	evaluationCachePrefix              = "eval:"
	evaluationRulesCachePrefix         = "r:"
	evaluationDistributionsCachePrefix = "d:"
)

var _ storage.EvaluationStore = &EvaluationCache{}

// EvaluationCache wraps an EvaluationStore and provides caching
type EvaluationCache struct {
	logger logrus.FieldLogger
	cache  Cacher
	store  storage.EvaluationStore
}

// NewEvaluationCache creates an EvaluationCache by wrapping a storage.EvaluationStore
func NewEvaluationCache(logger logrus.FieldLogger, cacher Cacher, store storage.EvaluationStore) *EvaluationCache {
	return &EvaluationCache{
		logger: logger.WithField("cache", "memory"),
		cache:  cacher,
		store:  store,
	}
}

// GetEvaluationRules returns rules applicable to flagKey provided
// Note: Rules MUST be returned in order by Rank
func (e *EvaluationCache) GetEvaluationRules(ctx context.Context, flagKey string) ([]*storage.EvaluationRule, error) {
	panic("not implemented") // TODO: Implement
}

func (e *EvaluationCache) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	panic("not implemented") // TODO: Implement
}

func evalCacheKey(t, k string) string {
	return evaluationCachePrefix + t + k
}
