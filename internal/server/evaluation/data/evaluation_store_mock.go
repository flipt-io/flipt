package data

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
)

var _ EvaluationStore = &evaluationStoreMock{}

type evaluationStoreMock struct {
	mock.Mock
}

func (e *evaluationStoreMock) String() string {
	return "mock"
}

func (e *evaluationStoreMock) GetVersion(ctx context.Context, ns storage.NamespaceRequest) (string, error) {
	args := e.Called(ctx, ns)
	return args.String(0), args.Error(1)
}

func (e *evaluationStoreMock) ListFlags(ctx context.Context, req *storage.ListRequest[storage.NamespaceRequest]) (storage.ResultSet[*flipt.Flag], error) {
	args := e.Called(ctx, req)
	return args.Get(0).(storage.ResultSet[*flipt.Flag]), args.Error(1)
}

func (e *evaluationStoreMock) GetFlag(ctx context.Context, flag storage.ResourceRequest) (*flipt.Flag, error) {
	args := e.Called(ctx, flag)
	return args.Get(0).(*flipt.Flag), args.Error(1)
}

func (e *evaluationStoreMock) GetEvaluationRules(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRule, error) {
	args := e.Called(ctx, flag)
	return args.Get(0).([]*storage.EvaluationRule), args.Error(1)
}

func (e *evaluationStoreMock) GetEvaluationDistributions(ctx context.Context, ruleID storage.IDRequest) ([]*storage.EvaluationDistribution, error) {
	args := e.Called(ctx, ruleID)
	return args.Get(0).([]*storage.EvaluationDistribution), args.Error(1)
}

func (e *evaluationStoreMock) GetEvaluationRollouts(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRollout, error) {
	args := e.Called(ctx, flag)
	return args.Get(0).([]*storage.EvaluationRollout), args.Error(1)
}
