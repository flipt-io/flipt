package server

import (
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
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
