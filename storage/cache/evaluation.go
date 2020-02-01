package cache

import (
	"context"

	"github.com/markphelps/flipt/storage"
	"github.com/sirupsen/logrus"
)

const (
	evaluationRulesCachePrefix         = "eval:rules:flag:"
	evaluationDistributionsCachePrefix = "eval:dist:rule:"
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

// GetEvaluationRules returns all rules applicable to the flagKey provided from the cache if they exist; delegating to the underlying store and caching the result if no error
func (e *EvaluationCache) GetEvaluationRules(ctx context.Context, flagKey string) ([]*storage.EvaluationRule, error) {
	key := evaluationRulesCachePrefix + flagKey

	// check if rules exists in cache
	if data, ok := e.cache.Get(key); ok {
		e.logger.Debugf("cache hit: %q", key)
		cacheHitTotal.WithLabelValues("eval_rules", "memory").Inc()

		rules, ok := data.([]*storage.EvaluationRule)
		if !ok {
			// not rules slice, bad cache
			return nil, ErrCacheCorrupt
		}

		return rules, nil
	}

	// else, get them and add to cache
	rules, err := e.store.GetEvaluationRules(ctx, flagKey)
	if err != nil {
		return rules, err
	}

	e.cache.Set(key, rules)
	e.logger.Debugf("cache miss; added: %q", key)
	cacheMissTotal.WithLabelValues("eval_rules", "memory").Inc()

	return rules, nil
}

// GetEvaluationDistributions returns all distributions applicable to the ruleID provided from the cache if they exist; delegating to the underlying store and caching the result if no error
func (e *EvaluationCache) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	key := evaluationDistributionsCachePrefix + ruleID

	// check if distributions exists in cache
	if data, ok := e.cache.Get(key); ok {
		e.logger.Debugf("cache hit: %q", key)
		cacheHitTotal.WithLabelValues("eval_distributions", "memory").Inc()

		distributions, ok := data.([]*storage.EvaluationDistribution)
		if !ok {
			// not distributions slice, bad cache
			return nil, ErrCacheCorrupt
		}

		return distributions, nil
	}

	// else, get them and add to cache
	distributions, err := e.store.GetEvaluationDistributions(ctx, ruleID)
	if err != nil {
		return distributions, err
	}

	e.cache.Set(key, distributions)
	e.logger.Debugf("cache miss; added %q", key)
	cacheMissTotal.WithLabelValues("eval_distributions", "memory").Inc()

	return distributions, nil
}
