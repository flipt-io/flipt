package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"go.flipt.io/flipt/internal/cleanup"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/gateway"
	"go.flipt.io/flipt/internal/server/auth"
	authoidc "go.flipt.io/flipt/internal/server/auth/method/oidc"
	authtoken "go.flipt.io/flipt/internal/server/auth/method/token"
	storageauth "go.flipt.io/flipt/internal/storage/auth"
	storageoplock "go.flipt.io/flipt/internal/storage/oplock"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func authenticationGRPC(
	ctx context.Context,
	logger *zap.Logger,
	cfg config.AuthenticationConfig,
	store storageauth.Store,
	oplock storageoplock.Service,
) (grpcRegisterers, []grpc.UnaryServerInterceptor, func(context.Context) error, error) {
	var (
		register = grpcRegisterers{
			auth.NewServer(logger, store),
		}
		interceptors []grpc.UnaryServerInterceptor
		shutdown     = func(context.Context) error {
			return nil
		}
	)

	// register auth method token service
	if cfg.Methods.Token.Enabled {
		// attempt to bootstrap authentication store
		clientToken, err := storageauth.Bootstrap(ctx, store)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("configuring token authentication: %w", err)
		}

		if clientToken != "" {
			logger.Info("access token created", zap.String("client_token", clientToken))
		}

		register.Add(authtoken.NewServer(logger, store))

		logger.Debug("authentication method \"token\" server registered")
	}

	var authOpts []containers.Option[auth.InterceptorOptions]
	// register auth method oidc service
	if cfg.Methods.OIDC.Enabled {
		oidcServer := authoidc.NewServer(logger, store, cfg)
		register.Add(oidcServer)
		// OIDC server exposes unauthenticated endpoints
		authOpts = append(authOpts, auth.WithServerSkipsAuthentication(oidcServer))

		logger.Debug("authentication method \"oidc\" server registered")
	}

	// only enable enforcement middleware if authentication required
	if cfg.Required {
		interceptors = append(interceptors, auth.UnaryInterceptor(
			logger,
			store,
			authOpts...,
		))

		logger.Info("authentication middleware enabled")
	}

	if cfg.ShouldRunCleanup() {
		cleanupAuthService := cleanup.NewAuthenticationService(
			logger,
			oplock,
			store,
			cfg,
		)
		cleanupAuthService.Run(ctx)

		shutdown = func(ctx context.Context) error {
			logger.Info("shutting down authentication cleanup service...")

			return cleanupAuthService.Shutdown(ctx)
		}
	}

	return register, interceptors, shutdown, nil
}

func authenticationHTTPMount(
	ctx context.Context,
	cfg config.AuthenticationConfig,
	r chi.Router,
	conn *grpc.ClientConn,
) error {
	var (
		muxOpts    []runtime.ServeMuxOption
		middleware = func(next http.Handler) http.Handler {
			return next
		}
	)

	// register OIDC middleware if method is enabled
	if cfg.Methods.OIDC.Enabled {
		oidcmiddleware := authoidc.NewHTTPMiddleware(cfg.Session)

		muxOpts = append(muxOpts,
			runtime.WithMetadata(authoidc.ForwardCookies),
			runtime.WithForwardResponseOption(oidcmiddleware.ForwardResponseOption))
		middleware = oidcmiddleware.Handler
	}

	mux := gateway.NewGatewayServeMux(muxOpts...)

	if err := rpcauth.RegisterAuthenticationServiceHandler(ctx, mux, conn); err != nil {
		return fmt.Errorf("registering auth grpc gateway: %w", err)
	}

	if cfg.Methods.Token.Enabled {
		if err := rpcauth.RegisterAuthenticationMethodTokenServiceHandler(ctx, mux, conn); err != nil {
			return fmt.Errorf("registering auth grpc gateway: %w", err)
		}
	}

	if cfg.Methods.OIDC.Enabled {
		if err := rpcauth.RegisterAuthenticationMethodOIDCServiceHandler(ctx, mux, conn); err != nil {
			return fmt.Errorf("registering auth grpc gateway: %w", err)
		}
	}

	r.Group(func(r chi.Router) {
		r.Use(middleware)

		r.Mount("/auth/v1", mux)
	})

	return nil
}
