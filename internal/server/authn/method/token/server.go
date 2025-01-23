package token

import (
	"context"
	"fmt"

	storageauth "go.flipt.io/flipt/internal/storage/authn"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

const (
	storageMetadataNameKey        = "io.flipt.auth.token.name"
	storageMetadataDescriptionKey = "io.flipt.auth.token.description"
	storageMetadataNamespaceKey   = "io.flipt.auth.token.namespace"
)

// Server is an implementation of auth.AuthenticationMethodTokenServiceServer
//
// It is used to create static tokens within the backing AuthenticationStore.
type Server struct {
	logger *zap.Logger
	store  storageauth.Store
	auth.UnimplementedAuthenticationMethodTokenServiceServer
}

// NewServer constructs and configures a new *Server.
func NewServer(logger *zap.Logger, store storageauth.Store) *Server {
	return &Server{
		logger: logger,
		store:  store,
	}
}

// RegisterGRPC registers the server as an Server on the provided grpc server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	auth.RegisterAuthenticationMethodTokenServiceServer(server, s)
}

func (s *Server) AllowsNamespaceScopedAuthentication(ctx context.Context) bool {
	return false
}

// CreateToken adapts and delegates the token request to the backing AuthenticationStore.
//
// Implicitly, the Authentication created will be of type auth.Method_TOKEN.
// Name, Description, and NamespaceKey are both stored in Authentication.Metadata.
// Given the token is created successfully, the generate clientToken string is returned.
// Along with the created Authentication, which includes it's identifier and associated timestamps.
func (s *Server) CreateToken(ctx context.Context, req *auth.CreateTokenRequest) (*auth.CreateTokenResponse, error) {
	metadata := req.Metadata
	if metadata == nil {
		metadata = map[string]string{}
	}

	metadata[storageMetadataNameKey] = req.GetName()
	if req.GetDescription() != "" {
		metadata[storageMetadataDescriptionKey] = req.GetDescription()
	}

	if req.GetNamespaceKey() != "" {
		metadata[storageMetadataNamespaceKey] = req.GetNamespaceKey()
	}

	clientToken, authentication, err := s.store.CreateAuthentication(ctx, &storageauth.CreateAuthenticationRequest{
		Method:    auth.Method_METHOD_TOKEN,
		ExpiresAt: req.ExpiresAt,
		Metadata:  metadata,
	})
	if err != nil {
		return nil, fmt.Errorf("attempting to create token: %w", err)
	}

	return &auth.CreateTokenResponse{
		ClientToken:    clientToken,
		Authentication: authentication,
	}, nil
}
