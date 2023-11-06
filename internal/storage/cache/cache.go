package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/gogo/protobuf/jsonpb"
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

// storage:evaluationRules:<namespaceKey>:<flagKey>
const evaluationRulesCacheKeyFmt = "s:er:%s:%s"

func NewStore(store storage.Store, cacher cache.Cacher, logger *zap.Logger) *Store {
	return &Store{Store: store, cacher: cacher, logger: logger}
}

func (s *Store) set(ctx context.Context, key string, values []*storage.EvaluationRule) {
	marshaller := jsonpb.Marshaler{EmitDefaults: false}

	var bytes bytes.Buffer

	bytes.WriteByte('[')

	for _, value := range values {
		if err := marshaller.Marshal(&bytes, value); err != nil {
			s.logger.Error("marshalling for storage cache", zap.Error(err))
			return
		}
	}

	bytes.WriteByte(']')

	if err := s.cacher.Set(ctx, key, bytes.Bytes()); err != nil {
		s.logger.Error("setting in storage cache", zap.Error(err))
		return
	}
}

func (s *Store) get(ctx context.Context, key string) (bool, []*storage.EvaluationRule) {
	cachePayload, cacheHit, err := s.cacher.Get(ctx, key)
	if err != nil {
		s.logger.Error("getting from storage cache", zap.Error(err))
		return false, nil
	} else if !cacheHit {
		return false, nil
	}

	var values []*storage.EvaluationRule
	decoder := json.NewDecoder(bytes.NewReader(cachePayload))
	_, err = decoder.Token() // read the opening bracket
	if err != nil {
		s.logger.Error("reading opening bracket from storage cache", zap.Error(err))
		return false, nil
	}

	for decoder.More() {
		value := &storage.EvaluationRule{}
		if err := jsonpb.UnmarshalNext(decoder, value); err != nil {
			s.logger.Error("unmarshalling from storage cache", zap.Error(err))
			return false, nil
		}

		values = append(values, value)
	}

	return true, values
}

func (s *Store) GetEvaluationRules(ctx context.Context, namespaceKey, flagKey string) ([]*storage.EvaluationRule, error) {
	cacheKey := fmt.Sprintf(evaluationRulesCacheKeyFmt, namespaceKey, flagKey)

	cacheHit, rules := s.get(ctx, cacheKey)
	if cacheHit {
		return rules, nil
	}

	rules, err := s.Store.GetEvaluationRules(ctx, namespaceKey, flagKey)
	if err != nil {
		return nil, err
	}

	s.set(ctx, cacheKey, rules)
	return rules, nil
}
