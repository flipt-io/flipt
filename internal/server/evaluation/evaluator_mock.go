package evaluation

import (
	"context"

	"github.com/stretchr/testify/mock"
	"go.flipt.io/flipt/rpc/flipt"
)

type evaluatorMock struct {
	mock.Mock
}

func (e *evaluatorMock) String() string {
	return "mock"
}

func (e *evaluatorMock) Evaluate(ctx context.Context, er *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
	args := e.Called(ctx, er)
	return args.Get(0).(*flipt.EvaluationResponse), args.Error(1)
}
