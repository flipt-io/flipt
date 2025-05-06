package common

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/core"
	"go.flipt.io/flipt/rpc/v2/evaluation"
)

var _ storage.Store = &StoreMock{}

func NewMockStore(t interface {
	mock.TestingT
	Cleanup(func())
},
) *StoreMock {
	mock := &StoreMock{}
	mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

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

func (m *StoreMock) GetFlag(ctx context.Context, flag storage.ResourceRequest) (*core.Flag, error) {
	args := m.Called(ctx, flag)
	return args.Get(0).(*core.Flag), args.Error(1)
}

func (m *StoreMock) ListFlags(ctx context.Context, req *storage.ListRequest[storage.NamespaceRequest]) (storage.ResultSet[*core.Flag], error) {
	args := m.Called(ctx, req)
	return args.Get(0).(storage.ResultSet[*core.Flag]), args.Error(1)
}

func (m *StoreMock) CountFlags(ctx context.Context, ns storage.NamespaceRequest) (uint64, error) {
	args := m.Called(ctx, ns)
	return args.Get(0).(uint64), args.Error(1)
}

func (m *StoreMock) GetEvaluationRules(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRule, error) {
	args := m.Called(ctx, flag)
	return args.Get(0).([]*storage.EvaluationRule), args.Error(1)
}

func (m *StoreMock) GetEvaluationDistributions(ctx context.Context, r storage.ResourceRequest, rule storage.IDRequest) ([]*storage.EvaluationDistribution, error) {
	args := m.Called(ctx, rule)
	return args.Get(0).([]*storage.EvaluationDistribution), args.Error(1)
}

func (m *StoreMock) GetEvaluationRollouts(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRollout, error) {
	args := m.Called(ctx, flag)
	return args.Get(0).([]*storage.EvaluationRollout), args.Error(1)
}

func (m *StoreMock) EvaluationNamespaceSnapshot(ctx context.Context, ns string) (*evaluation.EvaluationNamespaceSnapshot, error) {
	args := m.Called(ctx, ns)
	return args.Get(0).(*evaluation.EvaluationNamespaceSnapshot), args.Error(1)
}
