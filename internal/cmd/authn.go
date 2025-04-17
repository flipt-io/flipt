package cmd

import (
	"context"
	"crypto"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/fullstorydev/grpchan"
	"github.com/go-chi/chi/v5"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/hashicorp/cap/jwt"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/gateway"
	"go.flipt.io/flipt/internal/server/authn"
	"go.flipt.io/flipt/internal/server/authn/method"
	authgithub "go.flipt.io/flipt/internal/server/authn/method/github"
	authjwt "go.flipt.io/flipt/internal/server/authn/method/jwt"
	authkubernetes "go.flipt.io/flipt/internal/server/authn/method/kubernetes"
	authoidc "go.flipt.io/flipt/internal/server/authn/method/oidc"
	authmiddlewaregrpc "go.flipt.io/flipt/internal/server/authn/middleware/grpc"
	authmiddlewarehttp "go.flipt.io/flipt/internal/server/authn/middleware/http"
	"go.flipt.io/flipt/internal/server/authn/public"
	storageauth "go.flipt.io/flipt/internal/storage/authn"
	storageauthmemory "go.flipt.io/flipt/internal/storage/authn/memory"
	storageauthredis "go.flipt.io/flipt/internal/storage/authn/redis"
	"go.flipt.io/flipt/internal/storage/authn/static"
	rpcauth "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func getAuthStore(
	logger *zap.Logger,
	cfg *config.Config,
) (storageauth.Store, error) {

	var (
		cleanupGracePeriod                   = cfg.Authentication.Session.Storage.Cleanup.GracePeriod
		store              storageauth.Store = storageauthmemory.NewStore(logger, storageauthmemory.WithCleanupGracePeriod(cleanupGracePeriod))
	)

	if cfg.Authentication.Session.Storage.Type == config.AuthenticationSessionStorageTypeRedis {
		rdb, err := storageauthredis.NewClient(cfg.Authentication.Session.Storage.Redis)
		if err != nil {
			return nil, fmt.Errorf("failed to create redis client: %w", err)
		}

		store = storageauthredis.NewStore(rdb, logger, storageauthredis.WithCleanupGracePeriod(cleanupGracePeriod))
	}

	// if token method is enabled we decorate the store with a static store implementation
	// which is populated with the configured tokens
	if cfg.Authentication.Methods.Token.Enabled {
		var err error
		store, err = static.NewStore(store, logger, cfg.Authentication.Methods.Token.Method.Storage)
		if err != nil {
			return nil, err
		}
	}

	return store, nil
}

func authenticationGRPC(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.Config,
	handlers *grpchan.HandlerMap,
	authOpts ...containers.Option[authmiddlewaregrpc.InterceptorOptions],
) ([]grpc.UnaryServerInterceptor, func(context.Context) error, error) {

	var (
		shutdown = func(ctx context.Context) error {
			return nil
		}
		authCfg = cfg.Authentication
	)

	if !authCfg.Enabled() {
		rpcauth.RegisterPublicAuthenticationServiceServer(handlers, public.NewServer(logger, authCfg))
		rpcauth.RegisterAuthenticationServiceServer(handlers, authn.NewServer(logger, storageauthmemory.NewStore(logger)))
		return nil, shutdown, nil
	}

	store, err := getAuthStore(logger, cfg)
	if err != nil {
		return nil, nil, err
	}

	rpcauth.RegisterPublicAuthenticationServiceServer(handlers, public.NewServer(logger, authCfg))
	rpcauth.RegisterAuthenticationServiceServer(handlers, authn.NewServer(logger, storageauthmemory.NewStore(logger)))

	shutdown = store.Shutdown

	var interceptors []grpc.UnaryServerInterceptor

	// register auth method oidc service
	if authCfg.Methods.OIDC.Enabled {
		rpcauth.RegisterAuthenticationMethodOIDCServiceServer(handlers, authoidc.NewServer(logger, store, authCfg))

		logger.Debug("authentication method \"oidc\" server registered")
	}

	if authCfg.Methods.Github.Enabled {
		rpcauth.RegisterAuthenticationMethodGithubServiceServer(handlers, authgithub.NewServer(logger, store, authCfg))

		logger.Debug("authentication method \"github\" registered")
	}

	var jwtValidator method.JWTValidator

	if authCfg.Methods.Kubernetes.Enabled {
		jwtValidator, err = authkubernetes.NewValidator(logger, authCfg.Methods.Kubernetes.Method)
		if err != nil {
			return nil, nil, fmt.Errorf("configuring kubernetes authentication: %w", err)
		}

		kubernetesServer, err := authkubernetes.NewServer(logger, authCfg, jwtValidator)
		if err != nil {
			return nil, nil, fmt.Errorf("configuring kubernetes authentication: %w", err)
		}
		rpcauth.RegisterAuthenticationMethodKubernetesServiceServer(handlers, kubernetesServer)

		logger.Debug("authentication method \"kubernetes\" server registered")
	}

	// Set up JWT validator if JWT auth is enabled
	if authCfg.Methods.JWT.Enabled {
		var (
			authJWT = authCfg.Methods.JWT
			ks      jwt.KeySet
		)

		if authJWT.Method.JWKSURL != "" {
			ks, err = jwt.NewJSONWebKeySet(ctx, authJWT.Method.JWKSURL, "")
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create JSON web key set: %w", err)
			}
		} else if authJWT.Method.PublicKeyFile != "" {
			keyPEMBlock, err := os.ReadFile(authJWT.Method.PublicKeyFile)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to read key file: %w", err)
			}

			publicKey, err := jwt.ParsePublicKeyPEM(keyPEMBlock)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to parse public key PEM block: %w", err)
			}

			ks, err = jwt.NewStaticKeySet([]crypto.PublicKey{publicKey})
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create static key set: %w", err)
			}
		}

		validator, err := jwt.NewValidator(ks)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create JWT validator: %w", err)
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

		// wrap the validator in a method.JWTValidator implementation
		jwtValidator = authjwt.NewValidator(validator, exp)
	}

	// only enable enforcement middleware if authentication required
	if authCfg.Required {
		// Register JWT validation middleware if either JWT or Kubernetes auth is enabled
		if authCfg.Methods.JWT.Enabled || authCfg.Methods.Kubernetes.Enabled {
			interceptors = append(interceptors, selector.UnaryServerInterceptor(authmiddlewaregrpc.JWTAuthenticationInterceptor(logger, jwtValidator, authOpts...), authmiddlewaregrpc.JWTInterceptorSelector()))
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
					return nil, nil, fmt.Errorf("failed compiling string for pattern: %s: %w", em, err)
				}

				rgxs = append(rgxs, rgx)
			}

			interceptors = append(interceptors, selector.UnaryServerInterceptor(authmiddlewaregrpc.EmailMatchingInterceptor(logger, rgxs, authOpts...), authmiddlewaregrpc.ClientTokenInterceptorSelector()))
		}

		// at this point, we have already registered all authentication methods that are enabled
		// so atleast one authentication method should pass if authentication is required
		interceptors = append(interceptors, authmiddlewaregrpc.AuthenticationRequiredInterceptor(logger, authOpts...))

		logger.Info("authentication middleware enabled")
	}

	return interceptors, shutdown, nil
}

