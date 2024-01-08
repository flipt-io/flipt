package server

import (
	"context"

	"go.flipt.io/flipt/internal/server/evaluation"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var _ flipt.FliptServer = &Server{}

// MultiVariateEvaluator is an abstraction for evaluating a flag against a set of rules for multi-variate flags.
type MultiVariateEvaluator interface {
	Evaluate(ctx context.Context, flag *flipt.Flag, r *rpcevaluation.EvaluationRequest) (*flipt.EvaluationResponse, error)
}

// Server serves the Flipt backend
type Server struct {
	logger *zap.Logger
	store  storage.Store
	flipt.UnimplementedFliptServer
	evaluator MultiVariateEvaluator
}

// New creates a new Server
func New(logger *zap.Logger, store storage.Store) *Server {
	return &Server{
		logger:    logger,
		store:     store,
		evaluator: evaluation.NewEvaluator(logger, store),
	}
}

// RegisterGRPC registers the *Server onto the provided grpc Server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	flipt.RegisterFliptServer(server, s)
}

func (s *Server) AllowsNamespaceScopedAuthentication(ctx context.Context) bool {
	return true
}
