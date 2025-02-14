package evaluation

import (
	"context"

	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// EnvironmentStore is the minimal abstraction for interacting with the storage layer for evaluation.
type EnvironmentStore interface {
	GetDefault(context.Context) environments.Environment
}

// Server serves the Flipt evaluate v2 gRPC Server.
type Server struct {
	logger *zap.Logger
	store  EnvironmentStore
	evaluation.UnimplementedEvaluationServiceServer
}

// New is constructs a new Server.
func New(logger *zap.Logger, store EnvironmentStore) *Server {
	return &Server{
		logger: logger,
		store:  store,
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

// getEvalStore returns the relevant instance of storage used to fetch data for evaluations.
func (s *Server) getEvalStore(ctx context.Context) (storage.ReadOnlyStore, error) {
	// TODO(georgemac): update this to support overriding default environment based
	// on configuration or header metadata on the request
	return s.store.GetDefault(ctx).EvaluationStore()
}
