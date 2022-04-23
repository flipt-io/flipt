package cache

import (
	"context"
	"errors"

	"github.com/markphelps/flipt/storage"
)

const (
	evaluationRulesCachePrefix         = "eval:rules:flag:"
	evaluationDistributionsCachePrefix = "eval:dist:rule:"
)

// GetEvaluationRules returns all rules applicable to the flagKey provided from the cache if they exist; delegating to the underlying store and caching the result if no error
func (c *Store) GetEvaluationRules(ctx context.Context, flagKey string) ([]*storage.EvaluationRule, error) {
	var (
		key   = evaluationRulesCachePrefix + flagKey
		label = c.cache.String()
	)

	// check if rules exists in cache
	data, err := c.cache.Get(ctx, key)
	if err == nil {
		c.logger.Debugf("cache hit: %q", key)
		cacheHitTotal.WithLabelValues(label).Inc()

		rules, ok := data.([]*storage.EvaluationRule)
		if !ok {
			// not rules slice, bad cache
			// TODO: should we just invalidate the cache instead of returning an error?
			return nil, ErrCorrupt
		}

		return rules, nil
	}

	if !errors.Is(err, ErrMiss) {
		// some other error
		return nil, err
	}

	// rules not in cache, delegate to underlying store
	rules, err := c.store.GetEvaluationRules(ctx, flagKey)
	if err != nil {
		return rules, err
	}

	if len(rules) > 0 {
		if err := c.cache.Set(ctx, key, rules); err != nil {
			return rules, err
		}

		c.logger.Debugf("cache miss; added: %q", key)
		cacheMissTotal.WithLabelValues(label).Inc()
	}

	return rules, nil
}

// GetEvaluationDistributions returns all distributions applicable to the ruleID provided from the cache if they exist; delegating to the underlying store and caching the result if no error
func (c *Store) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	var (
		key   = evaluationDistributionsCachePrefix + ruleID
		label = c.cache.String()
	)

	// check if distributions exists in cache
	data, err := c.cache.Get(ctx, key)
	if err == nil {
		c.logger.Debugf("cache hit: %q", key)
		cacheHitTotal.WithLabelValues(label).Inc()

		distributions, ok := data.([]*storage.EvaluationDistribution)
		if !ok {
			// not distributions slice, bad cache
			// TODO: should we just invalidate the cache instead of returning an error?

			return nil, ErrCorrupt
		}

		return distributions, nil
	}

	if !errors.Is(err, ErrMiss) {
		// some other error
		return nil, err
	}

	// distributions not in cache, delegate to underlying store
	distributions, err := c.store.GetEvaluationDistributions(ctx, ruleID)
	if err != nil {
		return distributions, err
	}

	if len(distributions) > 0 {
		if err := c.cache.Set(ctx, key, distributions); err != nil {
			return distributions, err
		}

		c.logger.Debugf("cache miss; added %q", key)
		cacheMissTotal.WithLabelValues(label).Inc()
	}

	return distributions, nil
}
