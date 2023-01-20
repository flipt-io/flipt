package auth

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/internal/storage"
	storageauth "go.flipt.io/flipt/internal/storage/auth"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var _ auth.AuthenticationServiceServer = &Server{}

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

// RegisterGRPC registers the server as an Server on the provided grpc server.
func (s *Server) RegisterGRPC(server *grpc.Server) {
	auth.RegisterAuthenticationServiceServer(server, s)
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

// ListAuthentications produces a set of authentications for the provided method filter and pagination parameters.
func (s *Server) ListAuthentications(ctx context.Context, r *auth.ListAuthenticationsRequest) (*auth.ListAuthenticationsResponse, error) {
	req := &storage.ListRequest[storageauth.ListAuthenticationsPredicate]{
		QueryParams: storage.QueryParams{
			Limit:     uint64(r.Limit),
			PageToken: r.PageToken,
		},
	}

	if r.Method != auth.Method_METHOD_NONE {
		req.Predicate.Method = &r.Method
	}

	results, err := s.store.ListAuthentications(ctx, req)
	if err != nil {
		s.logger.Error("listing authentication", zap.Error(err))

		return nil, fmt.Errorf("listing authentications: %w", err)
	}

	return &auth.ListAuthenticationsResponse{
		Authentications: results.Results,
		NextPageToken:   results.NextPageToken,
	}, nil
}

// DeleteAuthentication deletes the authentication with the supplied ID.
func (s *Server) DeleteAuthentication(ctx context.Context, req *auth.DeleteAuthenticationRequest) (*emptypb.Empty, error) {
	s.logger.Debug("DeleteAuthentication", zap.String("id", req.Id))

	return &emptypb.Empty{}, s.store.DeleteAuthentications(ctx, storageauth.Delete(storageauth.WithID(req.Id)))
}

// ExpireAuthenticationSelf expires the Authentication which was derived from the request context.
// If no expire_at is provided, the current time is used. This is useful for logging out a user.
// If the expire_at is greater than the current expiry time, the expiry time is extended.
func (s *Server) ExpireAuthenticationSelf(ctx context.Context, req *auth.ExpireAuthenticationSelfRequest) (*emptypb.Empty, error) {
	if auth := GetAuthenticationFrom(ctx); auth != nil {
		s.logger.Debug("ExpireAuthentication", zap.String("id", auth.Id))

		if req.ExpiresAt == nil || !req.ExpiresAt.IsValid() {
			req.ExpiresAt = timestamppb.Now()
		}

		return &emptypb.Empty{}, s.store.ExpireAuthenticationByID(ctx, auth.Id, req.ExpiresAt)
	}

	return nil, errUnauthenticated
}
