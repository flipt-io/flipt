package evaluation

import (
	"context"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Storer is the minimal abstraction for interacting with the storage layer for evaluation.
type Storer interface {
	GetFlag(ctx context.Context, flag storage.ResourceRequest) (*flipt.Flag, error)
	GetEvaluationRules(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRule, error)
	GetEvaluationDistributions(ctx context.Context, ruleID storage.IDRequest) ([]*storage.EvaluationDistribution, error)
	GetEvaluationRollouts(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRollout, error)
}

// Server serves the Flipt evaluate v2 gRPC Server.
type Server struct {
	logger    *zap.Logger
	store     Storer
	evaluator *Evaluator
	evaluation.UnimplementedEvaluationServiceServer
}

// New is constructs a new Server.
func New(logger *zap.Logger, store Storer) *Server {
	return &Server{
		logger:    logger,
		store:     store,
		evaluator: NewEvaluator(logger, store),
	}
}

// RegisterGRPC registers the EvaluateServer onto the provided gRPC Server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	evaluation.RegisterEvaluationServiceServer(server, s)
}

func (s *Server) AllowsNamespaceScopedAuthentication(ctx context.Context) bool {
	return true
}

func (s *Server) SkipsAuthorization(ctx context.Context) bool {
	return true
}
