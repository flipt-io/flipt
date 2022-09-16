package server

import (
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/storage"
	"go.uber.org/zap"
)

var _ flipt.FliptServer = &Server{}

const (
	defaultListLimit = 20
	maxListLimit     = 50
)

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
