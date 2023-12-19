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

const (
	// storage:evaluationRules:<namespaceKey>:<flagKey>
	evaluationRulesCacheKeyFmt = "s:er:%s:%s"
	// storage:evaluationRollouts:<namespaceKey>:<flagKey>
	evaluationRolloutsCacheKeyFmt = "s:ero:%s:%s"
)

func NewStore(store storage.Store, cacher cache.Cacher, logger *zap.Logger) *Store {
	return &Store{Store: store, cacher: cacher, logger: logger}
}

func (s *Store) set(ctx context.Context, key string, value any) {
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

func (s *Store) get(ctx context.Context, key string, value any) bool {
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

func (s *Store) GetEvaluationRules(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRule, error) {
	cacheKey := fmt.Sprintf(evaluationRulesCacheKeyFmt, flag.Namespace(), flag.Key)

	var rules []*storage.EvaluationRule

	cacheHit := s.get(ctx, cacheKey, &rules)
	if cacheHit {
		return rules, nil
	}

	rules, err := s.Store.GetEvaluationRules(ctx, flag)
	if err != nil {
		return nil, err
	}

	s.set(ctx, cacheKey, rules)
	return rules, nil
}

func (s *Store) GetEvaluationRollouts(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRollout, error) {
	cacheKey := fmt.Sprintf(evaluationRolloutsCacheKeyFmt, flag.Namespace(), flag.Key)

	var rollouts []*storage.EvaluationRollout

	cacheHit := s.get(ctx, cacheKey, &rollouts)
	if cacheHit {
		return rollouts, nil
	}

	rollouts, err := s.Store.GetEvaluationRollouts(ctx, flag)
	if err != nil {
		return nil, err
	}

	s.set(ctx, cacheKey, rollouts)
	return rollouts, nil
}
