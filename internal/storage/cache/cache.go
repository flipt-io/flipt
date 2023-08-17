package cache

import (
	"context"
	"encoding/json"
	"fmt"

	"go.flipt.io/flipt/internal/cache"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
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
	// storage:flag:<namespaceKey>:<flagKey>
	flagCacheKeyFmt = "s:f:%s:%s"
)

func NewStore(store storage.Store, cacher cache.Cacher, logger *zap.Logger) *Store {
	return &Store{Store: store, cacher: cacher, logger: logger}
}

func (s *Store) set(ctx context.Context, key string, value interface{}, marshal func(interface{}) ([]byte, error)) {
	cachePayload, err := marshal(value)
	if err != nil {
		s.logger.Error("marshalling for storage cache", zap.Error(err))
		return
	}

	err = s.cacher.Set(ctx, key, cachePayload)
	if err != nil {
		s.logger.Error("setting in storage cache", zap.Error(err))
	}
}

func (s *Store) get(ctx context.Context, key string, value interface{}, unmarshal func([]byte, interface{}) error) bool {
	cachePayload, cacheHit, err := s.cacher.Get(ctx, key)
	if err != nil {
		s.logger.Error("getting from storage cache", zap.Error(err))
		return false
	} else if !cacheHit {
		return false
	}

	err = unmarshal(cachePayload, value)
	if err != nil {
		s.logger.Error("unmarshalling from storage cache", zap.Error(err))
		return false
	}

	return true
}

func (s *Store) setProto(ctx context.Context, key string, value proto.Message) {
	s.set(ctx, key, value, func(v interface{}) ([]byte, error) {
		return proto.Marshal(v.(proto.Message))
	})
}

func (s *Store) getProto(ctx context.Context, key string, value proto.Message) bool {
	return s.get(ctx, key, value, func(data []byte, v interface{}) error {
		return proto.Unmarshal(data, v.(proto.Message))
	})
}

func (s *Store) setJSON(ctx context.Context, key string, value any) {
	s.set(ctx, key, value, json.Marshal)
}

func (s *Store) getJSON(ctx context.Context, key string, value any) bool {
	return s.get(ctx, key, value, json.Unmarshal)
}

func (s *Store) GetFlag(ctx context.Context, namespaceKey, key string) (*flipt.Flag, error) {
	cacheKey := fmt.Sprintf(flagCacheKeyFmt, namespaceKey, key)

	var flag = &flipt.Flag{}

	cacheHit := s.getProto(ctx, cacheKey, flag)
	if cacheHit {
		return flag, nil
	}

	flag, err := s.Store.GetFlag(ctx, namespaceKey, key)
	if err != nil {
		return nil, err
	}

	s.setProto(ctx, cacheKey, flag)
	return flag, nil
}

func (s *Store) GetEvaluationRules(ctx context.Context, namespaceKey, flagKey string) ([]*storage.EvaluationRule, error) {
	cacheKey := fmt.Sprintf(evaluationRulesCacheKeyFmt, namespaceKey, flagKey)

	var rules []*storage.EvaluationRule

	cacheHit := s.getJSON(ctx, cacheKey, &rules)
	if cacheHit {
		return rules, nil
	}

	rules, err := s.Store.GetEvaluationRules(ctx, namespaceKey, flagKey)
	if err != nil {
		return nil, err
	}

	s.setJSON(ctx, cacheKey, rules)
	return rules, nil
}
