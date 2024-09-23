package edge

import (
	"context"
	"crypto"
	"fmt"
	"os"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/hashicorp/cap/jwt"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/containers"
	authmiddlewaregrpc "go.flipt.io/flipt/internal/server/authn/middleware/grpc"
	"go.flipt.io/flipt/internal/server/authn/public"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func authenticationGRPC(
	ctx context.Context,
	logger *zap.Logger,
	cfg *config.Config,
	authOpts ...containers.Option[authmiddlewaregrpc.InterceptorOptions],
) (_ grpcRegisterers, interceptors []grpc.UnaryServerInterceptor, err error) {
	var (
		authCfg  = cfg.Authentication
		register = grpcRegisterers{
			public.NewServer(logger, authCfg),
		}
	)

	if !authCfg.Enabled() {
		return grpcRegisterers{
			public.NewServer(logger, authCfg),
		}, nil, nil
	}

	// only enable enforcement middleware if authentication required
	if authCfg.Required {
		if authCfg.Methods.JWT.Enabled {
			authJWT := authCfg.Methods.JWT

			var ks jwt.KeySet

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

			interceptors = append(interceptors, selector.UnaryServerInterceptor(authmiddlewaregrpc.JWTAuthenticationInterceptor(logger, *validator, exp, authOpts...), authmiddlewaregrpc.JWTInterceptorSelector()))
		}

		if authCfg.Methods.Cloud.Enabled {
			jwksURL := fmt.Sprintf("https://%s/api/auth/jwks", cfg.Cloud.Host)

			ks, err := jwt.NewJSONWebKeySet(ctx, jwksURL, "")
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create JSON web key set: %w", err)
			}

			validator, err := jwt.NewValidator(ks)
			if err != nil {
				return nil, nil, fmt.Errorf("failed to create JWT validator: %w", err)
			}

			// intentionally restricted to a set of common asymmetric algorithms
			exp := jwt.Expected{
				SigningAlgorithms: []jwt.Alg{jwt.RS256, jwt.RS512, jwt.ES256, jwt.ES512, jwt.EdDSA},
				Issuer:            cfg.Cloud.Host,
				Audiences:         []string{cfg.Cloud.Organization},
			}

			interceptors = append(interceptors, selector.UnaryServerInterceptor(authmiddlewaregrpc.JWTAuthenticationInterceptor(logger, *validator, exp, authOpts...), authmiddlewaregrpc.JWTInterceptorSelector()))
		}

		// at this point, we have already registered all authentication methods that are enabled
		// so atleast one authentication method should pass if authentication is required
		interceptors = append(interceptors, authmiddlewaregrpc.AuthenticationRequiredInterceptor(logger, authOpts...))

		logger.Info("authentication middleware enabled")
	}

	return register, interceptors, nil
}
