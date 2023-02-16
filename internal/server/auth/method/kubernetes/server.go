package kubernetes

import (
	"context"

	"go.flipt.io/flipt/internal/config"
	storageauth "go.flipt.io/flipt/internal/storage/auth"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
)

type serviceAccount struct {
	Name string `json:"name"`
}

type serviceAccountVerifier interface {
	verify(jwt string) (serviceAccount, error)
}

// Server is the core server-side implementation of the "kubernetes" authentication method.
type Server struct {
	logger *zap.Logger
	store  storageauth.Store
	config config.AuthenticationConfig

	verifier serviceAccountVerifier

	auth.UnimplementedAuthenticationMethodKubernetesServiceServer
}

// NewServer constructs a new Server instance based on the provided logger, store and configuration.
func NewServer(logger *zap.Logger, store storageauth.Store, config config.AuthenticationConfig) *Server {
	return &Server{
		logger: logger,
		store:  store,
		config: config,
	}
}

func (s *Server) VerifyServiceAccount(ctx context.Context, req *auth.VerifyServiceAccountRequest) (*auth.VerifyServiceAccountResponse, error) {
	return &auth.VerifyServiceAccountResponse{}, nil
}
