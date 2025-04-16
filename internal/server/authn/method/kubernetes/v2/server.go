package v2

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Server implements the Kubernetes authentication method server
type Server struct {
	logger        *zap.Logger
	authenticator Authenticator
	config        config.AuthenticationConfig

	auth.UnimplementedAuthenticationMethodKubernetesServiceServer
}

// NewServer creates a new Server instance
func NewServer(logger *zap.Logger, config config.AuthenticationConfig) (*Server, error) {
	authenticator, err := New(logger, config.Methods.Kubernetes.Method)
	if err != nil {
		return nil, fmt.Errorf("creating kubernetes authenticator: %w", err)
	}

	return &Server{
		logger:        logger,
		authenticator: authenticator,
		config:        config,
	}, nil
}

// RegisterGRPC registers the server with the gRPC server
func (s *Server) RegisterGRPC(srv *grpc.Server) {
	auth.RegisterAuthenticationMethodKubernetesServiceServer(srv, s)
}

// SkipsAuthentication returns true as this endpoint needs to be accessible without authentication
func (s *Server) SkipsAuthentication(ctx context.Context) bool {
	return true
}

// VerifyServiceAccount implements v1 compatibility by returning the service account token as the client token
func (s *Server) VerifyServiceAccount(ctx context.Context, req *auth.VerifyServiceAccountRequest) (*auth.VerifyServiceAccountResponse, error) {
	// Verify and authenticate the token
	authentication, err := s.authenticator.Authenticate(ctx, req.ServiceAccountToken)
	if err != nil {
		return nil, fmt.Errorf("authenticating service account token: %w", err)
	}

	s.logger.Debug("authentication", zap.Any("authentication", authentication))

	// For v1 compatibility, return the same token as client_token
	return &auth.VerifyServiceAccountResponse{
		ClientToken:    req.ServiceAccountToken,
		Authentication: authentication,
	}, nil
}
