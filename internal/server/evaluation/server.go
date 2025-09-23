package evaluation

import (
	"context"

	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

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
	return true
}
