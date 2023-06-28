package grpc_middleware

import (
	"context"
	"fmt"

	"github.com/stretchr/testify/mock"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/internal/server/cache"
	"go.flipt.io/flipt/internal/storage"
	storageauth "go.flipt.io/flipt/internal/storage/auth"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ storage.Store = &storeMock{}

type storeMock struct {
	mock.Mock
}

func (m *storeMock) String() string {
	return "mock"
}

func (m *storeMock) GetNamespace(ctx context.Context, key string) (*flipt.Namespace, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(*flipt.Namespace), args.Error(1)
}

func (m *storeMock) ListNamespaces(ctx context.Context, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Namespace], error) {
	args := m.Called(ctx, opts)
	return args.Get(0).(storage.ResultSet[*flipt.Namespace]), args.Error(1)
}

func (m *storeMock) CountNamespaces(ctx context.Context) (uint64, error) {
	args := m.Called(ctx)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *storeMock) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Namespace), args.Error(1)
}

func (m *storeMock) UpdateNamespace(ctx context.Context, r *flipt.UpdateNamespaceRequest) (*flipt.Namespace, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Namespace), args.Error(1)
}

func (m *storeMock) DeleteNamespace(ctx context.Context, r *flipt.DeleteNamespaceRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *storeMock) GetFlag(ctx context.Context, namespaceKey string, key string) (*flipt.Flag, error) {
	args := m.Called(ctx, namespaceKey, key)
	return args.Get(0).(*flipt.Flag), args.Error(1)
}

func (m *storeMock) ListFlags(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Flag], error) {
	args := m.Called(ctx, namespaceKey, opts)
	return args.Get(0).(storage.ResultSet[*flipt.Flag]), args.Error(1)
}

func (m *storeMock) CountFlags(ctx context.Context, namespaceKey string) (uint64, error) {
	args := m.Called(ctx, namespaceKey)
	return args.Get(0).(uint64), args.Error(1)
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

func (m *storeMock) GetSegment(ctx context.Context, namespaceKey string, key string) (*flipt.Segment, error) {
	args := m.Called(ctx, namespaceKey, key)
	return args.Get(0).(*flipt.Segment), args.Error(1)
}

func (m *storeMock) ListSegments(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Segment], error) {
	args := m.Called(ctx, namespaceKey, opts)
	return args.Get(0).(storage.ResultSet[*flipt.Segment]), args.Error(1)
}

func (m *storeMock) CountSegments(ctx context.Context, namespaceKey string) (uint64, error) {
	args := m.Called(ctx, namespaceKey)
	return args.Get(0).(uint64), args.Error(1)
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

func (m *storeMock) ListRollouts(ctx context.Context, namespaceKey string, flagKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Rollout], error) {
	args := m.Called(ctx, namespaceKey, flagKey, opts)
	return args.Get(0).(storage.ResultSet[*flipt.Rollout]), args.Error(1)
}

func (m *storeMock) CountRollouts(ctx context.Context, namespaceKey string, flagKey string) (uint64, error) {
	args := m.Called(ctx, namespaceKey, flagKey)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *storeMock) GetRollout(ctx context.Context, namespaceKey string, key string) (*flipt.Rollout, error) {
	args := m.Called(ctx, namespaceKey, key)
	return args.Get(0).(*flipt.Rollout), args.Error(1)
}

func (m *storeMock) CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (*flipt.Rollout, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Rollout), args.Error(1)
}

func (m *storeMock) UpdateRollout(ctx context.Context, r *flipt.UpdateRolloutRequest) (*flipt.Rollout, error) {
	args := m.Called(ctx, r)
	return args.Get(0).(*flipt.Rollout), args.Error(1)
}

func (m *storeMock) DeleteRollout(ctx context.Context, r *flipt.DeleteRolloutRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *storeMock) OrderRollouts(ctx context.Context, r *flipt.OrderRolloutsRequest) error {
	args := m.Called(ctx, r)
	return args.Error(0)
}

func (m *storeMock) GetRule(ctx context.Context, namespaceKey string, id string) (*flipt.Rule, error) {
	args := m.Called(ctx, namespaceKey, id)
	return args.Get(0).(*flipt.Rule), args.Error(1)
}

func (m *storeMock) ListRules(ctx context.Context, namespaceKey string, flagKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Rule], error) {
	args := m.Called(ctx, namespaceKey, flagKey, opts)
	return args.Get(0).(storage.ResultSet[*flipt.Rule]), args.Error(1)
}

func (m *storeMock) CountRules(ctx context.Context, namespaceKey, flagKey string) (uint64, error) {
	args := m.Called(ctx, namespaceKey)
	return args.Get(0).(uint64), args.Error(1)
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

func (m *storeMock) GetEvaluationRules(ctx context.Context, namespaceKey string, flagKey string) ([]*storage.EvaluationRule, error) {
	args := m.Called(ctx, namespaceKey, flagKey)
	return args.Get(0).([]*storage.EvaluationRule), args.Error(1)
}

func (m *storeMock) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	args := m.Called(ctx, ruleID)
	return args.Get(0).([]*storage.EvaluationDistribution), args.Error(1)
}

var _ storageauth.Store = &authStoreMock{}

type authStoreMock struct {
	mock.Mock
}

func (a *authStoreMock) CreateAuthentication(ctx context.Context, r *storageauth.CreateAuthenticationRequest) (string, *auth.Authentication, error) {
	args := a.Called(ctx, r)
	return args.String(0), args.Get(1).(*auth.Authentication), args.Error(2)
}

func (a *authStoreMock) GetAuthenticationByClientToken(ctx context.Context, clientToken string) (*auth.Authentication, error) {
	return nil, nil
}

func (a *authStoreMock) GetAuthenticationByID(ctx context.Context, id string) (*auth.Authentication, error) {
	return nil, nil
}

func (a *authStoreMock) ListAuthentications(ctx context.Context, r *storage.ListRequest[storageauth.ListAuthenticationsPredicate]) (set storage.ResultSet[*auth.Authentication], err error) {
	return set, err
}

func (a *authStoreMock) DeleteAuthentications(ctx context.Context, r *storageauth.DeleteAuthenticationsRequest) error {
	args := a.Called(ctx, r)
	return args.Error(0)
}

func (a *authStoreMock) ExpireAuthenticationByID(ctx context.Context, id string, expireAt *timestamppb.Timestamp) error {
	return nil
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

type auditSinkSpy struct {
	sendAuditsCalled int
	events           []audit.Event
	fmt.Stringer
}

func (a *auditSinkSpy) SendAudits(es []audit.Event) error {
	a.sendAuditsCalled++
	a.events = append(a.events, es...)
	return nil
}

func (a *auditSinkSpy) String() string {
	return "auditSinkSpy"
}

func (a *auditSinkSpy) Close() error { return nil }

type auditExporterSpy struct {
	audit.EventExporter
	sinkSpy *auditSinkSpy
}

func newAuditExporterSpy(logger *zap.Logger) *auditExporterSpy {
	aspy := &auditSinkSpy{events: make([]audit.Event, 0)}
	as := []audit.Sink{aspy}

	return &auditExporterSpy{
		EventExporter: audit.NewSinkSpanExporter(logger, as),
		sinkSpy:       aspy,
	}
}

func (a *auditExporterSpy) GetSendAuditsCalled() int {
	return a.sinkSpy.sendAuditsCalled
}

func (a *auditExporterSpy) GetEvents() []audit.Event {
	return a.sinkSpy.events
}
