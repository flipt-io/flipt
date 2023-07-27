package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"go.flipt.io/flipt/internal/cache"
	"go.flipt.io/flipt/internal/storage"
	"go.uber.org/zap"
)

var _ storage.Store = &Store{}

type Store struct {
	storage.Store
	cacher cache.Cacher
	logger *zap.Logger
}

func New(store storage.Store, cacher cache.Cacher, logger *zap.Logger) storage.Store {
	return &Store{store, cacher, logger}
}

func (s *Store) setCache(ctx context.Context, key string, value any) {
	cachePayload, err := json.Marshal(value)
	if err != nil {
		s.logger.Error("marshalling for storage cache", zap.Error(err))
		return
	}

	err = s.cacher.Set(ctx, key, cachePayload)
	if err != nil {
		s.logger.Error("setting in storage cache", zap.Error(err))
	}
}

func (s *Store) getCache(ctx context.Context, key string, value any) bool {
	cachePayload, cacheHit, err := s.cacher.Get(ctx, key)
	if err != nil {
		s.logger.Error("getting from storage cache", zap.Error(err))
		return false
	} else if !cacheHit {
		return false
	}

	err = json.Unmarshal(cachePayload, value)
	if err != nil {
		s.logger.Error("unmarshalling from storage cache", zap.Error(err))
		return false
	}

	return true
}

func (s *Store) GetEvaluationRules(ctx context.Context, namespaceKey, flagKey string) ([]*storage.EvaluationRule, error) {
	cacheKey := fmt.Sprintf("s:er:%s:%s", namespaceKey, flagKey)
	var rules []*storage.EvaluationRule

	cacheHit := s.getCache(ctx, cacheKey, &rules)
	if cacheHit {
		s.logger.Debug("evaluation rules storage cache hit", zap.String("cache_key", cacheKey))
		return rules, nil
	}

	rules, err := s.Store.GetEvaluationRules(ctx, namespaceKey, flagKey)
	if err != nil {
		return nil, err
	}

	s.setCache(ctx, cacheKey, rules)

	return rules, nil
}
