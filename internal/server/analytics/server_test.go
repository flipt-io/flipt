package analytics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt/analytics"
	"go.uber.org/zap"
)

type testClient struct{}

func (t *testClient) GetFlagEvaluationsCount(ctx context.Context, req *analytics.GetFlagEvaluationsCountRequest) ([]string, []float32, error) {
	return []string{"2000-01-01 00:00:00", "2000-01-01 00:01:00", "2000-01-01 00:02:00"}, []float32{20.0, 30.0, 40.0}, nil
}

func TestServer(t *testing.T) {
	server := New(zap.NewNop(), &testClient{})

	res, err := server.GetFlagEvaluationsCount(context.TODO(), &analytics.GetFlagEvaluationsCountRequest{
		NamespaceKey: "default",
		FlagKey:      "flag1",
		From:         "2024-01-01 00:00:00",
		To:           "2024-01-01 00:00:00",
	})
	require.Nil(t, err)

	assert.Equal(t, res.Timestamps, []string{"2000-01-01 00:00:00", "2000-01-01 00:01:00", "2000-01-01 00:02:00"})
	assert.Equal(t, res.Values, []float32{20.0, 30.0, 40.0})
}
