package evaluation

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt"
)

func TestEvaluationRequest_Request(t *testing.T) {
	req := &EvaluationRequest{
		EnvironmentKey: "production",
		NamespaceKey:   "team-a",
	}

	requests := req.Request()
	require.Len(t, requests, 1)
	assert.Equal(t, flipt.NewRequest(
		flipt.ScopeNamespace,
		flipt.ActionEvaluate,
		flipt.WithEnvironment("production"),
		flipt.WithNamespace("team-a"),
	), requests[0])
}

func TestEvaluationRequest_Request_Defaults(t *testing.T) {
	requests := (&EvaluationRequest{}).Request()
	require.Len(t, requests, 1)
	assert.Equal(t, flipt.NewRequest(flipt.ScopeNamespace, flipt.ActionEvaluate), requests[0])
}

func TestBatchEvaluationRequest_Request(t *testing.T) {
	req := &BatchEvaluationRequest{
		Requests: []*EvaluationRequest{
			{EnvironmentKey: "production", NamespaceKey: "team-a"},
			{EnvironmentKey: "production", NamespaceKey: "team-b"},
		},
	}

	requests := req.Request()
	require.Len(t, requests, 2)
	assert.Equal(t, "team-a", *requests[0].Namespace)
	assert.Equal(t, "team-b", *requests[1].Namespace)
	for _, r := range requests {
		assert.Equal(t, flipt.ScopeNamespace, r.Scope)
		assert.Equal(t, flipt.ActionEvaluate, r.Action)
		assert.Equal(t, "production", *r.Environment)
	}
}
