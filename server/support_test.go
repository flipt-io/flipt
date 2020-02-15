package server

import (
	"context"

	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
	"github.com/stretchr/testify/mock"
)

var (
	_ storage.FlagStore    = &flagStoreMock{}
	_ storage.SegmentStore = &segmentStoreMock{}
	_ storage.RuleStore    = &ruleStoreMock{}
)

type flagStoreMock struct {
	mock.Mock
}

func (m *flagStoreMock) GetFlag(ctx context.Context, key string) (*flipt.Flag, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*flipt.Flag), args.Error(1)
}

func (m *flagStoreMock) ListFlags(ctx context.Context, opts ...storage.QueryOption) ([]*flipt.Flag, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]*flipt.Flag), args.Error(1)
}

func (m *flagStoreMock) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Flag), args.Error(1)
}

func (m *flagStoreMock) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Flag), args.Error(1)
}

func (m *flagStoreMock) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *flagStoreMock) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Variant), args.Error(1)
}

func (m *flagStoreMock) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Variant), args.Error(1)
}

func (m *flagStoreMock) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

type segmentStoreMock struct {
	mock.Mock
}

func (m *segmentStoreMock) GetSegment(ctx context.Context, key string) (*flipt.Segment, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*flipt.Segment), args.Error(1)
}

func (m *segmentStoreMock) ListSegments(ctx context.Context, opts ...storage.QueryOption) ([]*flipt.Segment, error) {
	args := m.Called(ctx, opts)
	return args.Get(0).([]*flipt.Segment), args.Error(1)
}

func (m *segmentStoreMock) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Segment), args.Error(1)
}

func (m *segmentStoreMock) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Segment), args.Error(1)
}

func (m *segmentStoreMock) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *segmentStoreMock) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Constraint), args.Error(1)
}

func (m *segmentStoreMock) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Constraint), args.Error(1)
}

func (m *segmentStoreMock) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

type ruleStoreMock struct {
	mock.Mock
}

func (m *ruleStoreMock) GetRule(ctx context.Context, id string) (*flipt.Rule, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*flipt.Rule), args.Error(1)
}

func (m *ruleStoreMock) ListRules(ctx context.Context, flagKey string, opts ...storage.QueryOption) ([]*flipt.Rule, error) {
	args := m.Called(ctx, flagKey, opts)
	return args.Get(0).([]*flipt.Rule), args.Error(1)
}

func (m *ruleStoreMock) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Rule), args.Error(1)
}

func (m *ruleStoreMock) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Rule), args.Error(1)
}

func (m *ruleStoreMock) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *ruleStoreMock) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *ruleStoreMock) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Distribution), args.Error(1)
}

func (m *ruleStoreMock) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Distribution), args.Error(1)
}

func (m *ruleStoreMock) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

type evaluationStoreMock struct {
	mock.Mock
}

func (m *evaluationStoreMock) GetEvaluationRules(ctx context.Context, flagKey string) ([]*storage.EvaluationRule, error) {
	args := m.Called(ctx, flagKey)
	return args.Get(0).([]*storage.EvaluationRule), args.Error(1)
}

func (m *evaluationStoreMock) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	args := m.Called(ctx, ruleID)
	return args.Get(0).([]*storage.EvaluationDistribution), args.Error(1)
}
