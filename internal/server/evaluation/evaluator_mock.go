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

func (e *evaluatorMock) Evaluate(ctx context.Context, flag *flipt.Flag, er *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
	args := e.Called(ctx, flag, er)
	return args.Get(0).(*flipt.EvaluationResponse), args.Error(1)
}
