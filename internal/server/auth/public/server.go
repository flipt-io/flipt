package public

import (
	"context"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	logger *zap.Logger

	// response is static for the lifetime of the configuration
	// so we can compute it upfront and re-use.
	resp *auth.ListAuthenticationMethodsResponse

	auth.UnimplementedPublicAuthenticationServiceServer
}

func NewServer(logger *zap.Logger, conf config.AuthenticationConfig) *Server {
	server := &Server{
		logger: logger,
		resp:   &auth.ListAuthenticationMethodsResponse{},
	}

	for _, info := range conf.Methods.AllMethods() {
		server.resp.Methods = append(server.resp.Methods, &auth.MethodInfo{
			Method:            info.AuthenticationMethodInfo.Method,
			Enabled:           info.Enabled,
			SessionCompatible: info.AuthenticationMethodInfo.SessionCompatible,
			Metadata:          info.AuthenticationMethodInfo.Metadata,
		})
	}

	return server
}

func (s *Server) ListAuthenticationMethods(_ context.Context, _ *emptypb.Empty) (*auth.ListAuthenticationMethodsResponse, error) {
	return s.resp, nil
}

func (s *Server) RegisterGRPC(server *grpc.Server) {
	auth.RegisterPublicAuthenticationServiceServer(server, s)
}

func (s *Server) SkipsAuthentication(ctx context.Context) bool {
	return true
}
