package grpc_middleware

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.flipt.io/flipt/errors"

	errs "errors"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"github.com/hashicorp/cap/jwt"
	"go.flipt.io/flipt/internal/containers"
	middlewarecommon "go.flipt.io/flipt/internal/server/authn/middleware/common"
	"go.flipt.io/flipt/rpc/flipt"
	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	authenticationHeaderKey = "authorization"
	cookieHeaderKey         = "grpcgateway-cookie"
)

type authenticationScheme uint8

const (
	_ authenticationScheme = iota
	authenticationSchemeBearer
	authenticationSchemeJWT
)

func (a authenticationScheme) String() string {
	switch a {
	case authenticationSchemeBearer:
		return "Bearer"
	case authenticationSchemeJWT:
		return "JWT"
	default:
		return ""
	}
}

var errUnauthenticated = errors.ErrUnauthenticatedf("request was not authenticated")

type authenticationContextKey struct{}

// ClientTokenAuthenticator is the minimum subset of an authentication provider
// required by the middleware to perform lookups for Authentication instances
// using a obtained clientToken.
type ClientTokenAuthenticator interface {
	GetAuthenticationByClientToken(ctx context.Context, clientToken string) (*authrpc.Authentication, error)
}

// GetAuthenticationFrom is a utility for extracting an Authentication stored
// on a context.Context instance
func GetAuthenticationFrom(ctx context.Context) *authrpc.Authentication {
	auth := ctx.Value(authenticationContextKey{})
	if auth == nil {
		return nil
	}

	return auth.(*authrpc.Authentication)
}

// ContextWithAuthentication returns a context with the specified authentication
func ContextWithAuthentication(ctx context.Context, a *authrpc.Authentication) context.Context {
	return context.WithValue(ctx, authenticationContextKey{}, a)
}

// InterceptorOptions configure the basic AuthUnaryInterceptors
type InterceptorOptions struct {
	skippedServers []any
}

func skipped(ctx context.Context, info *grpc.UnaryServerInfo, o InterceptorOptions) bool {
	if skipSrv, ok := info.Server.(SkipsAuthenticationServer); ok && skipSrv.SkipsAuthentication(ctx) {
		return true
	}

	// TODO: refactor to remove this check
	for _, s := range o.skippedServers {
		if s == info.Server {
			return true
		}
	}

	return false
}

// WithServerSkipsAuthentication can be used to configure an auth unary interceptor
// which skips authentication when the provided server instance matches the intercepted
// calls parent server instance.
// This allows the caller to registers servers which explicitly skip authentication (e.g. OIDC).
func WithServerSkipsAuthentication(server any) containers.Option[InterceptorOptions] {
	return func(o *InterceptorOptions) {
		o.skippedServers = append(o.skippedServers, server)
	}
}

// ScopedAuthenticationServer is a grpc.Server which allows for specific scoped authentication.
type ScopedAuthenticationServer interface {
	AllowsNamespaceScopedAuthentication(ctx context.Context) bool
}

// SkipsAuthenticationServer is a grpc.Server which should always skip authentication.
type SkipsAuthenticationServer interface {
	SkipsAuthentication(ctx context.Context) bool
}

// AuthenticationRequiredInterceptor is a grpc.UnaryServerInterceptor which requires that
// all requests contain an Authentication instance on the context.
func AuthenticationRequiredInterceptor(logger *zap.Logger, o ...containers.Option[InterceptorOptions]) grpc.UnaryServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// skip auth for any preconfigured servers
		if skipped(ctx, info, opts) {
			logger.Debug("skipping authentication for server", zap.String("method", info.FullMethod))
			return handler(ctx, req)
		}

		auth := GetAuthenticationFrom(ctx)
		if auth == nil {
			logger.Error("unauthenticated", zap.String("reason", "authentication required"))
			return ctx, errUnauthenticated
		}

		return handler(ctx, req)
	}
}

// JWTInterceptorSelector is a grpc.UnaryServerInterceptor which selects requests
// which contain a JWT in the authorization header.
func JWTInterceptorSelector() selector.Matcher {
	return selector.MatchFunc(func(ctx context.Context, _ interceptors.CallMeta) bool {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return false
		}

		_, err := jwtFromMetadata(md)
		return err == nil
	})
}

