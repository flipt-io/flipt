package cache

import (
	"context"
	"errors"

	flipt "github.com/markphelps/flipt/rpc/flipt"
	"github.com/markphelps/flipt/storage"
)

const ruleCachePrefix = "rule:"

// GetRule returns the rule from the cache if it exists; otherwise it delegates to the underlying store
// caching the result if no error
func (c *Store) GetRule(ctx context.Context, id string) (*flipt.Rule, error) {
	var (
		key   = ruleCachePrefix + id
		label = c.cache.String()
	)

	// check if rule exists in cache
	data, err := c.cache.Get(ctx, key)
	if err == nil {
		c.logger.Debugf("cache hit: %q", key)
		cacheHitTotal.WithLabelValues(label).Inc()

		rule, ok := data.(*flipt.Rule)
		if !ok {
			// not rule, bad cache
			// TODO: should we just invalidate the cache instead of returning an error?
			return nil, ErrCorrupt
		}

		return rule, nil
	}

	if !errors.Is(err, ErrMiss) {
		// some other error
		return nil, err
	}

	// rule not in cache, delegate to underlying store
	rule, err := c.store.GetRule(ctx, id)
	if err != nil {
		return rule, err
	}

	if err := c.cache.Set(ctx, key, rule); err != nil {
		return rule, err
	}

	c.logger.Debugf("cache miss; added: %q", key)
	cacheMissTotal.WithLabelValues(label).Inc()
	return rule, nil
}

// ListRules delegates to the underlying store
func (c *Store) ListRules(ctx context.Context, flagKey string, opts ...storage.QueryOption) ([]*flipt.Rule, error) {
	return c.store.ListRules(ctx, flagKey, opts...)
}

// CreateRule delegates to the underlying store, flushing the cache in the process
func (c *Store) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	c.cache.Flush(ctx)
	c.logger.Debug("flushed cache")
	return c.store.CreateRule(ctx, r)
}

// UpdateRule delegates to the underlying store, flushing the cache in the process
func (c *Store) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	c.cache.Flush(ctx)
	c.logger.Debug("flushed cache")
	return c.store.UpdateRule(ctx, r)
}

// DeleteRule delegates to the underlying store, flushing the cache in the process
func (c *Store) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	c.cache.Flush(ctx)
	c.logger.Debug("flushed cache")
	return c.store.DeleteRule(ctx, r)
}

// OrderRules delegates to the underlying store, flushing the cache in the process
func (c *Store) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	c.cache.Flush(ctx)
	c.logger.Debug("flushed cache")
	return c.store.OrderRules(ctx, r)
}

// CreateDistribution delegates to the underlying store, flushing the cache in the process
func (c *Store) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	c.cache.Flush(ctx)
	c.logger.Debug("flushed cache")
	return c.store.CreateDistribution(ctx, r)
}

// UpdateDistribution delegates to the underlying store, flushing the cache in the process
func (c *Store) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	c.cache.Flush(ctx)
	c.logger.Debug("flushed cache")
	return c.store.UpdateDistribution(ctx, r)
}

// DeleteDistribution delegates to the underlying store, flushing the cache in the process
func (c *Store) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	c.cache.Flush(ctx)
	c.logger.Debug("flushed cache")
	return c.store.DeleteDistribution(ctx, r)
}
