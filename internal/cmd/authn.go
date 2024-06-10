package cmd

import (
	"context"
	"crypto"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/cap/jwt"
	"go.flipt.io/flipt/internal/cleanup"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/gateway"
	"go.flipt.io/flipt/internal/server/authn"
	"go.flipt.io/flipt/internal/server/authn/method"
	authgithub "go.flipt.io/flipt/internal/server/authn/method/github"
	authkubernetes "go.flipt.io/flipt/internal/server/authn/method/kubernetes"
	authoidc "go.flipt.io/flipt/internal/server/authn/method/oidc"
	authtoken "go.flipt.io/flipt/internal/server/authn/method/token"
	authmiddlewaregrpc "go.flipt.io/flipt/internal/server/authn/middleware/grpc"
	authmiddlewarehttp "go.flipt.io/flipt/internal/server/authn/middleware/http"
	"go.flipt.io/flipt/internal/server/authn/public"
	storageauth "go.flipt.io/flipt/internal/storage/authn"
	storageauthcache "go.flipt.io/flipt/internal/storage/authn/cache"
	storageauthmemory "go.flipt.io/flipt/internal/storage/authn/memory"
	authsql "go.flipt.io/flipt/internal/storage/authn/sql"
	oplocksql "go.flipt.io/flipt/internal/storage/oplock/sql"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func getAuthStore(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.Config,
	forceMigrate bool,
) (storageauth.Store, func(context.Context) error, error) {
	var (
		store    storageauth.Store = storageauthmemory.NewStore()
		shutdown                   = func(context.Context) error { return nil }
	)

	if cfg.Authentication.RequiresDatabase() {
		_, builder, driver, dbShutdown, err := getDB(ctx, logger, cfg, forceMigrate)
		if err != nil {
			return nil, nil, err
		}

		store = authsql.NewStore(driver, builder, logger)
		shutdown = dbShutdown

		if cfg.Authentication.ShouldRunCleanup() {
			var (
				oplock  = oplocksql.New(logger, driver, builder)
				cleanup = cleanup.NewAuthenticationService(
					logger,
					oplock,
					store,
					cfg.Authentication,
				)
			)

			cleanup.Run(ctx)

			dbShutdown := shutdown
			shutdown = func(ctx context.Context) error {
				logger.Info("shutting down authentication cleanup service...")

				if err := cleanup.Shutdown(ctx); err != nil {
					_ = dbShutdown(ctx)
					return err
				}

				return dbShutdown(ctx)
			}
		}
	}

	return store, shutdown, nil
}

func authenticationGRPC(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.Config,
	forceMigrate bool,
	tokenDeletedEnabled bool,
	authOpts ...containers.Option[authmiddlewaregrpc.InterceptorOptions],
) (grpcRegisterers, []grpc.UnaryServerInterceptor, func(context.Context) error, error) {

	shutdown := func(ctx context.Context) error {
		return nil
	}

	authCfg := cfg.Authentication

	// NOTE: we skip attempting to connect to any database in the situation that either the git, local, or object
	// FS backends are configured.
	// All that is required to establish a connection for authentication is to either make auth required
	// or configure at-least one authentication method (e.g. enable token method).
	if !authCfg.Enabled() && (cfg.Storage.Type != config.DatabaseStorageType) {
		return grpcRegisterers{
			public.NewServer(logger, authCfg),
			authn.NewServer(logger, storageauthmemory.NewStore()),
		}, nil, shutdown, nil
	}

	store, shutdown, err := getAuthStore(ctx, logger, cfg, forceMigrate)
	if err != nil {
		return nil, nil, nil, err
	}

	if cfg.Cache.Enabled {
		cacher, _, err := getCache(ctx, cfg)
		if err != nil {
			return nil, nil, nil, err
		}
		store = storageauthcache.NewStore(store, cacher, logger)
	}

	var (
		authServer   = authn.NewServer(logger, store, authn.WithAuditLoggingEnabled(tokenDeletedEnabled))
		publicServer = public.NewServer(logger, authCfg)

		register = grpcRegisterers{
			publicServer,
			authServer,
		}
		interceptors []grpc.UnaryServerInterceptor
	)

	// register auth method token service
	if authCfg.Methods.Token.Enabled {
		opts := []storageauth.BootstrapOption{}

		// if a bootstrap token is provided, use it
		if authCfg.Methods.Token.Method.Bootstrap.Token != "" {
			opts = append(opts, storageauth.WithToken(authCfg.Methods.Token.Method.Bootstrap.Token))
		}

		// if a bootstrap expiration is provided, use it
		if authCfg.Methods.Token.Method.Bootstrap.Expiration != 0 {
			opts = append(opts, storageauth.WithExpiration(authCfg.Methods.Token.Method.Bootstrap.Expiration))
		}

		// add any additional metadata if defined
		for k, v := range authCfg.Methods.Token.Method.Bootstrap.Metadata {
			opts = append(opts, storageauth.WithMetadataAttribute(strings.ToLower(k), v))
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
	if authCfg.Methods.OIDC.Enabled {
		oidcServer := authoidc.NewServer(logger, store, authCfg)
		register.Add(oidcServer)

		logger.Debug("authentication method \"oidc\" server registered")
	}

	if authCfg.Methods.Github.Enabled {
		githubServer := authgithub.NewServer(logger, store, authCfg)
		register.Add(githubServer)

		logger.Debug("authentication method \"github\" registered")
	}

	if authCfg.Methods.Kubernetes.Enabled {
		kubernetesServer, err := authkubernetes.New(logger, store, authCfg)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("configuring kubernetes authentication: %w", err)
		}
		register.Add(kubernetesServer)

		logger.Debug("authentication method \"kubernetes\" server registered")
	}

	// only enable enforcement middleware if authentication required
	if authCfg.Required {
		if authCfg.Methods.JWT.Enabled {
			authJWT := authCfg.Methods.JWT

			var ks jwt.KeySet

			if authJWT.Method.JWKSURL != "" {
				ks, err = jwt.NewJSONWebKeySet(ctx, authJWT.Method.JWKSURL, "")
				if err != nil {
					return nil, nil, nil, fmt.Errorf("failed to create JSON web key set: %w", err)
				}
			} else if authJWT.Method.PublicKeyFile != "" {
				keyPEMBlock, err := os.ReadFile(authJWT.Method.PublicKeyFile)
				if err != nil {
					return nil, nil, nil, fmt.Errorf("failed to read key file: %w", err)
				}

				publicKey, err := jwt.ParsePublicKeyPEM(keyPEMBlock)
				if err != nil {
					return nil, nil, nil, fmt.Errorf("failed to parse public key PEM block: %w", err)
				}

				ks, err = jwt.NewStaticKeySet([]crypto.PublicKey{publicKey})
				if err != nil {
					return nil, nil, nil, fmt.Errorf("failed to create static key set: %w", err)
				}
			}

			validator, err := jwt.NewValidator(ks)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to create JWT validator: %w", err)
			}

			// intentionally restricted to a set of common asymmetric algorithms
			// if we want to support symmetric algorithms, we need to add support for
			// private keys in the configuration
			exp := jwt.Expected{
				SigningAlgorithms: []jwt.Alg{jwt.RS256, jwt.RS512, jwt.ES256, jwt.ES512, jwt.EdDSA},
			}

			if authJWT.Method.ValidateClaims.Issuer != "" {
				exp.Issuer = authJWT.Method.ValidateClaims.Issuer
			}

			if authJWT.Method.ValidateClaims.Subject != "" {
				exp.Subject = authJWT.Method.ValidateClaims.Subject
			}

			if len(authJWT.Method.ValidateClaims.Audiences) != 0 {
				exp.Audiences = authJWT.Method.ValidateClaims.Audiences
			}

			interceptors = append(interceptors, selector.UnaryServerInterceptor(authmiddlewaregrpc.JWTAuthenticationInterceptor(logger, *validator, exp, authOpts...), authmiddlewaregrpc.JWTInterceptorSelector()))
		}

		if authCfg.Methods.Cloud.Enabled {
			jwksURL := fmt.Sprintf("https://%s/api/auth/jwks", cfg.Cloud.Host)

			ks, err := jwt.NewJSONWebKeySet(ctx, jwksURL, "")
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to create JSON web key set: %w", err)
			}

			validator, err := jwt.NewValidator(ks)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("failed to create JWT validator: %w", err)
			}

			// intentionally restricted to a set of common asymmetric algorithms
			exp := jwt.Expected{
				SigningAlgorithms: []jwt.Alg{jwt.RS256, jwt.RS512, jwt.ES256, jwt.ES512, jwt.EdDSA},
				Issuer:            cfg.Cloud.Host,
				Audiences:         []string{cfg.Cloud.Organization},
			}

			interceptors = append(interceptors, selector.UnaryServerInterceptor(authmiddlewaregrpc.JWTAuthenticationInterceptor(logger, *validator, exp, authOpts...), authmiddlewaregrpc.JWTInterceptorSelector()))
		}

		interceptors = append(interceptors, selector.UnaryServerInterceptor(authmiddlewaregrpc.ClientTokenAuthenticationInterceptor(
			logger,
			store,
			authOpts...,
		), authmiddlewaregrpc.ClientTokenInterceptorSelector()))

		if authCfg.Methods.OIDC.Enabled && len(authCfg.Methods.OIDC.Method.EmailMatches) != 0 {
			rgxs := make([]*regexp.Regexp, 0, len(authCfg.Methods.OIDC.Method.EmailMatches))

			for _, em := range authCfg.Methods.OIDC.Method.EmailMatches {
				rgx, err := regexp.Compile(em)
				if err != nil {
					return nil, nil, nil, fmt.Errorf("failed compiling string for pattern: %s: %w", em, err)
				}

				rgxs = append(rgxs, rgx)
			}

			interceptors = append(interceptors, selector.UnaryServerInterceptor(authmiddlewaregrpc.EmailMatchingInterceptor(logger, rgxs, authOpts...), authmiddlewaregrpc.ClientTokenInterceptorSelector()))
		}

		if authCfg.Methods.Token.Enabled {
			interceptors = append(interceptors, selector.UnaryServerInterceptor(authmiddlewaregrpc.NamespaceMatchingInterceptor(logger, authOpts...), authmiddlewaregrpc.ClientTokenInterceptorSelector()))
		}

		// at this point, we have already registered all authentication methods that are enabled
		// so atleast one authentication method should pass if authentication is required
		interceptors = append(interceptors, authmiddlewaregrpc.AuthenticationRequiredInterceptor(logger, authOpts...))

		logger.Info("authentication middleware enabled")
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
		authmiddleware = authmiddlewarehttp.NewHTTPMiddleware(cfg.Session)
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

	if cfg.SessionEnabled() {
		muxOpts = append(muxOpts, runtime.WithMetadata(method.ForwardCookies), runtime.WithMetadata(method.ForwardPrefix))

		methodMiddleware := method.NewHTTPMiddleware(cfg.Session)
		muxOpts = append(muxOpts, runtime.WithForwardResponseOption(methodMiddleware.ForwardResponseOption))

		if cfg.Methods.OIDC.Enabled {
			muxOpts = append(muxOpts, registerFunc(ctx, conn, rpcauth.RegisterAuthenticationMethodOIDCServiceHandler))
		}

		if cfg.Methods.Github.Enabled {
			muxOpts = append(muxOpts, registerFunc(ctx, conn, rpcauth.RegisterAuthenticationMethodGithubServiceHandler))
		}

		middleware = append(middleware, methodMiddleware.Handler)
	}

	if cfg.Methods.Kubernetes.Enabled {
		muxOpts = append(muxOpts, registerFunc(ctx, conn, rpcauth.RegisterAuthenticationMethodKubernetesServiceHandler))
	}

	r.Group(func(r chi.Router) {
		r.Use(middleware...)

		r.Mount("/auth/v1", gateway.NewGatewayServeMux(logger, muxOpts...))
	})
}
