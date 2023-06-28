package evaluation

import (
	"go.flipt.io/flipt/internal/storage"
	rpcEvalution "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// EvaluateServer serves the Flipt evaluate v2 gRPC Server.
type EvaluateServer struct {
	logger    *zap.Logger
	store     storage.Store
	evaluator MultiVariateEvaluator
	rpcEvalution.UnimplementedEvaluationServiceServer
}

// NewEvaluateServer is constructs a new EvaluateServer.
func NewEvaluateServer(logger *zap.Logger, store storage.Store) *EvaluateServer {
	return &EvaluateServer{
		logger:    logger,
		store:     store,
		evaluator: NewEvaluator(logger, store),
	}
}

// RegisterGRPC registers the EvaluateServer onto the provided gRPC Server.
func (e *EvaluateServer) RegisterGRPC(server *grpc.Server) {
	rpcEvalution.RegisterEvaluationServiceServer(server, e)
}
