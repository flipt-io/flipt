package basic

import (
	"context"
	"crypto/subtle"
	"time"

	"go.flipt.io/flipt/internal/config"
	authmiddlewaregrpc "go.flipt.io/flipt/internal/server/auth/middleware/grpc"
	storageauth "go.flipt.io/flipt/internal/storage/auth"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	logger *zap.Logger
	store  storageauth.Store
	config config.AuthenticationConfig

	auth.UnimplementedAuthenticationMethodBasicServiceServer
}

func New(logger *zap.Logger, store storageauth.Store, cfg config.AuthenticationConfig) *Server {
	return &Server{logger: logger, store: store, config: cfg}
}

// RegisterGRPC registers the server instnace on the provided gRPC server.
func (s *Server) RegisterGRPC(srv *grpc.Server) {
	auth.RegisterAuthenticationMethodBasicServiceServer(srv, s)
}

// SkipsAuthentication returns true because the service is itself used to obtain authentication credentials
func (s *Server) SkipsAuthentication(ctx context.Context) bool {
	return true
}

// Login creates a flipt client token and associated authentication entry given
// a valid username and password is supplied which matches the stored username
// and bcrypted password.
func (s *Server) Login(ctx context.Context, r *auth.LoginRequest) (*auth.CallbackResponse, error) {
	if r.Username == "" {
		s.logger.Info("Login attempt failed with empty username")
		return nil, authmiddlewaregrpc.ErrUnauthenticated
	}

	if r.Password == "" {
		s.logger.Info("Login attempt failed with empty password")
		return nil, authmiddlewaregrpc.ErrUnauthenticated
	}

	if subtle.ConstantTimeCompare([]byte(s.config.Methods.Basic.Method.Username), []byte(r.Username)) != 1 {
		s.logger.Info("Login attempt failed with unexpected username")
		return nil, authmiddlewaregrpc.ErrUnauthenticated
	}

	if err := bcrypt.CompareHashAndPassword([]byte(s.config.Methods.Basic.Method.BCryptPassword), []byte(r.Password)); err != nil {
		s.logger.Info("Login attempt failed password comparison", zap.Error(err))
		return nil, authmiddlewaregrpc.ErrUnauthenticated
	}

	clientToken, a, err := s.store.CreateAuthentication(ctx, &storageauth.CreateAuthenticationRequest{
		Method:    auth.Method_METHOD_BASIC,
		ExpiresAt: timestamppb.New(time.Now().UTC().Add(s.config.Session.TokenLifetime)),
	})
	if err != nil {
		return nil, err
	}

	return &auth.CallbackResponse{
		ClientToken:    clientToken,
		Authentication: a,
	}, nil
}
