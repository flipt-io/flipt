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
	authkubernetes "go.flipt.io/flipt/internal/server/auth/method/kubernetes"
	authoidc "go.flipt.io/flipt/internal/server/auth/method/oidc"
	authtoken "go.flipt.io/flipt/internal/server/auth/method/token"
	"go.flipt.io/flipt/internal/server/auth/public"
	storageauth "go.flipt.io/flipt/internal/storage/auth"
	authsql "go.flipt.io/flipt/internal/storage/auth/sql"
	oplocksql "go.flipt.io/flipt/internal/storage/oplock/sql"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func authenticationGRPC(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.Config,
	auditLoggingEnabled bool,
) (grpcRegisterers, []grpc.UnaryServerInterceptor, func(context.Context) error, error) {
	db, driver, shutdown, err := getDB(ctx, logger, cfg)
	if err != nil {
		return nil, nil, shutdown, err
	}

	var (
		sqlBuilder = fliptsql.BuilderFor(db, driver)
		store      = authsql.NewStore(driver, sqlBuilder, logger)
		oplock     = oplocksql.New(logger, driver, sqlBuilder)
	)

	var (
		public   = public.NewServer(logger, cfg.Authentication)
		register = grpcRegisterers{
			public,
			auth.NewServer(logger, store, auth.WithAuditLoggingEnabled(auditLoggingEnabled)),
		}
		authOpts = []containers.Option[auth.InterceptorOptions]{
			auth.WithServerSkipsAuthentication(public),
		}
		interceptors []grpc.UnaryServerInterceptor
	)

	// register auth method token service
	if cfg := cfg.Authentication.Methods.Token; cfg.Enabled {
		opts := []storageauth.BootstrapOption{}

		// if a bootstrap token is provided, use it
		if cfg.Method.Bootstrap.Token != "" {
			opts = append(opts, storageauth.WithToken(cfg.Method.Bootstrap.Token))
		}

		// if a bootstrap expiration is provided, use it
		if cfg.Method.Bootstrap.Expiration != 0 {
			opts = append(opts, storageauth.WithExpiration(cfg.Method.Bootstrap.Expiration))
		}

		// attempt to bootstrap authentication store
		clientToken, err := storageauth.Bootstrap(ctx, store, opts...)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("configuring token authentication: %w", err)
		}

		if clientToken != "" {
			logger.Info("access token created", zap.String("client_token", clientToken))
		}

		register.Add(authtoken.NewServer(logger, store))

		logger.Debug("authentication method \"token\" server registered")
	}

	// register auth method oidc service
	if cfg.Authentication.Methods.OIDC.Enabled {
		oidcServer := authoidc.NewServer(logger, store, cfg.Authentication)
		register.Add(oidcServer)
		// OIDC server exposes unauthenticated endpoints
		authOpts = append(authOpts, auth.WithServerSkipsAuthentication(oidcServer))

		logger.Debug("authentication method \"oidc\" server registered")
	}

	if cfg.Authentication.Methods.Kubernetes.Enabled {
		kubernetesServer, err := authkubernetes.New(logger, store, cfg.Authentication)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("configuring kubernetes authentication: %w", err)
		}
		register.Add(kubernetesServer)

		// OIDC server exposes unauthenticated endpoints
		authOpts = append(authOpts, auth.WithServerSkipsAuthentication(kubernetesServer))

		logger.Debug("authentication method \"kubernetes\" server registered")
	}

	// only enable enforcement middleware if authentication required
	if cfg.Authentication.Required {
		interceptors = append(interceptors, auth.UnaryInterceptor(
			logger,
			store,
			authOpts...,
		))

		logger.Info("authentication middleware enabled")
	}

	if cfg.Authentication.ShouldRunCleanup() {
		cleanupAuthService := cleanup.NewAuthenticationService(
			logger,
			oplock,
			store,
			cfg.Authentication,
		)
		cleanupAuthService.Run(ctx)

		backupShutdown := shutdown
		shutdown = func(ctx context.Context) error {
			logger.Info("shutting down authentication cleanup service...")

			if err := cleanupAuthService.Shutdown(ctx); err != nil {
				return err
			}

			return backupShutdown(ctx)
		}
	}

	return register, interceptors, shutdown, nil
}

func registerFunc(ctx context.Context, conn *grpc.ClientConn, fn func(context.Context, *runtime.ServeMux, *grpc.ClientConn) error) runtime.ServeMuxOption {
	return func(mux *runtime.ServeMux) {
		if err := fn(ctx, mux, conn); err != nil {
			panic(err)
		}
	}
}

func authenticationHTTPMount(
	ctx context.Context,
	logger *zap.Logger,
	cfg config.AuthenticationConfig,
	r chi.Router,
	conn *grpc.ClientConn,
) {
	var (
		authmiddleware = auth.NewHTTPMiddleware(cfg.Session)
		middleware     = []func(next http.Handler) http.Handler{authmiddleware.Handler}
		muxOpts        = []runtime.ServeMuxOption{
			registerFunc(ctx, conn, rpcauth.RegisterPublicAuthenticationServiceHandler),
			registerFunc(ctx, conn, rpcauth.RegisterAuthenticationServiceHandler),
			runtime.WithErrorHandler(authmiddleware.ErrorHandler),
		}
	)

	if cfg.Methods.Token.Enabled {
		muxOpts = append(muxOpts, registerFunc(ctx, conn, rpcauth.RegisterAuthenticationMethodTokenServiceHandler))
	}

	if cfg.Methods.OIDC.Enabled {
		oidcmiddleware := authoidc.NewHTTPMiddleware(cfg.Session)
		muxOpts = append(muxOpts,
			runtime.WithMetadata(authoidc.ForwardCookies),
			runtime.WithForwardResponseOption(oidcmiddleware.ForwardResponseOption),
			registerFunc(ctx, conn, rpcauth.RegisterAuthenticationMethodOIDCServiceHandler))

		middleware = append(middleware, oidcmiddleware.Handler)
	}

	if cfg.Methods.Kubernetes.Enabled {
		muxOpts = append(muxOpts, registerFunc(ctx, conn, rpcauth.RegisterAuthenticationMethodKubernetesServiceHandler))
	}

	r.Group(func(r chi.Router) {
		r.Use(middleware...)

		r.Mount("/auth/v1", gateway.NewGatewayServeMux(logger, muxOpts...))
	})
}
