package server

import (
	"context"

	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var _ flipt.FliptServer = &Server{}

// EnvironmentStore is the minimal abstraction for interacting with the storage layer for evaluation.
type EnvironmentStore interface {
	GetFromContext(context.Context) environments.Environment
}

// Server serves the Flipt backend
type Server struct {
	logger *zap.Logger
	store  EnvironmentStore
	flipt.UnimplementedFliptServer
}

// New creates a new Server
func New(logger *zap.Logger, store EnvironmentStore) *Server {
	return &Server{
		logger: logger,
		store:  store,
	}
}

// RegisterGRPC registers the *Server onto the provided grpc Server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	flipt.RegisterFliptServer(server, s)
}

func (s *Server) getStore(ctx context.Context) (storage.ReadOnlyStore, error) {
	return s.store.GetFromContext(ctx).EvaluationStore()
}