func JWTAuthenticationInterceptor(logger *zap.Logger, validator jwt.Validator, expected jwt.Expected, o ...containers.Option[InterceptorOptions]) grpc.UnaryServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// skip auth for any preconfigured servers
		if skipped(ctx, info, opts) {
			logger.Debug("skipping authentication for server", zap.String("method", info.FullMethod))
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Error("unauthenticated", zap.String("reason", "metadata not found on context"))
			return ctx, errUnauthenticated
		}

		token, err := jwtFromMetadata(md)
		if err != nil {
			logger.Error("unauthenticated",
				zap.String("reason", "no authorization provided"),
				zap.Error(err))

			return ctx, errUnauthenticated
		}

		jwtClaims, err := validator.Validate(ctx, token, expected)
		if err != nil {
			logger.Error("unauthenticated",
				zap.String("reason", "error validating jwt"),
				zap.Error(err))

			if errs.Is(err, context.Canceled) {
				err = status.Error(codes.Canceled, err.Error())
				return ctx, err
			}

			if errs.Is(err, context.DeadlineExceeded) {
				err = status.Error(codes.DeadlineExceeded, err.Error())
				return ctx, err
			}

			return ctx, errUnauthenticated
		}

		metadata := map[string]string{}

		for k, v := range jwtClaims {
			if strings.HasPrefix(k, "io.flipt.auth") {
				metadata[k] = fmt.Sprintf("%v", v)
				continue
			}

			if v, ok := v.(string); ok && k == "iss" {
				metadata["io.flipt.auth.jwt.issuer"] = v
				continue
			}

			if k == "user" {
				userClaims, ok := v.(map[string]interface{})
				if ok {
					for _, fields := range [][2]string{
						{"email", "email"},
						{"sub", "sub"},
						{"image", "picture"},
						{"name", "name"},
						{"role", "role"},
					} {
						if v, ok := userClaims[fields[0]]; ok {
							metadata[fmt.Sprintf("io.flipt.auth.jwt.%s", fields[1])] = fmt.Sprintf("%v", v)
						}
					}
				}
			}
		}

		auth := &authrpc.Authentication{
			Method:   authrpc.Method_METHOD_JWT,
			Metadata: metadata,
		}

		return handler(ContextWithAuthentication(ctx, auth), req)
	}
}

func ClientTokenInterceptorSelector() selector.Matcher {
	return selector.MatchFunc(func(ctx context.Context, _ interceptors.CallMeta) bool {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return false
		}

		_, err := clientTokenFromMetadata(md)
		return err == nil
	})
}

// ClientTokenAuthenticationInterceptor is a grpc.UnaryServerInterceptor which extracts a clientToken found
// within the authorization field on the incoming requests metadata.
// The fields value is expected to be in the form "Bearer <clientToken>".
func ClientTokenAuthenticationInterceptor(logger *zap.Logger, authenticator ClientTokenAuthenticator, o ...containers.Option[InterceptorOptions]) grpc.UnaryServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// skip auth for any preconfigured servers
		if skipped(ctx, info, opts) {
			logger.Debug("skipping authentication for server", zap.String("method", info.FullMethod))
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Error("unauthenticated", zap.String("reason", "metadata not found on context"))
			return ctx, errUnauthenticated
		}

		clientToken, err := clientTokenFromMetadata(md)
		if err != nil {
			logger.Error("unauthenticated",
				zap.String("reason", "no authorization provided"),
				zap.Error(err))

			return ctx, errUnauthenticated
		}

		auth, err := authenticator.GetAuthenticationByClientToken(ctx, clientToken)
		if err != nil {
			logger.Error("unauthenticated",
				zap.String("reason", "error retrieving authentication for client token"),
				zap.Error(err))

			if errs.Is(err, context.Canceled) {
				err = status.Error(codes.Canceled, err.Error())
				return ctx, err
			}

			if errs.Is(err, context.DeadlineExceeded) {
				err = status.Error(codes.DeadlineExceeded, err.Error())
				return ctx, err
			}

			return ctx, errUnauthenticated
		}

		if auth.ExpiresAt != nil && auth.ExpiresAt.AsTime().Before(time.Now()) {
			logger.Error("unauthenticated",
				zap.String("reason", "authorization expired"),
				zap.String("authentication_id", auth.Id),
			)
			return ctx, errUnauthenticated
		}

		return handler(ContextWithAuthentication(ctx, auth), req)
	}
}

