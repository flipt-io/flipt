package grpc_middleware

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	authmiddlewaregrpc "go.flipt.io/flipt/internal/server/authn/middleware/grpc"
	enginerego "go.flipt.io/flipt/internal/server/authz/engine/rego"
	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const benchRolePolicy = `package flipt.authz.v2

import rego.v1

default allow := false

allow if {
	input.request.action == "evaluate"
	input.authentication.metadata["io.flipt.auth.role"] == "admin"
}
`

func newBenchEngine(b *testing.B) *enginerego.Engine {
	b.Helper()

	policyPath := filepath.Join(b.TempDir(), "policy.rego")
	require.NoError(b, os.WriteFile(policyPath, []byte(benchRolePolicy), 0o600))

	cfg := config.Default()
	cfg.Authorization.Required = true
	cfg.Authorization.Backend = config.AuthorizationBackendLocal
	cfg.Authorization.Local = &config.AuthorizationLocalConfig{
		Policy: &config.AuthorizationSourceLocalConfig{
			Path:         policyPath,
			PollInterval: time.Hour,
		},
	}

	engine, err := enginerego.NewEngine(b.Context(), zap.NewNop(), cfg)
	require.NoError(b, err)

	return engine
}

func benchAuth() *authrpc.Authentication {
	return &authrpc.Authentication{
		Id:     "token-id",
		Method: authrpc.Method_METHOD_TOKEN,
		Metadata: map[string]string{
			"io.flipt.auth.role": "admin",
		},
	}
}

func benchEvaluateRequest() *rpcevaluation.EvaluationRequest {
	return &rpcevaluation.EvaluationRequest{
		EnvironmentKey: "onoffinc",
		NamespaceKey:   "admin",
		FlagKey:        "auto-approval",
		EntityId:       "94beca53-f8d8-4bd9-bfd2-551736059489",
	}
}

func BenchmarkAuthorize_Evaluate_Single(b *testing.B) {
	engine := newBenchEngine(b)
	interceptor := AuthorizationRequiredInterceptor(zap.NewNop(), engine)
	info := &grpc.UnaryServerInfo{Server: &mockServer{}, FullMethod: "/flipt.evaluation.EvaluationService/Evaluate"}
	handler := func(context.Context, any) (any, error) { return nil, nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ctx := authmiddlewaregrpc.ContextWithAuthentication(b.Context(), benchAuth())
		if _, err := interceptor(ctx, benchEvaluateRequest(), info, handler); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHandlerDirect_Evaluate_Single(b *testing.B) {
	handler := func(context.Context, any) (any, error) { return nil, nil }

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := handler(b.Context(), benchEvaluateRequest()); err != nil {
			b.Fatal(err)
		}
	}
}
