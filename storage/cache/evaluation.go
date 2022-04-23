package cache

import (
	"context"
	"fmt"

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
	data, ok, err := c.cache.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("getting rules from cache: %w", err)
	}

	if ok {
		c.logger.Debugf("cache hit: %q", key)

		rules, ok := data.([]*storage.EvaluationRule)
		if !ok {
			// not rules slice, bad cache
			// TODO: should we just invalidate the cache instead of returning an error?
			return nil, ErrCorrupt
		}

		return rules, nil
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
	}

	return rules, nil
}

// GetEvaluationDistributions returns all distributions applicable to the ruleID provided from the cache if they exist; delegating to the underlying store and caching the result if no error
func (c *Store) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	key := evaluationDistributionsCachePrefix + ruleID

	// check if distributions exists in cache
	data, ok, err := c.cache.Get(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("getting distributions from cache: %w", err)
	}

	if ok {
		c.logger.Debugf("cache hit: %q", key)

		distributions, ok := data.([]*storage.EvaluationDistribution)
		if !ok {
			// not distributions slice, bad cache
			// TODO: should we just invalidate the cache instead of returning an error?

			return nil, ErrCorrupt
		}

		return distributions, nil
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
	}

	return distributions, nil
}