// EmailMatchingInterceptor is a grpc.UnaryServerInterceptor only used in the case where the user is using OIDC
// and wants to whitelist a group of users issuing operations against the Flipt server.
func EmailMatchingInterceptor(logger *zap.Logger, rgxs []*regexp.Regexp, o ...containers.Option[InterceptorOptions]) grpc.UnaryServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// skip auth for any preconfigured servers
		if skipped(ctx, info, opts) {
			logger.Debug("skipping authentication for server", zap.String("method", info.FullMethod))
			return handler(ctx, req)
		}

		auth := GetAuthenticationFrom(ctx)
		if auth == nil {
			panic("authentication not found in context, middleware installed incorrectly")
		}

		// this mechanism only applies to authentications created using OIDC
		if auth.Method != authrpc.Method_METHOD_OIDC {
			return handler(ctx, req)
		}

		email, ok := auth.Metadata["io.flipt.auth.oidc.email"]
		if !ok {
			logger.Debug("no email provided but required for auth")
			return ctx, errUnauthenticated
		}

		matched := false

		for _, rgx := range rgxs {
			if matched = rgx.MatchString(email); matched {
				break
			}
		}

		if !matched {
			logger.Error("unauthenticated", zap.String("reason", "email is not allowed"))
			return ctx, errUnauthenticated
		}

		return handler(ctx, req)
	}
}

func NamespaceMatchingInterceptor(logger *zap.Logger, o ...containers.Option[InterceptorOptions]) grpc.UnaryServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// skip auth for any preconfigured servers
		if skipped(ctx, info, opts) {
			logger.Debug("skipping authentication for server", zap.String("method", info.FullMethod))
			return handler(ctx, req)
		}

		auth := GetAuthenticationFrom(ctx)
		if auth == nil {
			panic("authentication not found in context, middleware installed incorrectly")
		}

		// this mechanism only applies to static toke authentications
		if auth.Method != authrpc.Method_METHOD_TOKEN {
			return handler(ctx, req)
		}

		namespace, ok := auth.Metadata["io.flipt.auth.token.namespace"]
		if !ok {
			// if no namespace is provided then we should allow the request
			return handler(ctx, req)
		}

		nsServer, ok := info.Server.(ScopedAuthenticationServer)
		if !ok || !nsServer.AllowsNamespaceScopedAuthentication(ctx) {
			logger.Error("unauthenticated",
				zap.String("reason", "namespace is not allowed"))
			return ctx, errUnauthenticated
		}

		namespace = strings.TrimSpace(namespace)
		if namespace == "" {
			return handler(ctx, req)
		}

		logger := logger.With(zap.String("expected_namespace", namespace))

		var reqNamespace string
		switch nsReq := req.(type) {
		case flipt.Namespaced:
			reqNamespace = nsReq.GetNamespaceKey()
			if reqNamespace == "" {
				reqNamespace = "default"
			}
		case flipt.BatchNamespaced:
			// ensure that all namespaces referenced in
			// the batch are the same
			for _, ns := range nsReq.GetNamespaceKeys() {
				if ns == "" {
					ns = "default"
				}

				if reqNamespace == "" {
					reqNamespace = ns
					continue
				}

				if reqNamespace != ns {
					logger.Error("unauthenticated",
						zap.String("reason", "same namespace is not set for all requests"))
					return ctx, errUnauthenticated
				}
			}
		default:
			// if the the token has a namespace but the request does not then we should reject the request
			logger.Error("unauthenticated",
				zap.String("reason", "namespace is required when using namespace scoped token"))
			return ctx, errUnauthenticated
		}

		if reqNamespace != namespace {
			logger.Error("unauthenticated",
				zap.String("reason", "namespace is not allowed"))
			return ctx, errUnauthenticated
		}

		return handler(ctx, req)
	}
}

func clientTokenFromMetadata(md metadata.MD) (string, error) {
	if authenticationHeader := md.Get(authenticationHeaderKey); len(authenticationHeader) > 0 {
		return fromAuthorization(authenticationHeader[0], authenticationSchemeBearer)
	}

	cookie, err := cookieFromMetadata(md, middlewarecommon.TokenCookieKey)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func cookieFromMetadata(md metadata.MD, key string) (*http.Cookie, error) {
	// sadly net/http does not expose cookie parsing
	// outside of http.Request.
	// so instead we fabricate a request around the cookie
	// in order to extract it appropriately.
	return (&http.Request{
		Header: http.Header{"Cookie": md.Get(cookieHeaderKey)},
	}).Cookie(key)
}

func jwtFromMetadata(md metadata.MD) (string, error) {
	if authenticationHeader := md.Get(authenticationHeaderKey); len(authenticationHeader) > 0 {
		return fromAuthorization(authenticationHeader[0], authenticationSchemeJWT)
	}

	return "", errUnauthenticated
}

func fromAuthorization(auth string, scheme authenticationScheme) (string, error) {
	// Ensure auth is prefixed with the scheme
	if a := strings.TrimPrefix(auth, scheme.String()+" "); auth != a {
		return a, nil
	}

	return "", errUnauthenticated
}
