package testing

import (
	"context"
	"net"
	"testing"

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

// GRPCServer wraps a gRPC server with an in-memory store for OIDC testing.
type GRPCServer struct {
	*grpc.Server

	ClientConn *grpc.ClientConn
	Store      *memory.Store

	errc chan error
}

// Client returns an OIDC service client connected to this test server.
func (s *GRPCServer) Client() auth.AuthenticationMethodOIDCServiceClient {
	return auth.NewAuthenticationMethodOIDCServiceClient(s.ClientConn)
}

// StartGRPCServer starts an in-memory gRPC test server with the given config and options.
func StartGRPCServer(t *testing.T, ctx context.Context, logger *zap.Logger, conf config.AuthenticationConfig, opts ...oidc.Option) *GRPCServer {
	t.Helper()

	var (
		store    = memory.NewStore(logger)
		listener = bufconn.Listen(1024 * 1024)
		server   = grpc.NewServer(
			grpc.ChainUnaryInterceptor(middleware.ErrorUnaryInterceptor),
		)
		grpcServer = &GRPCServer{
			Server: server,
			Store:  store,
			errc:   make(chan error, 1),
		}
	)

	oidcServer := oidc.NewServer(logger, store, oidc.NewRegistry(conf), conf, opts...)
	auth.RegisterAuthenticationMethodOIDCServiceServer(server, oidcServer)

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

// Stop shuts down the gRPC test server and returns any serve error.
func (s *GRPCServer) Stop() error {
	if err := s.ClientConn.Close(); err != nil {
		return err
	}

	s.Server.Stop()

	return <-s.errc
}
