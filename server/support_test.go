package server

import (
	"context"

	"github.com/stretchr/testify/mock"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/server/cache"
	"go.flipt.io/flipt/storage"
)

var (
	_ storage.Store = &storeMock{}
)

type storeMock struct {
	mock.Mock
}

func (m *storeMock) String() string {
	return "mock"
}

func (m *storeMock) GetFlag(ctx context.Context, key string) (*flipt.Flag, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*flipt.Flag), args.Error(1)
}

func (m *storeMock) ListFlags(ctx context.Context, opts ...storage.QueryOption) ([]*flipt.Flag, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]*flipt.Flag), args.Error(1)
}

func (m *storeMock) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Flag), args.Error(1)
}

func (m *storeMock) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Flag), args.Error(1)
}

func (m *storeMock) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *storeMock) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Variant), args.Error(1)
}

func (m *storeMock) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Variant), args.Error(1)
}

func (m *storeMock) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *storeMock) GetSegment(ctx context.Context, key string) (*flipt.Segment, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*flipt.Segment), args.Error(1)
}

func (m *storeMock) ListSegments(ctx context.Context, opts ...storage.QueryOption) ([]*flipt.Segment, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]*flipt.Segment), args.Error(1)
}

func (m *storeMock) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Segment), args.Error(1)
}

func (m *storeMock) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Segment), args.Error(1)
}

func (m *storeMock) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *storeMock) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Constraint), args.Error(1)
}

func (m *storeMock) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Constraint), args.Error(1)
}

func (m *storeMock) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *storeMock) GetRule(ctx context.Context, id string) (*flipt.Rule, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*flipt.Rule), args.Error(1)
}

func (m *storeMock) ListRules(ctx context.Context, flagKey string, opts ...storage.QueryOption) ([]*flipt.Rule, error) {
	args := m.Called(ctx, flagKey, opts)
	return args.Get(0).([]*flipt.Rule), args.Error(1)
}

func (m *storeMock) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Rule), args.Error(1)
}

func (m *storeMock) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Rule), args.Error(1)
}

func (m *storeMock) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *storeMock) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *storeMock) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Distribution), args.Error(1)
}

func (m *storeMock) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Distribution), args.Error(1)
}

func (m *storeMock) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *storeMock) GetEvaluationRules(ctx context.Context, flagKey string) ([]*storage.EvaluationRule, error) {
	args := m.Called(ctx, flagKey)
	return args.Get(0).([]*storage.EvaluationRule), args.Error(1)
}

func (m *storeMock) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	args := m.Called(ctx, ruleID)
	return args.Get(0).([]*storage.EvaluationDistribution), args.Error(1)
}

type cacheSpy struct {
	cache.Cacher

	getKeys   map[string]struct{}
	getCalled int

	setItems  map[string][]byte
	setCalled int

	deleteKeys   map[string]struct{}
	deleteCalled int
}

func newCacheSpy(c cache.Cacher) *cacheSpy {
	return &cacheSpy{
		Cacher:     c,
		getKeys:    make(map[string]struct{}),
		setItems:   make(map[string][]byte),
		deleteKeys: make(map[string]struct{}),
	}
}

func (c *cacheSpy) Get(ctx context.Context, key string) ([]byte, bool, error) {
	c.getCalled++
	c.getKeys[key] = struct{}{}
	return c.Cacher.Get(ctx, key)
}

func (c *cacheSpy) Set(ctx context.Context, key string, value []byte) error {
	c.setCalled++
	c.setItems[key] = value
	return c.Cacher.Set(ctx, key, value)
}

func (c *cacheSpy) Delete(ctx context.Context, key string) error {
	c.deleteCalled++
	c.deleteKeys[key] = struct{}{}
	return c.Cacher.Delete(ctx, key)
}
