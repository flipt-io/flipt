package evaluation

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
)

var _ EvaluationStorer = &evaluationStoreMock{}

type evaluationStoreMock struct {
	mock.Mock
}

func (e *evaluationStoreMock) String() string {
	return "mock"
}

func (e *evaluationStoreMock) GetFlag(ctx context.Context, namespaceKey, key string) (*flipt.Flag, error) {
	args := e.Called(ctx, namespaceKey, key)
	return args.Get(0).(*flipt.Flag), args.Error(1)
}

func (e *evaluationStoreMock) GetEvaluationRules(ctx context.Context, namespaceKey string, flagKey string) ([]*storage.EvaluationRule, error) {
	args := e.Called(ctx, namespaceKey, flagKey)
	return args.Get(0).([]*storage.EvaluationRule), args.Error(1)
}

func (e *evaluationStoreMock) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	args := e.Called(ctx, ruleID)
	return args.Get(0).([]*storage.EvaluationDistribution), args.Error(1)
}
