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
	// storage:namespaces
	namespaceCacheKeyPrefix = "s:n"
	// storage:flags
	flagCacheKeyPrefix = "s:f"
	// storage:evaluationRules
	evaluationRulesCacheKeyPrefix = "s:er"
	// storage:evaluationRollouts
	evaluationRolloutsCacheKeyPrefix = "s:ero"
	// storage:evaluationDistributions
	evaluationDistributionsCacheKeyPrefix = "s:erd"
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

func (s *Store) UpdateNamespace(ctx context.Context, r *flipt.UpdateNamespaceRequest) (*flipt.Namespace, error) {
	namespace, err := s.Store.UpdateNamespace(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetKey()))
	return namespace, err
}

func (s *Store) DeleteNamespace(ctx context.Context, r *flipt.DeleteNamespaceRequest) error {
	err := s.Store.DeleteNamespace(ctx, r)
	if err != nil {
		return err
	}

	cacheNsKey := s.cacheNsKey(storage.NewNamespace(r.GetKey()))
	err = s.cacher.Delete(ctx, cacheNsKey)
	if err != nil {
		s.logger.Error("deleting from storage cache", zap.Error(err))
	}

	return nil
}

func (s *Store) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	v, err := s.Store.CreateFlag(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return v, err
}

func (s *Store) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	v, err := s.Store.UpdateFlag(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return v, err
}

func (s *Store) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	err := s.Store.DeleteFlag(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return err
}

func (s *Store) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	v, err := s.Store.CreateVariant(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return v, err
}

func (s *Store) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	v, err := s.Store.UpdateVariant(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return v, err
}

func (s *Store) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	err := s.Store.DeleteVariant(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return err
}

func (s *Store) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	v, err := s.Store.CreateSegment(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return v, err
}

func (s *Store) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	v, err := s.Store.UpdateSegment(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return v, err
}

func (s *Store) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	err := s.Store.DeleteSegment(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return err
}

func (s *Store) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	v, err := s.Store.CreateConstraint(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return v, err
}

func (s *Store) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	v, err := s.Store.UpdateConstraint(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return v, err
}

func (s *Store) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	err := s.Store.DeleteConstraint(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return err
}

func (s *Store) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	v, err := s.Store.CreateRule(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return v, err
}

func (s *Store) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	v, err := s.Store.UpdateRule(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return v, err
}

func (s *Store) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	err := s.Store.DeleteRule(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return err
}

func (s *Store) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	err := s.Store.OrderRules(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return err
}

func (s *Store) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	v, err := s.Store.CreateDistribution(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return v, err
}

func (s *Store) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	v, err := s.Store.UpdateDistribution(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return v, err
}

func (s *Store) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	err := s.Store.DeleteDistribution(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return err
}

