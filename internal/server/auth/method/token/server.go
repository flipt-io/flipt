package token

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
)

const (
	storageMetadataNameKey        = "io.flipt.auth.token.name"
	storageMetadataDescriptionKey = "io.flipt.auth.token.description"
)

// Server is an implementation of auth.AuthenticationMethodTokenServiceServer
//
// It is used to create static tokens within the backing AuthenticationStore.
type Server struct {
	logger *zap.Logger
	store  storage.AuthenticationStore
	auth.UnimplementedAuthenticationMethodTokenServiceServer
}

// NewServer constructs and configures a new *Server.
func NewServer(logger *zap.Logger, store storage.AuthenticationStore) *Server {
	return &Server{
		logger: logger,
		store:  store,
	}
}

// CreateToken adapts and delegates the token request to the backing AuthenticationStore.
//
// Implicitly, the Authentication created will be of type auth.Method_TOKEN.
// Name and Description are both stored in Authentication.Metadata.
// Given the token is created successfully, the generate clientToken string is returned.
// Along with the created Authentication, which includes it's identifier and associated timestamps.
func (s *Server) CreateToken(ctx context.Context, req *auth.CreateTokenRequest) (*auth.CreateTokenResponse, error) {
	clientToken, authentication, err := s.store.CreateAuthentication(ctx, &storage.CreateAuthenticationRequest{
		Method:    auth.Method_TOKEN,
		ExpiresAt: req.ExpiresAt,
		Metadata: map[string]string{
			storageMetadataNameKey:        req.GetName(),
			storageMetadataDescriptionKey: req.GetDescription(),
		},
	})
	if err != nil {
		return nil, fmt.Errorf("attempting to create token: %w", err)
	}

	return &auth.CreateTokenResponse{
		ClientToken:    clientToken,
		Authentication: authentication,
	}, nil
}
