package public

import (
	"context"
	"path"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/authn/method"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Server struct {
	logger *zap.Logger

	conf config.AuthenticationConfig

	auth.UnimplementedPublicAuthenticationServiceServer
}

func NewServer(logger *zap.Logger, conf config.AuthenticationConfig) *Server {
	server := &Server{
		logger: logger,
		conf:   conf,
	}

	return server
}

func (s *Server) ListAuthenticationMethods(ctx context.Context, _ *emptypb.Empty) (*auth.ListAuthenticationMethodsResponse, error) {
	var prefix string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if forwardPrefix := md.Get(method.ForwardedPrefixKey); len(forwardPrefix) > 0 {
			prefix = path.Join(forwardPrefix...)
		}
	}

	resp := &auth.ListAuthenticationMethodsResponse{}
	for _, info := range s.conf.Methods.AllMethods(config.WithForwardPrefix(ctx, prefix)) {
		resp.Methods = append(resp.Methods, &auth.MethodInfo{
			Method:            info.AuthenticationMethodInfo.Method,
			Enabled:           info.Enabled,
			SessionCompatible: info.AuthenticationMethodInfo.SessionCompatible,
			Metadata:          info.AuthenticationMethodInfo.Metadata,
		})
	}

	return resp, nil
}

func (s *Server) RegisterGRPC(server *grpc.Server) {
	auth.RegisterPublicAuthenticationServiceServer(server, s)
}

func (s *Server) SkipsAuthentication(ctx context.Context) bool {
	return true
}
