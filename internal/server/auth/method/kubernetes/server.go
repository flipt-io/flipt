package kubernetes

import (
	"context"
	"errors"

	"go.flipt.io/flipt/internal/config"
	storageauth "go.flipt.io/flipt/internal/storage/auth"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
)

type resource struct {
	Name string `json:"name"`
	UID  string `json:"uid"`
}

type claims struct {
	Namespace      string   `json:"namespace"`
	Pod            resource `json:"pod"`
	ServiceAccount resource `json:"serviceaccount"`
}

type tokenVerifier interface {
	verify(jwt string) (claims, error)
}

// Server is the core server-side implementation of the "kubernetes" authentication method.
type Server struct {
	logger *zap.Logger
	store  storageauth.Store
	config config.AuthenticationConfig

	verifier tokenVerifier

	auth.UnimplementedAuthenticationMethodKubernetesServiceServer
}

// NewServer constructs a new Server instance based on the provided logger, store and configuration.
func NewServer(logger *zap.Logger, store storageauth.Store, config config.AuthenticationConfig) *Server {
	return &Server{
		logger:   logger,
		store:    store,
		config:   config,
		verifier: noopTokenVerifier{},
	}
}

func (s *Server) VerifyServiceAccount(ctx context.Context, req *auth.VerifyServiceAccountRequest) (*auth.VerifyServiceAccountResponse, error) {
	return &auth.VerifyServiceAccountResponse{}, nil
}

type noopTokenVerifier struct{}

func (n noopTokenVerifier) verify(string) (claims, error) {
	return claims{}, errors.New("invalid token")
}
