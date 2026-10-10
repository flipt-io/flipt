package authz

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/rpc/flipt"
	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestInput(t *testing.T) {
	auth := &authrpc.Authentication{
		Id:     "some-id",
		Method: authrpc.Method_METHOD_TOKEN,
		Metadata: map[string]string{
			"io.flipt.auth.role": "admin",
		},
		ExpiresAt: &timestamppb.Timestamp{Seconds: 1700000000},
	}

	input := Input(flipt.NewRequest(flipt.ScopeNamespace, flipt.ActionEvaluate,
		flipt.WithEnvironment("production"),
		flipt.WithNamespace("default")), auth)

	assert.Equal(t, map[string]any{
		"request": map[string]any{
			"scope":       "namespace",
			"environment": "production",
			"namespace":   "default",
			"action":      "evaluate",
		},
		"authentication": map[string]any{
			"id":     "some-id",
			"method": 1,
			"metadata": map[string]string{
				"io.flipt.auth.role": "admin",
			},
			"expires_at": map[string]any{
				"seconds": float64(1700000000),
			},
		},
	}, input)
}

func TestRequestInput_OmitsUnsetScope(t *testing.T) {
	input := RequestInput(flipt.NewRequest(flipt.ScopeEnvironment, flipt.ActionRead,
		flipt.WithNoEnvironment(),
		flipt.WithNoNamespace()))

	assert.Equal(t, map[string]any{
		"scope":  "environment",
		"action": "read",
	}, input)
}

func TestAuthenticationInput_OmitsUnsetTimestamps(t *testing.T) {
	input := AuthenticationInput(&authrpc.Authentication{
		Method: authrpc.Method_METHOD_OIDC,
	})

	assert.Equal(t, map[string]any{
		"id":       "",
		"method":   2,
		"metadata": map[string]string(nil),
	}, input)
}

func TestInput_UsesPlainGoValues(t *testing.T) {
	input := Input(flipt.NewRequest(flipt.ScopeNamespace, flipt.ActionEvaluate,
		flipt.WithEnvironment("production"),
		flipt.WithNamespace("default")), &authrpc.Authentication{
		Method:   authrpc.Method_METHOD_TOKEN,
		Metadata: map[string]string{"io.flipt.auth.role": "admin"},
	})

	var assertPlain func(t *testing.T, v any)
	assertPlain = func(t *testing.T, v any) {
		t.Helper()

		switch v := v.(type) {
		case map[string]any:
			for _, item := range v {
				assertPlain(t, item)
			}
		case map[string]string, string, int, float64, bool, nil:
		default:
			t.Fatalf("expected plain Go value, got %T", v)
		}
	}

	assertPlain(t, input)
	require.Equal(t, "evaluate", input["request"].(map[string]any)["action"])
}
