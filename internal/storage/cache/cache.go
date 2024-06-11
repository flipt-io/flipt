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

type marshaller[T any] interface {
	Marshal(T) ([]byte, error)
}

type marshalFunc[T any] func(T) ([]byte, error)

func (f marshalFunc[T]) Marshal(v T) ([]byte, error) {
	return f(v)
}

type unmarshaller[T any] interface {
	Unmarshal([]byte, T) error
}

type unmarshalFunc[T any] func([]byte, T) error

func (f unmarshalFunc[T]) Unmarshal(b []byte, v T) error {
	return f(b, v)
}

func set[T any](ctx context.Context, s *Store, m marshaller[T], key string, value T) {
	cachePayload, err := m.Marshal(value)
	if err != nil {
		s.logger.Error("marshalling for storage cache", zap.Error(err))
		return
	}

	err = s.cacher.Set(ctx, key, cachePayload)
	if err != nil {
		s.logger.Error("setting in storage cache", zap.Error(err))
	}
}

func get[T any](ctx context.Context, s *Store, u unmarshaller[T], key string, value T) bool {
	cachePayload, cacheHit, err := s.cacher.Get(ctx, key)
	if err != nil {
		s.logger.Error("getting from storage cache", zap.Error(err))
		return false
	} else if !cacheHit {
		return false
	}

	err = u.Unmarshal(cachePayload, value)
	if err != nil {
		s.logger.Error("unmarshalling from storage cache", zap.Error(err))
		return false
	}

	return true
}

var _ storage.Store = &Store{}

type Store struct {
	storage.Store
	cacher cache.Cacher
	logger *zap.Logger
}

const (
	// storage:flags
	flagCacheKeyPrefix = "s:f"
	// storage:evaluationRules
	evaluationRulesCacheKeyPrefix = "s:er"
	// storage:evaluationRollouts
	evaluationRolloutsCacheKeyPrefix = "s:ero"
)

func NewStore(store storage.Store, cacher cache.Cacher, logger *zap.Logger) *Store {
	return &Store{Store: store, cacher: cacher, logger: logger}
}

func (s *Store) setJSON(ctx context.Context, key string, value any) {
	set(ctx, s, marshalFunc[any](json.Marshal), key, value)
}

func (s *Store) getJSON(ctx context.Context, key string, value any) bool {
	return get(ctx, s, unmarshalFunc[any](json.Unmarshal), key, value)
}

func (s *Store) setProtobuf(ctx context.Context, key string, value proto.Message) {
	set(ctx, s, marshalFunc[proto.Message](proto.Marshal), key, value)
}

func (s *Store) getProtobuf(ctx context.Context, key string, value proto.Message) bool {
	return get(ctx, s, unmarshalFunc[proto.Message](proto.Unmarshal), key, value)
}

func (s *Store) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	flag, err := s.Store.CreateFlag(ctx, r)
	if err != nil {
		return nil, err
	}

	cacheKey := cacheKey(flagCacheKeyPrefix, storage.NewResource(r.NamespaceKey, r.Key))
	s.setProtobuf(ctx, cacheKey, flag)
	return flag, nil
}

func (s *Store) GetFlag(ctx context.Context, r storage.ResourceRequest) (*flipt.Flag, error) {
	var (
		f        = &flipt.Flag{}
		cacheKey = cacheKey(flagCacheKeyPrefix, r)
		cacheHit = s.getProtobuf(ctx, cacheKey, f)
	)

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
	cacheKey := cacheKey(flagCacheKeyPrefix, storage.NewResource(r.NamespaceKey, r.Key))
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
	cacheKey := cacheKey(flagCacheKeyPrefix, storage.NewResource(r.NamespaceKey, r.Key))
	err := s.cacher.Delete(ctx, cacheKey)
	if err != nil {
		s.logger.Error("deleting from storage cache", zap.Error(err))
	}

	return s.Store.DeleteFlag(ctx, r)
}

func (s *Store) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	// delete from cache as flag has changed
	cacheKey := cacheKey(flagCacheKeyPrefix, storage.NewResource(r.NamespaceKey, r.FlagKey))
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
	cacheKey := cacheKey(flagCacheKeyPrefix, storage.NewResource(r.NamespaceKey, r.FlagKey))
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
	cacheKey := cacheKey(flagCacheKeyPrefix, storage.NewResource(r.NamespaceKey, r.FlagKey))
	err := s.cacher.Delete(ctx, cacheKey)
	if err != nil {
		s.logger.Error("deleting from storage cache", zap.Error(err))
	}

	return s.Store.DeleteVariant(ctx, r)
}

func (s *Store) GetEvaluationRules(ctx context.Context, r storage.ResourceRequest) ([]*storage.EvaluationRule, error) {
	var (
		rules    []*storage.EvaluationRule
		cacheKey = cacheKey(evaluationRulesCacheKeyPrefix, r)
		cacheHit = s.getJSON(ctx, cacheKey, &rules)
	)

	if cacheHit {
		return rules, nil
	}

	rules, err := s.Store.GetEvaluationRules(ctx, r)
	if err != nil {
		return nil, err
	}

	s.setJSON(ctx, cacheKey, rules)
	return rules, nil
}

func (s *Store) GetEvaluationRollouts(ctx context.Context, r storage.ResourceRequest) ([]*storage.EvaluationRollout, error) {
	var (
		rollouts []*storage.EvaluationRollout
		cacheKey = cacheKey(evaluationRolloutsCacheKeyPrefix, r)
		cacheHit = s.getJSON(ctx, cacheKey, &rollouts)
	)

	if cacheHit {
		return rollouts, nil
	}

	rollouts, err := s.Store.GetEvaluationRollouts(ctx, r)
	if err != nil {
		return nil, err
	}

	s.setJSON(ctx, cacheKey, rollouts)
	return rollouts, nil
}

func cacheKey(prefix string, r storage.ResourceRequest) string {
	// <prefix>:<namespaceKey>:<flagKey>
	return fmt.Sprintf("%s:%s:%s", prefix, r.Namespace(), r.Key)
}
