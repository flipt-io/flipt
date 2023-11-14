package grpc_middleware

import (
	"context"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"go.flipt.io/flipt/internal/containers"
	middlewarecommon "go.flipt.io/flipt/internal/server/auth/middleware/common"
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

var ErrUnauthenticated = status.Error(codes.Unauthenticated, "request was not authenticated")

type authenticationContextKey struct{}

// Authenticator is the minimum subset of an authentication provider
// required by the middleware to perform lookups for Authentication instances
// using a obtained clientToken.
type Authenticator interface {
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

// InterceptorOptions configure the UnaryInterceptor
type InterceptorOptions struct {
	skippedServers []any
}

func skipped(ctx context.Context, server any, o InterceptorOptions) bool {
	if skipSrv, ok := server.(SkipsAuthenticationServer); ok && skipSrv.SkipsAuthentication(ctx) {
		return true
	}

	// TODO: refactor to remove this check
	for _, s := range o.skippedServers {
		if s == server {
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

// UnaryInterceptor is a grpc.UnaryServerInterceptor which extracts a clientToken found
// within the authorization field on the incoming requests metadata.
// The fields value is expected to be in the form "Bearer <clientToken>".
func UnaryInterceptor(logger *zap.Logger, authenticator Authenticator, o ...containers.Option[InterceptorOptions]) grpc.UnaryServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// skip auth for any preconfigured servers
		if skipped(ctx, info.Server, opts) {
			logger.Debug("skipping authentication for server", zap.String("method", info.FullMethod))
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Error("unauthenticated", zap.String("reason", "metadata not found on context"))
			return ctx, ErrUnauthenticated
		}

		clientToken, err := clientTokenFromMetadata(md)
		if err != nil {
			logger.Error("unauthenticated",
				zap.String("reason", "no authorization provided"),
				zap.Error(err))

			return ctx, ErrUnauthenticated
		}

		auth, err := authenticator.GetAuthenticationByClientToken(ctx, clientToken)
		if err != nil {
			logger.Error("unauthenticated",
				zap.String("reason", "error retrieving authentication for client token"),
				zap.Error(err))

			if errors.Is(err, context.Canceled) {
				err = status.Error(codes.Canceled, err.Error())
				return ctx, err
			}

			if errors.Is(err, context.DeadlineExceeded) {
				err = status.Error(codes.DeadlineExceeded, err.Error())
				return ctx, err
			}

			return ctx, ErrUnauthenticated
		}

		if auth.ExpiresAt != nil && auth.ExpiresAt.AsTime().Before(time.Now()) {
			logger.Error("unauthenticated",
				zap.String("reason", "authorization expired"),
				zap.String("authentication_id", auth.Id),
			)
			return ctx, ErrUnauthenticated
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
		if skipped(ctx, info.Server, opts) {
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
			return ctx, ErrUnauthenticated
		}

		matched := false

		for _, rgx := range rgxs {
			if matched = rgx.MatchString(email); matched {
				break
			}
		}

		if !matched {
			logger.Error("unauthenticated", zap.String("reason", "email is not allowed"))
			return ctx, ErrUnauthenticated
		}

		return handler(ctx, req)
	}
}

func NamespaceMatchingInterceptor(logger *zap.Logger, o ...containers.Option[InterceptorOptions]) grpc.UnaryServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// skip auth for any preconfigured servers
		if skipped(ctx, info.Server, opts) {
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
			return ctx, ErrUnauthenticated
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
						zap.String("reason", "namespace is not allowed"))
					return ctx, ErrUnauthenticated
				}
			}
		default:
			// if the the token has a namespace but the request does not then we should reject the request
			logger.Error("unauthenticated",
				zap.String("reason", "namespace is not allowed"))
			return ctx, ErrUnauthenticated
		}

		if reqNamespace != namespace {
			logger.Error("unauthenticated",
				zap.String("reason", "namespace is not allowed"))
			return ctx, ErrUnauthenticated
		}

		return handler(ctx, req)
	}
}

func clientTokenFromMetadata(md metadata.MD) (string, error) {
	if authenticationHeader := md.Get(authenticationHeaderKey); len(authenticationHeader) > 0 {
		return clientTokenFromAuthorization(authenticationHeader[0])
	}

	cookie, err := cookieFromMetadata(md, middlewarecommon.TokenCookieKey)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

func clientTokenFromAuthorization(auth string) (string, error) {
	// ensure token was prefixed with "Bearer "
	if clientToken := strings.TrimPrefix(auth, "Bearer "); auth != clientToken {
		return clientToken, nil
	}

	return "", ErrUnauthenticated
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
