package server

import (
	"context"

	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var _ flipt.FliptServer = &Server{}

// Server serves the Flipt backend
type Server struct {
	logger *zap.Logger
	store  storage.Store
	flipt.UnimplementedFliptServer
}

// New creates a new Server
func New(logger *zap.Logger, store storage.Store) *Server {
	return &Server{
		logger: logger,
		store:  store,
	}
}

// RegisterGRPC registers the *Server onto the provided grpc Server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	flipt.RegisterFliptServer(server, s)
}

func (s *Server) AllowsNamespaceScopedAuthentication(ctx context.Context) bool {
	return true
}
