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
	GetFromContext(context.Context) (environments.Environment, error)
}

// Server serves the Flipt backend
type Server struct {
	logger              *zap.Logger
	store               EnvironmentStore
	includeFlagMetadata bool
	flipt.UnimplementedFliptServer
}

// Option is a functional option for configuring the Server
type Option func(*Server)

// WithFlagMetadata configures whether flag metadata should be included in responses
func WithFlagMetadata(include bool) Option {
	return func(s *Server) {
		s.includeFlagMetadata = include
	}
}

// New creates a new Server
func New(logger *zap.Logger, store EnvironmentStore, opts ...Option) *Server {
	s := &Server{
		logger: logger,
		store:  store,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// RegisterGRPC registers the *Server onto the provided grpc Server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	flipt.RegisterFliptServer(server, s)
}

func (s *Server) getStore(ctx context.Context) (storage.ReadOnlyStore, error) {
	env, err := s.store.GetFromContext(ctx)
	if err != nil {
		return nil, err
	}
	return env.EvaluationStore()
}
