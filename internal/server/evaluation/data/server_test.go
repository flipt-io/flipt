package data

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/metadata"
)

func TestEvaluationSnapshotNamespace(t *testing.T) {
	var (
		store  = &evaluationStoreMock{}
		logger = zaptest.NewLogger(t)
		s      = New(logger, store)
	)

	t.Run("If-None-Match header match", func(t *testing.T) {
		ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs("GrpcGateway-If-None-Match", "92e200311a56800b3e475bf2d2442724535e87bf"))

		store.On("GetVersion", mock.Anything, mock.Anything).Return("etag", nil)

		resp, err := s.EvaluationSnapshotNamespace(ctx, &evaluation.EvaluationNamespaceSnapshotRequest{
			Key: "namespace",
		})

		require.NoError(t, err)
		assert.Nil(t, resp)

		store.AssertExpectations(t)
	})
}
