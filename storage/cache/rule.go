package cache

import (
	"context"

	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
	"github.com/sirupsen/logrus"
)

const (
	ruleCachePrefix  = "rule:"
	rulesCachePrefix = "rules:flag:"
)

var _ storage.RuleStore = &RuleCache{}

// RuleCache wraps a RuleStore and provides caching
type RuleCache struct {
	logger logrus.FieldLogger
	cache  Cacher
	store  storage.RuleStore
}

// NewRuleCache creates a RuleCache by wrapping a storage.RuleStore
func NewRuleCache(logger logrus.FieldLogger, cacher Cacher, store storage.RuleStore) *RuleCache {
	return &RuleCache{
		logger: logger.WithField("cache", "memory"),
		cache:  cacher,
		store:  store,
	}
}

// GetRule returns the rule from the cache if it exists; otherwise it delegates to the underlying store
// caching the result if no error
func (c *RuleCache) GetRule(ctx context.Context, r *flipt.GetRuleRequest) (*flipt.Rule, error) {
	key := ruleCachePrefix + r.Id

	// check if rule exists in cache
	if data, ok := c.cache.Get(key); ok {
		c.logger.Debugf("cache hit: %q", key)
		cacheHitTotal.WithLabelValues("rule", "memory").Inc()

		rule, ok := data.(*flipt.Rule)
		if !ok {
			// not rule, bad cache
			return nil, ErrCacheCorrupt
		}

		return rule, nil
	}

	// else, get it and add to cache
	rule, err := c.store.GetRule(ctx, r)
	if err != nil {
		return rule, err
	}

	c.cache.Set(key, rule)
	c.logger.Debugf("cache miss; added: %q", key)
	cacheMissTotal.WithLabelValues("rule", "memory").Inc()

	return rule, nil
}

// ListRules returns all rules from the cache if they exist; otherwise it delegates to the underlying store
// caching the result if no error
func (c *RuleCache) ListRules(ctx context.Context, r *flipt.ListRuleRequest) ([]*flipt.Rule, error) {
	key := rulesCachePrefix + r.FlagKey

	// check if rules exists in cache
	if data, ok := c.cache.Get(key); ok {
		c.logger.Debugf("cache hit: %q", key)
		cacheHitTotal.WithLabelValues("rules", "memory").Inc()

		rules, ok := data.([]*flipt.Rule)
		if !ok {
			// not rules slice, bad cache
			return nil, ErrCacheCorrupt
		}

		return rules, nil
	}

	// else, get them and add to cache
	rules, err := c.store.ListRules(ctx, r)
	if err != nil {
		return rules, err
	}

	if len(rules) > 0 {
		c.cache.Set(key, rules)
		c.logger.Debugf("cache miss; added %q", key)
		cacheMissTotal.WithLabelValues("rules", "memory").Inc()
	}

	return rules, nil
}

// CreateRule delegates to the underlying store, flushing the cache in the process
func (c *RuleCache) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.CreateRule(ctx, r)
}

// UpdateRule delegates to the underlying store, flushing the cache in the process
func (c *RuleCache) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.UpdateRule(ctx, r)
}

// DeleteRule delegates to the underlying store, flushing the cache in the process
func (c *RuleCache) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.DeleteRule(ctx, r)
}

// OrderRules delegates to the underlying store, flushing the cache in the process
func (c *RuleCache) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.OrderRules(ctx, r)
}

// CreateDistribution delegates to the underlying store, flushing the cache in the process
func (c *RuleCache) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.CreateDistribution(ctx, r)
}

// UpdateDistribution delegates to the underlying store, flushing the cache in the process
func (c *RuleCache) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.UpdateDistribution(ctx, r)
}

// DeleteDistribution delegates to the underlying store, flushing the cache in the process
func (c *RuleCache) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.DeleteDistribution(ctx, r)
}
