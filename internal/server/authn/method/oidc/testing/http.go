package testing

import (
	"context"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/gateway"
	"go.flipt.io/flipt/internal/server/authn/method"
	"go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
)

type HTTPServer struct {
	*GRPCServer
}

func StartHTTPServer(
	t *testing.T,
	ctx context.Context,
	logger *zap.Logger,
	conf config.AuthenticationConfig,
	router chi.Router,
) *HTTPServer {
	t.Helper()

	var (
		httpServer = &HTTPServer{
			GRPCServer: StartGRPCServer(t, ctx, logger, conf),
		}

		oidcmiddleware = method.NewHTTPMiddleware(conf.Session)
		mux            = gateway.NewGatewayServeMux(
			logger,
			runtime.WithMetadata(method.ForwardCookies),
			runtime.WithMetadata(method.ForwardPrefix),
			runtime.WithForwardResponseOption(oidcmiddleware.ForwardResponseOption),
		)
	)

	err := auth.RegisterAuthenticationMethodOIDCServiceHandler(
		ctx,
		mux,
		httpServer.GRPCServer.ClientConn,
	)
	require.NoError(t, err)

	router.Use(oidcmiddleware.Handler)
	router.Mount("/auth/v1", mux)

	return httpServer
}

func (s *HTTPServer) Stop() error {
	return s.GRPCServer.Stop()
}
