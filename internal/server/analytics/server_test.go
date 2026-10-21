package analytics

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt/analytics"
	"go.uber.org/zap"
)

type testClient struct {
	req *FlagEvaluationsCountRequest
}

func (t *testClient) GetFlagEvaluationsCount(ctx context.Context, req *FlagEvaluationsCountRequest) ([]string, []float32, error) {
	t.req = req
	return []string{"2000-01-01 00:00:00", "2000-01-01 00:01:00", "2000-01-01 00:02:00"}, []float32{20.0, 30.0, 40.0}, nil
}

func (t *testClient) String() string {
	return "test_client"
}

func TestServer(t *testing.T) {
	server := New(zap.NewNop(), &testClient{})

	res, err := server.GetFlagEvaluationsCount(context.TODO(), &analytics.GetFlagEvaluationsCountRequest{
		NamespaceKey: "default",
		FlagKey:      "flag1",
		From:         "2024-01-01 00:00:00",
		To:           "2024-01-01 00:00:00",
	})
	require.NoError(t, err)

	assert.Equal(t, []string{"2000-01-01 00:00:00", "2000-01-01 00:01:00", "2000-01-01 00:02:00"}, res.Timestamps)
	assert.Equal(t, []float32{20.0, 30.0, 40.0}, res.Values)
}

func TestServerBlankNamespaceQueriesDefault(t *testing.T) {
	// authorization treats a blank namespace as the default namespace, so the
	// query must use the same namespace
	client := &testClient{}
	server := New(zap.NewNop(), client)

	_, err := server.GetFlagEvaluationsCount(t.Context(), &analytics.GetFlagEvaluationsCountRequest{
		FlagKey: "flag1",
		From:    "2024-01-01 00:00:00",
		To:      "2024-01-01 00:00:00",
	})
	require.NoError(t, err)

	require.NotNil(t, client.req)
	assert.Equal(t, "default", client.req.NamespaceKey)
	assert.Equal(t, "default", (&analytics.GetFlagEvaluationsCountRequest{}).Request()[0].Namespace)
}
