package common

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
)

var _ storage.Store = &StoreMock{}

type StoreMock struct {
	mock.Mock
}

func (m *StoreMock) String() string {
	return "mock"
}

func (m *StoreMock) GetVersion(ctx context.Context, ns storage.NamespaceRequest) (string, error) {
	args := m.Called(ctx, ns)
	return args.String(0), args.Error(1)
}

func (m *StoreMock) GetNamespace(ctx context.Context, ns storage.NamespaceRequest) (*flipt.Namespace, error) {
	args := m.Called(ctx, ns)
	return args.Get(0).(*flipt.Namespace), args.Error(1)
}

func (m *StoreMock) ListNamespaces(ctx context.Context, req *storage.ListRequest[storage.ReferenceRequest]) (storage.ResultSet[*flipt.Namespace], error) {
	args := m.Called(ctx, req)
	return args.Get(0).(storage.ResultSet[*flipt.Namespace]), args.Error(1)
}

func (m *StoreMock) CountNamespaces(ctx context.Context, p storage.ReferenceRequest) (uint64, error) {
	args := m.Called(ctx, p)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *StoreMock) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Namespace), args.Error(1)
}

func (m *StoreMock) UpdateNamespace(ctx context.Context, r *flipt.UpdateNamespaceRequest) (*flipt.Namespace, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Namespace), args.Error(1)
}

func (m *StoreMock) DeleteNamespace(ctx context.Context, r *flipt.DeleteNamespaceRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *StoreMock) GetFlag(ctx context.Context, flag storage.ResourceRequest) (*flipt.Flag, error) {
	args := m.Called(ctx, flag)
	return args.Get(0).(*flipt.Flag), args.Error(1)
}

func (m *StoreMock) ListFlags(ctx context.Context, req *storage.ListRequest[storage.NamespaceRequest]) (storage.ResultSet[*flipt.Flag], error) {
	args := m.Called(ctx, req)
	return args.Get(0).(storage.ResultSet[*flipt.Flag]), args.Error(1)
}

func (m *StoreMock) CountFlags(ctx context.Context, ns storage.NamespaceRequest) (uint64, error) {
	args := m.Called(ctx, ns)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *StoreMock) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Flag), args.Error(1)
}

func (m *StoreMock) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Flag), args.Error(1)
}

func (m *StoreMock) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *StoreMock) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Variant), args.Error(1)
}

func (m *StoreMock) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Variant), args.Error(1)
}

func (m *StoreMock) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *StoreMock) GetSegment(ctx context.Context, segment storage.ResourceRequest) (*flipt.Segment, error) {
	args := m.Called(ctx, segment)
	return args.Get(0).(*flipt.Segment), args.Error(1)
}

func (m *StoreMock) ListSegments(ctx context.Context, req *storage.ListRequest[storage.NamespaceRequest]) (storage.ResultSet[*flipt.Segment], error) {
	args := m.Called(ctx, req)
	return args.Get(0).(storage.ResultSet[*flipt.Segment]), args.Error(1)
}

func (m *StoreMock) CountSegments(ctx context.Context, ns storage.NamespaceRequest) (uint64, error) {
	args := m.Called(ctx, ns)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *StoreMock) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Segment), args.Error(1)
}

func (m *StoreMock) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Segment), args.Error(1)
}

func (m *StoreMock) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *StoreMock) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Constraint), args.Error(1)
}

func (m *StoreMock) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Constraint), args.Error(1)
}

func (m *StoreMock) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *StoreMock) ListRollouts(ctx context.Context, req *storage.ListRequest[storage.ResourceRequest]) (storage.ResultSet[*flipt.Rollout], error) {
	args := m.Called(ctx, req)
	return args.Get(0).(storage.ResultSet[*flipt.Rollout]), args.Error(1)
}

func (m *StoreMock) CountRollouts(ctx context.Context, flag storage.ResourceRequest) (uint64, error) {
	args := m.Called(ctx, flag)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *StoreMock) GetRollout(ctx context.Context, ns storage.NamespaceRequest, id string) (*flipt.Rollout, error) {
	args := m.Called(ctx, ns, id)
	return args.Get(0).(*flipt.Rollout), args.Error(1)
}

func (m *StoreMock) CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (*flipt.Rollout, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Rollout), args.Error(1)
}

func (m *StoreMock) UpdateRollout(ctx context.Context, r *flipt.UpdateRolloutRequest) (*flipt.Rollout, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Rollout), args.Error(1)
}

func (m *StoreMock) DeleteRollout(ctx context.Context, r *flipt.DeleteRolloutRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *StoreMock) OrderRollouts(ctx context.Context, r *flipt.OrderRolloutsRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *StoreMock) GetRule(ctx context.Context, ns storage.NamespaceRequest, id string) (*flipt.Rule, error) {
	args := m.Called(ctx, ns, id)
	return args.Get(0).(*flipt.Rule), args.Error(1)
}

func (m *StoreMock) ListRules(ctx context.Context, req *storage.ListRequest[storage.ResourceRequest]) (storage.ResultSet[*flipt.Rule], error) {
	args := m.Called(ctx, req)
	return args.Get(0).(storage.ResultSet[*flipt.Rule]), args.Error(1)
}

func (m *StoreMock) CountRules(ctx context.Context, flag storage.ResourceRequest) (uint64, error) {
	args := m.Called(ctx, flag)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *StoreMock) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Rule), args.Error(1)
}

func (m *StoreMock) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Rule), args.Error(1)
}

func (m *StoreMock) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *StoreMock) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *StoreMock) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Distribution), args.Error(1)
}

func (m *StoreMock) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Distribution), args.Error(1)
}

func (m *StoreMock) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *StoreMock) GetEvaluationRules(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRule, error) {
	args := m.Called(ctx, flag)
	return args.Get(0).([]*storage.EvaluationRule), args.Error(1)
}

func (m *StoreMock) GetEvaluationDistributions(ctx context.Context, rule storage.IDRequest) ([]*storage.EvaluationDistribution, error) {
	args := m.Called(ctx, rule)
	return args.Get(0).([]*storage.EvaluationDistribution), args.Error(1)
}

func (m *StoreMock) GetEvaluationRollouts(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRollout, error) {
	args := m.Called(ctx, flag)
	return args.Get(0).([]*storage.EvaluationRollout), args.Error(1)
}