func (s *Store) CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (*flipt.Rollout, error) {
	v, err := s.Store.CreateRollout(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return v, err
}

func (s *Store) UpdateRollout(ctx context.Context, r *flipt.UpdateRolloutRequest) (*flipt.Rollout, error) {
	v, err := s.Store.UpdateRollout(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return v, err
}

func (s *Store) DeleteRollout(ctx context.Context, r *flipt.DeleteRolloutRequest) error {
	err := s.Store.DeleteRollout(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return err
}

func (s *Store) OrderRollouts(ctx context.Context, r *flipt.OrderRolloutsRequest) error {
	err := s.Store.OrderRollouts(ctx, r)
	s.cacheUpdateNamespacedVersion(ctx, storage.NewNamespace(r.GetNamespaceKey()))
	return err
}

func (s *Store) ListFlags(ctx context.Context, r *storage.ListRequest[storage.NamespaceRequest]) (storage.ResultSet[*flipt.Flag], error) {
	ns := storage.NewResource(r.Predicate.Namespace(), fmt.Sprintf("flags/%s-%d-%d", r.QueryParams.PageToken, r.QueryParams.Offset, r.QueryParams.Limit))
	namespaceVersion, _ := s.GetVersion(ctx, ns.NamespaceRequest)
	var (
		f        = storage.ResultSet[*flipt.Flag]{}
		cacheKey = cacheKey(flagCacheKeyPrefix, ns, namespaceVersion)
		cacheHit = s.getJSON(ctx, cacheKey, &f)
	)

	if cacheHit {
		return f, nil
	}

	f, err := s.Store.ListFlags(ctx, r)
	if err != nil {
		return f, err
	}

	s.setJSON(ctx, cacheKey, f)
	return f, nil
}

func (s *Store) GetFlag(ctx context.Context, r storage.ResourceRequest) (*flipt.Flag, error) {
	namespaceVersion, _ := s.GetVersion(ctx, r.NamespaceRequest)
	var (
		f        = &flipt.Flag{}
		cacheKey = cacheKey(flagCacheKeyPrefix, r, namespaceVersion)
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

func (s *Store) GetEvaluationRules(ctx context.Context, r storage.ResourceRequest) ([]*storage.EvaluationRule, error) {
	namespaceVersion, _ := s.GetVersion(ctx, r.NamespaceRequest)

	var (
		rules    []*storage.EvaluationRule
		cacheKey = cacheKey(evaluationRulesCacheKeyPrefix, r, namespaceVersion)
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
	namespaceVersion, _ := s.GetVersion(ctx, r.NamespaceRequest)

	var (
		rollouts []*storage.EvaluationRollout
		cacheKey = cacheKey(evaluationRolloutsCacheKeyPrefix, r, namespaceVersion)
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

func (s *Store) GetEvaluationDistributions(ctx context.Context, r storage.ResourceRequest, rule storage.IDRequest) ([]*storage.EvaluationDistribution, error) {
	namespaceVersion, _ := s.GetVersion(ctx, r.NamespaceRequest)

	var (
		distributions []*storage.EvaluationDistribution
		cacheKey      = fmt.Sprintf("%s/%s", cacheKey(evaluationDistributionsCacheKeyPrefix, r, namespaceVersion), rule.ID)
		cacheHit      = s.getJSON(ctx, cacheKey, &distributions)
	)

	if cacheHit {
		return distributions, nil
	}

	distributions, err := s.Store.GetEvaluationDistributions(ctx, r, rule)
	if err != nil {
		return nil, err
	}

	s.setJSON(ctx, cacheKey, distributions)
	return distributions, nil
}

func (s *Store) GetVersion(ctx context.Context, ns storage.NamespaceRequest) (string, error) {
	cacheNsKey := s.cacheNsKey(ns)
	version, hits, err := s.cacher.Get(ctx, cacheNsKey)
	if err != nil {
		s.logger.Error("failed to get version to storage cache", zap.Error(err))
	}

	if hits {
		return string(version), nil
	}

	originalVersion, err := s.Store.GetVersion(ctx, ns)
	if err == nil {
		err := s.cacher.Set(ctx, cacheNsKey, []byte(originalVersion))
		if err != nil {
			s.logger.Error("failed to set version to storage cache", zap.Error(err))
		}
	}
	return originalVersion, err
}

func (s *Store) cacheUpdateNamespacedVersion(ctx context.Context, r storage.NamespaceRequest) {
	cacheNsKey := s.cacheNsKey(r)
	version, err := s.Store.GetVersion(ctx, r)
	if err != nil {
		s.logger.Error("getting from storage cache", zap.Error(err))
		return
	}
	err = s.cacher.Set(ctx, cacheNsKey, []byte(version))
	if err != nil {
		s.logger.Error("updating from storage cache", zap.Error(err))
	}
}

func (*Store) cacheNsKey(r storage.NamespaceRequest) string {
	cacheKey := fmt.Sprintf("%s:%s", namespaceCacheKeyPrefix, r.Namespace())
	return cacheKey
}

func cacheKey(prefix string, r storage.ResourceRequest, version string) string {
	// <prefix>:<namespaceKey>:<flagKey>
	return fmt.Sprintf("%s:%s:%s:%s", prefix, r.Namespace(), version, r.Key)
}
