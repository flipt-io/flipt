package cache

import (
	"context"

	"github.com/markphelps/flipt/storage"
)

const (
	evaluationRulesCachePrefix         = "eval:rules:flag:"
	evaluationDistributionsCachePrefix = "eval:dist:rule:"
)

// GetEvaluationRules returns all rules applicable to the flagKey provided from the cache if they exist; delegating to the underlying store and caching the result if no error
func (c *Store) GetEvaluationRules(ctx context.Context, flagKey string) ([]*storage.EvaluationRule, error) {
	key := evaluationRulesCachePrefix + flagKey

	// check if rules exists in cache
	if data, ok := c.cache.Get(key); ok {
		c.logger.Debugf("cache hit: %q", key)
		cacheHitTotal.WithLabelValues("eval_rules", "memory").Inc()

		rules, ok := data.([]*storage.EvaluationRule)
		if !ok {
			// not rules slice, bad cache
			return nil, ErrCacheCorrupt
		}

		return rules, nil
	}

	// else, get them and add to cache
	rules, err := c.store.GetEvaluationRules(ctx, flagKey)
	if err != nil {
		return rules, err
	}

	if len(rules) > 0 {
		c.cache.Set(key, rules)
		c.logger.Debugf("cache miss; added: %q", key)
		cacheMissTotal.WithLabelValues("eval_rules", "memory").Inc()
	}

	return rules, nil
}

// GetEvaluationDistributions returns all distributions applicable to the ruleID provided from the cache if they exist; delegating to the underlying store and caching the result if no error
func (c *Store) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	key := evaluationDistributionsCachePrefix + ruleID

	// check if distributions exists in cache
	if data, ok := c.cache.Get(key); ok {
		c.logger.Debugf("cache hit: %q", key)
		cacheHitTotal.WithLabelValues("eval_distributions", "memory").Inc()

		distributions, ok := data.([]*storage.EvaluationDistribution)
		if !ok {
			// not distributions slice, bad cache
			return nil, ErrCacheCorrupt
		}

		return distributions, nil
	}

	// else, get them and add to cache
	distributions, err := c.store.GetEvaluationDistributions(ctx, ruleID)
	if err != nil {
		return distributions, err
	}

	if len(distributions) > 0 {
		c.cache.Set(key, distributions)
		c.logger.Debugf("cache miss; added %q", key)
		cacheMissTotal.WithLabelValues("eval_distributions", "memory").Inc()
	}

	return distributions, nil
}
