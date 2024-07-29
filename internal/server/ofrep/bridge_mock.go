package ofrep

import (
	"context"

	"github.com/stretchr/testify/mock"
)

var _ Bridge = &bridgeMock{}

type bridgeMock struct {
	mock.Mock
}

func (b *bridgeMock) OFREPEvaluationBridge(ctx context.Context, input EvaluationBridgeInput) (EvaluationBridgeOutput, error) {
	args := b.Called(ctx, input)
	return args.Get(0).(EvaluationBridgeOutput), args.Error(1)
}
