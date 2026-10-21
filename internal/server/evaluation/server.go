package evaluation

import (
	"context"

	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/server/authz"
	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// ResolveEnvironment returns the environment identified by key.
//
// When the key is blank or unknown, it falls back to the environment from the
// X-Flipt-Environment header, or the default environment when there is no
// header, for backwards compatibility with v1 clients which don't set an
// environment key.
//
// When authorization is required, an unknown key is not found instead.
// Authorization checked the requested key, so serving the header or default
// environment would serve an environment that was never authorized. The
// "default" key still resolves to the default environment.
func ResolveEnvironment(ctx context.Context, store EnvironmentStore, key string) (environments.Environment, error) {
	env, err := store.Get(ctx, key)
	if err == nil {
		return env, nil
	}

	if key != "" && authz.IsAuthorizationRequired(ctx) {
		return store.GetFromContext(common.WithFliptEnvironment(ctx, key))
	}

	return store.GetFromContext(ctx)
}

// EnvironmentStore is the minimal abstraction for interacting with the storage layer for evaluation.
type EnvironmentStore interface {
	GetFromContext(context.Context) (environments.Environment, error)
	Get(context.Context, string) (environments.Environment, error)
}

// Server serves the Flipt evaluate v2 gRPC Server.
type Server struct {
	logger         *zap.Logger
	store          EnvironmentStore
	tracingEnabled bool
	metricsEnabled bool
	// authzEnabled opts into authorization for evaluation endpoints.
	// It mirrors !AuthorizationConfig.Exclude.Evaluation and defaults to false
	// to preserve the historical skip-authorization behavior.
	authzEnabled bool
	evaluation.UnimplementedEvaluationServiceServer
}

// Option is a functional option for configuring the Server.
type Option func(*Server)

// WithTracing enables telemetry for the evaluation server.
func WithTracing(enabled bool) Option {
	return func(s *Server) {
		s.tracingEnabled = enabled
	}
}

// WithMetrics enables metrics for the evaluation server.
func WithMetrics(enabled bool) Option {
	return func(s *Server) {
		s.metricsEnabled = enabled
	}
}

// WithAuthorizationEnabled opts into authorization for evaluation endpoints.
func WithAuthorizationEnabled(enabled bool) Option {
	return func(s *Server) {
		s.authzEnabled = enabled
	}
}

// New is constructs a new Server.
func New(logger *zap.Logger, store EnvironmentStore, ops ...Option) *Server {
	s := &Server{
		logger: logger,
		store:  store,
	}

	for _, o := range ops {
		o(s)
	}

	return s
}

// RegisterGRPC registers the EvaluateServer onto the provided gRPC Server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	evaluation.RegisterEvaluationServiceServer(server, s)
}

func (s *Server) SkipsAuthorization(ctx context.Context) bool {
	return !s.authzEnabled
}
