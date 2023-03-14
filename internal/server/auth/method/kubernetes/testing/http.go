package testing

import (
	"context"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/gateway"
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
		mux        = gateway.NewGatewayServeMux(logger)
		httpServer = &HTTPServer{
			GRPCServer: StartGRPCServer(t, ctx, logger, conf),
		}
	)

	err := auth.RegisterAuthenticationMethodKubernetesServiceHandler(
		ctx,
		mux,
		httpServer.GRPCServer.ClientConn,
	)
	require.NoError(t, err)

	router.Mount("/auth/v1", mux)

	return httpServer
}

func (s *HTTPServer) Stop() error {
	return s.GRPCServer.Stop()
}
