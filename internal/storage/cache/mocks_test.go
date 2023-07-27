package cache

import (
	"context"

	"go.flipt.io/flipt/internal/cache"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
)

var _ storage.Store = &mockStore{}

type mockStore struct {
	EvaluationRules []*storage.EvaluationRule
	Flags           []flipt.Flag
	ReturnErr       error
}

func (s *mockStore) String() string {
	return "mockStore"
}

func (s *mockStore) GetEvaluationRules(ctx context.Context, namespaceKey, flagKey string) ([]*storage.EvaluationRule, error) {
	if s.ReturnErr != nil {
		return nil, s.ReturnErr
	}

	return s.EvaluationRules, nil
}

func (s *mockStore) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	return nil, nil
}

func (s *mockStore) GetNamespace(ctx context.Context, key string) (*flipt.Namespace, error) {
	return nil, nil
}

func (s *mockStore) ListNamespaces(ctx context.Context, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Namespace], error) {
	return storage.ResultSet[*flipt.Namespace]{}, nil
}

func (s *mockStore) CountNamespaces(ctx context.Context) (uint64, error) {
	return 0, nil
}

func (s *mockStore) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	return nil, nil
}

func (s *mockStore) UpdateNamespace(ctx context.Context, r *flipt.UpdateNamespaceRequest) (*flipt.Namespace, error) {
	return nil, nil
}

func (s *mockStore) DeleteNamespace(ctx context.Context, r *flipt.DeleteNamespaceRequest) error {
	return nil
}

func (s *mockStore) GetFlag(ctx context.Context, namespaceKey, key string) (*flipt.Flag, error) {
	return nil, nil
}

func (s *mockStore) ListFlags(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Flag], error) {
	return storage.ResultSet[*flipt.Flag]{}, nil
}

func (s *mockStore) CountFlags(ctx context.Context, namespaceKey string) (uint64, error) {
	return 0, nil
}

func (s *mockStore) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	return nil, nil
}

func (s *mockStore) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	return nil, nil
}

func (s *mockStore) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	return nil
}

func (s *mockStore) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	return nil, nil
}

func (s *mockStore) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	return nil, nil
}

func (s *mockStore) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	return nil
}

func (s *mockStore) GetSegment(ctx context.Context, namespaceKey, key string) (*flipt.Segment, error) {
	return nil, nil
}

func (s *mockStore) ListSegments(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Segment], error) {
	return storage.ResultSet[*flipt.Segment]{}, nil
}

func (s *mockStore) CountSegments(ctx context.Context, namespaceKey string) (uint64, error) {
	return 0, nil
}

func (s *mockStore) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	return nil, nil
}

func (s *mockStore) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	return nil, nil
}

func (s *mockStore) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	return nil
}

func (s *mockStore) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	return nil, nil
}

func (s *mockStore) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	return nil, nil
}

func (s *mockStore) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	return nil
}

func (s *mockStore) GetRule(ctx context.Context, namespaceKey, id string) (*flipt.Rule, error) {
	return nil, nil
}

func (s *mockStore) ListRules(ctx context.Context, namespaceKey, flagKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Rule], error) {
	return storage.ResultSet[*flipt.Rule]{}, nil
}

func (s *mockStore) CountRules(ctx context.Context, namespaceKey, flagKey string) (uint64, error) {
	return 0, nil
}

func (s *mockStore) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	return nil, nil
}

func (s *mockStore) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	return nil, nil
}

func (s *mockStore) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	return nil
}

func (s *mockStore) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	return nil
}

func (s *mockStore) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	return nil, nil
}

func (s *mockStore) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	return nil, nil
}

func (s *mockStore) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	return nil
}

var _ cache.Cacher = &mockCacher{}

type mockCacher struct {
	Cached      bool
	CachedValue []byte
	CacheKey    string
	GetErr      error
	SetErr      error
}

func (c *mockCacher) String() string {
	return "mockCacher"
}

func (c *mockCacher) Get(ctx context.Context, key string) ([]byte, bool, error) {
	c.CacheKey = key

	if c.GetErr != nil || !c.Cached {
		return nil, c.Cached, c.GetErr
	}

	return c.CachedValue, true, nil
}

func (c *mockCacher) Set(ctx context.Context, key string, value []byte) error {
	c.CacheKey = key
	c.CachedValue = value

	if c.SetErr != nil {
		return c.SetErr
	}

	return nil
}

func (c *mockCacher) Delete(ctx context.Context, key string) error {
	return nil
}
