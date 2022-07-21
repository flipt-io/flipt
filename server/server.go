package server

import (
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/storage"

	"github.com/sirupsen/logrus"
)

var _ flipt.FliptServer = &Server{}

// Server serves the Flipt backend
type Server struct {
	logger logrus.FieldLogger
	store  storage.Store
	flipt.UnimplementedFliptServer
}

// New creates a new Server
func New(logger logrus.FieldLogger, store storage.Store) *Server {
	return &Server{
		logger: logger,
		store:  store,
	}
}
