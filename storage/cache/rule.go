package cache

import (
	"context"

	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
	"github.com/sirupsen/logrus"
)

const (
	ruleCachePrefix = "r:"
	rulesCacheKey   = "r"
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

func (c *RuleCache) GetRule(ctx context.Context, r *flipt.GetRuleRequest) (*flipt.Rule, error) {
	key := ruleCacheKey(r.Id)

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

	c.cache.Set(ruleCacheKey(r.Id), rule)
	c.logger.Debugf("cache miss; added: %q", key)
	cacheMissTotal.WithLabelValues("rule", "memory").Inc()

	return rule, nil
}

func (c *RuleCache) ListRules(ctx context.Context, r *flipt.ListRuleRequest) ([]*flipt.Rule, error) {
	// check if rules exists in cache
	if data, ok := c.cache.Get(rulesCacheKey); ok {
		c.logger.Debug("cache hit: rules")
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

	c.cache.Set(rulesCacheKey, rules)
	c.logger.Debug("cache miss; added rules")
	cacheMissTotal.WithLabelValues("rules", "memory").Inc()

	return rules, nil
}

func (c *RuleCache) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.CreateRule(ctx, r)
}

func (c *RuleCache) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.UpdateRule(ctx, r)
}

func (c *RuleCache) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.DeleteRule(ctx, r)
}

func (c *RuleCache) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.OrderRules(ctx, r)
}

func (c *RuleCache) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.CreateDistribution(ctx, r)
}

func (c *RuleCache) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.UpdateDistribution(ctx, r)
}

func (c *RuleCache) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	c.cache.Flush()
	c.logger.Debug("flushed cache")
	return c.store.DeleteDistribution(ctx, r)
}

func ruleCacheKey(k string) string {
	return ruleCachePrefix + k
}
