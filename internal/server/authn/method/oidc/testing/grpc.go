package testing

import (
	"context"
	"net"
	"testing"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/server/authn/method/oidc"
	middleware "go.flipt.io/flipt/internal/server/middleware/grpc"
	"go.flipt.io/flipt/internal/storage/authn/memory"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type GRPCServer struct {
	*grpc.Server

	ClientConn *grpc.ClientConn
	Store      *memory.Store

	errc chan error
}

func (s *GRPCServer) Client() auth.AuthenticationMethodOIDCServiceClient {
	return auth.NewAuthenticationMethodOIDCServiceClient(s.ClientConn)
}

func StartGRPCServer(t *testing.T, ctx context.Context, logger *zap.Logger, conf config.AuthenticationConfig) *GRPCServer {
	t.Helper()

	var (
		store    = memory.NewStore()
		listener = bufconn.Listen(1024 * 1024)
		server   = grpc.NewServer(
			grpc_middleware.WithUnaryServerChain(
				middleware.ErrorUnaryInterceptor,
			),
		)
		grpcServer = &GRPCServer{
			Server: server,
			Store:  store,
			errc:   make(chan error, 1),
		}
	)

	auth.RegisterAuthenticationMethodOIDCServiceServer(server, oidc.NewServer(logger, store, conf))

	go func() {
		defer close(grpcServer.errc)
		grpcServer.errc <- server.Serve(listener)
	}()

	var (
		err    error
		dialer = func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}
	)

	grpcServer.ClientConn, err = grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer))
	require.NoError(t, err)

	return grpcServer
}

func (s *GRPCServer) Stop() error {
	if err := s.ClientConn.Close(); err != nil {
		return err
	}

	s.Server.Stop()

	return <-s.errc
}