// register creates a ServeMuxOption that registers a gRPC service handler
func register[T any](ctx context.Context, client T, register func(context.Context, *runtime.ServeMux, T) error) runtime.ServeMuxOption {
	return func(mux *runtime.ServeMux) {
		if err := register(ctx, mux, client); err != nil {
			panic(err)
		}
	}
}

func authenticationHTTPMount(
	ctx context.Context,
	logger *zap.Logger,
	cfg config.AuthenticationConfig,
	r chi.Router,
	conn grpc.ClientConnInterface,
) {
	var (
		authmiddleware = authmiddlewarehttp.NewHTTPMiddleware(cfg.Session)
		middleware     = []func(next http.Handler) http.Handler{authmiddleware.Handler}
		muxOpts        = []runtime.ServeMuxOption{
			register(ctx, rpcauth.NewPublicAuthenticationServiceClient(conn), rpcauth.RegisterPublicAuthenticationServiceHandlerClient),
			register(ctx, rpcauth.NewAuthenticationServiceClient(conn), rpcauth.RegisterAuthenticationServiceHandlerClient),
			runtime.WithErrorHandler(authmiddleware.ErrorHandler),
		}
	)

	if cfg.SessionEnabled() {
		muxOpts = append(muxOpts, runtime.WithMetadata(method.ForwardCookies), runtime.WithMetadata(method.ForwardPrefix))

		methodMiddleware := method.NewHTTPMiddleware(cfg.Session)
		muxOpts = append(muxOpts, runtime.WithForwardResponseOption(methodMiddleware.ForwardResponseOption))

		if cfg.Methods.OIDC.Enabled {
			muxOpts = append(muxOpts, register(ctx, rpcauth.NewAuthenticationMethodOIDCServiceClient(conn), rpcauth.RegisterAuthenticationMethodOIDCServiceHandlerClient))
		}

		if cfg.Methods.Github.Enabled {
			muxOpts = append(muxOpts, register(ctx, rpcauth.NewAuthenticationMethodGithubServiceClient(conn), rpcauth.RegisterAuthenticationMethodGithubServiceHandlerClient))
		}

		middleware = append(middleware, methodMiddleware.Handler)
	}

	if cfg.Methods.Kubernetes.Enabled {
		muxOpts = append(muxOpts, register(ctx, rpcauth.NewAuthenticationMethodKubernetesServiceClient(conn), rpcauth.RegisterAuthenticationMethodKubernetesServiceHandlerClient))
	}

	r.Group(func(r chi.Router) {
		r.Use(middleware...)
		r.Mount("/auth/v1", gateway.NewGatewayServeMux(logger, muxOpts...))
	})
}
