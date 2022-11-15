package auth

import (
	"context"

	storageauth "go.flipt.io/flipt/internal/storage/auth"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server is the core AuthenticationServiceServer implementations.
//
// It is the service which presents all Authentications created in the backing auth store.
type Server struct {
	logger *zap.Logger
	store  storageauth.Store

	auth.UnimplementedAuthenticationServiceServer
}

func NewServer(logger *zap.Logger, store storageauth.Store) *Server {
	return &Server{
		logger: logger,
		store:  store,
	}
}

// GetAuthenticationSelf returns the Authentication which was derived from the request context.
func (s *Server) GetAuthenticationSelf(ctx context.Context, _ *emptypb.Empty) (*auth.Authentication, error) {
	if auth := GetAuthenticationFrom(ctx); auth != nil {
		s.logger.Debug("GetAuthentication", zap.String("id", auth.Id))

		return auth, nil
	}

	return nil, errUnauthenticated
}

// GetAuthentication returns the Authentication identified by the supplied id.
func (s *Server) GetAuthentication(ctx context.Context, r *auth.GetAuthenticationRequest) (*auth.Authentication, error) {
	return s.store.GetAuthenticationByID(ctx, r.Id)
}

// DeleteAuthentication deletes the authentication with the supplied ID.
func (s *Server) DeleteAuthentication(ctx context.Context, req *auth.DeleteAuthenticationRequest) (*emptypb.Empty, error) {
	s.logger.Debug("DeleteAuthentication", zap.String("id", req.Id))

	return &emptypb.Empty{}, s.store.DeleteAuthentications(ctx, storageauth.Delete(storageauth.WithID(req.Id)))
}
