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
	// storage:flags:<namespaceKey>:<flagKey>
	flagCacheKeyFmt = "s:f:%s:%s"
	// storage:evaluationRules:<namespaceKey>:<flagKey>
	evaluationRulesCacheKeyFmt = "s:er:%s:%s"
	// storage:evaluationRollouts:<namespaceKey>:<flagKey>
	evaluationRolloutsCacheKeyFmt = "s:ero:%s:%s"
)

func NewStore(store storage.Store, cacher cache.Cacher, logger *zap.Logger) *Store {
	return &Store{Store: store, cacher: cacher, logger: logger}
}

func (s *Store) setJSON(ctx context.Context, key string, value any) {
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

func (s *Store) getJSON(ctx context.Context, key string, value any) bool {
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

func (s *Store) setProtobuf(ctx context.Context, key string, value proto.Message) {
	cachePayload, err := proto.Marshal(value)
	if err != nil {
		s.logger.Error("marshalling for storage cache", zap.Error(err))
		return
	}

	err = s.cacher.Set(ctx, key, cachePayload)
	if err != nil {
		s.logger.Error("setting in storage cache", zap.Error(err))
	}
}

func (s *Store) getProtobuf(ctx context.Context, key string, value proto.Message) bool {
	cachePayload, cacheHit, err := s.cacher.Get(ctx, key)
	if err != nil {
		s.logger.Error("getting from storage cache", zap.Error(err))
		return false
	} else if !cacheHit {
		return false
	}

	err = proto.Unmarshal(cachePayload, value)
	if err != nil {
		s.logger.Error("unmarshalling from storage cache", zap.Error(err))
		return false
	}

	return true
}

func (s *Store) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	flag, err := s.Store.CreateFlag(ctx, r)
	if err != nil {
		return nil, err
	}

	cacheKey := fmt.Sprintf(flagCacheKeyFmt, flag.NamespaceKey, flag.Key)
	s.setProtobuf(ctx, cacheKey, flag)
	return flag, nil
}

func (s *Store) GetFlag(ctx context.Context, r storage.ResourceRequest) (*flipt.Flag, error) {
	cacheKey := fmt.Sprintf(flagCacheKeyFmt, r.Namespace(), r.Key)

	var f = &flipt.Flag{}

	cacheHit := s.getProtobuf(ctx, cacheKey, f)
	if cacheHit {
		return f, nil
	}

	f, err := s.Store.GetFlag(ctx, r)
	if err != nil {
		return nil, err
	}

	s.setProtobuf(ctx, cacheKey, f)
	return f, nil
}

func (s *Store) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	// delete from cache as flag has changed
	cacheKey := fmt.Sprintf(flagCacheKeyFmt, r.NamespaceKey, r.Key)
	err := s.cacher.Delete(ctx, cacheKey)
	if err != nil {
		s.logger.Error("deleting from storage cache", zap.Error(err))
	}

	flag, err := s.Store.UpdateFlag(ctx, r)
	if err != nil {
		return nil, err
	}

	return flag, nil
}

func (s *Store) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	cacheKey := fmt.Sprintf(flagCacheKeyFmt, r.NamespaceKey, r.Key)
	err := s.cacher.Delete(ctx, cacheKey)
	if err != nil {
		s.logger.Error("deleting from storage cache", zap.Error(err))
	}

	return s.Store.DeleteFlag(ctx, r)
}

func (s *Store) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	// delete from cache as flag has changed
	cacheKey := fmt.Sprintf(flagCacheKeyFmt, r.NamespaceKey, r.FlagKey)
	err := s.cacher.Delete(ctx, cacheKey)
	if err != nil {
		s.logger.Error("deleting from storage cache", zap.Error(err))
	}

	variant, err := s.Store.CreateVariant(ctx, r)
	if err != nil {
		return nil, err
	}

	return variant, nil
}

func (s *Store) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	// delete from cache as flag has changed
	cacheKey := fmt.Sprintf(flagCacheKeyFmt, r.NamespaceKey, r.FlagKey)
	err := s.cacher.Delete(ctx, cacheKey)
	if err != nil {
		s.logger.Error("deleting from storage cache", zap.Error(err))
	}

	variant, err := s.Store.UpdateVariant(ctx, r)
	if err != nil {
		return nil, err
	}

	return variant, nil
}

func (s *Store) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	// delete from cache as flag has changed
	cacheKey := fmt.Sprintf(flagCacheKeyFmt, r.NamespaceKey, r.FlagKey)
	err := s.cacher.Delete(ctx, cacheKey)
	if err != nil {
		s.logger.Error("deleting from storage cache", zap.Error(err))
	}

	return s.Store.DeleteVariant(ctx, r)
}

func (s *Store) GetEvaluationRules(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRule, error) {
	cacheKey := fmt.Sprintf(evaluationRulesCacheKeyFmt, flag.Namespace(), flag.Key)

	var rules []*storage.EvaluationRule

	cacheHit := s.getJSON(ctx, cacheKey, &rules)
	if cacheHit {
		return rules, nil
	}

	rules, err := s.Store.GetEvaluationRules(ctx, flag)
	if err != nil {
		return nil, err
	}

	s.setJSON(ctx, cacheKey, rules)
	return rules, nil
}

func (s *Store) GetEvaluationRollouts(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRollout, error) {
	cacheKey := fmt.Sprintf(evaluationRolloutsCacheKeyFmt, flag.Namespace(), flag.Key)

	var rollouts []*storage.EvaluationRollout

	cacheHit := s.getJSON(ctx, cacheKey, &rollouts)
	if cacheHit {
		return rollouts, nil
	}

	rollouts, err := s.Store.GetEvaluationRollouts(ctx, flag)
	if err != nil {
		return nil, err
	}

	s.setJSON(ctx, cacheKey, rollouts)
	return rollouts, nil
}
