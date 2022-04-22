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
			return nil, ErrCorrupt
		}

		return rules, nil
	} else if !errors.Is(err, ErrNotFound) {
		c.logger.WithError(err).Warnf("failed to get cache: %q", key)
	}

	// else, get them and add to cache
	rules, err := c.store.GetEvaluationRules(ctx, flagKey)
	if err != nil {
		return rules, err
	}

	if len(rules) > 0 {
		if err := c.cache.Set(ctx, key, rules); err != nil {
			c.logger.WithError(err).Warnf("failed to set cache: %q", key)
		} else {
			c.logger.Debugf("cache miss; added: %q", key)
		}

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
			return nil, ErrCorrupt
		}

		return distributions, nil
	} else if !errors.Is(err, ErrNotFound) {
		c.logger.WithError(err).Warnf("failed to get cache: %q", key)
	}

	// else, get them and add to cache
	distributions, err := c.store.GetEvaluationDistributions(ctx, ruleID)
	if err != nil {
		return distributions, err
	}

	if len(distributions) > 0 {
		if err := c.cache.Set(ctx, key, distributions); err != nil {
			c.logger.WithError(err).Warnf("failed to set cache: %q", key)
		} else {
			c.logger.Debugf("cache miss; added %q", key)
		}

		cacheMissTotal.WithLabelValues(label).Inc()
	}

	return distributions, nil
}
