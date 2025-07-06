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

	"slices"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/server/authn/method"
	middlewarecommon "go.flipt.io/flipt/internal/server/authn/middleware/common"
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

func skipped(ctx context.Context, server any, o InterceptorOptions) bool {
	if skipSrv, ok := server.(SkipsAuthenticationServer); ok && skipSrv.SkipsAuthentication(ctx) {
		return true
	}

	// TODO: refactor to remove this check
	return slices.Contains(o.skippedServers, server)
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

// SkipsAuthenticationServer is a grpc.Server which should always skip authentication.
type SkipsAuthenticationServer interface {
	SkipsAuthentication(ctx context.Context) bool
}

// AuthenticationRequiredUnaryInterceptor is a grpc.UnaryServerInterceptor which requires that
// all requests contain an Authentication instance on the context.
func AuthenticationRequiredUnaryInterceptor(logger *zap.Logger, o ...containers.Option[InterceptorOptions]) grpc.UnaryServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// skip auth for any preconfigured servers
		if skipped(ctx, info.Server, opts) {
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

// AuthenticationRequiredStreamInterceptor is a grpc.StreamServerInterceptor which requires that
// all requests contain an Authentication instance on the context.
func AuthenticationRequiredStreamInterceptor(logger *zap.Logger, o ...containers.Option[InterceptorOptions]) grpc.StreamServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		// skip auth for any preconfigured servers
		if skipped(ctx, srv, opts) {
			logger.Debug("skipping authentication for server", zap.String("method", info.FullMethod))
			return handler(srv, stream)
		}

		auth := GetAuthenticationFrom(ctx)
		if auth == nil {
			logger.Error("unauthenticated", zap.String("reason", "authentication required"))
			return errUnauthenticated
		}

		return handler(srv, stream)
	}
}

// JWTInterceptorSelector is a selector.Matcher which selects requests
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

// JWTAuthenticationUnaryInterceptor is a grpc.UnaryServerInterceptor which extracts a JWT found
// within the authorization field on the incoming requests metadata.
func JWTAuthenticationUnaryInterceptor(logger *zap.Logger, validator method.JWTValidator, o ...containers.Option[InterceptorOptions]) grpc.UnaryServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// skip auth for any preconfigured servers
		if skipped(ctx, info.Server, opts) {
			logger.Debug("skipping authentication for server", zap.String("method", info.FullMethod))
			return handler(ctx, req)
		}

		ctx, err := authenticateJWT(ctx, logger, validator)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// JWTAuthenticationStreamInterceptor is a grpc.StreamServerInterceptor which extracts a JWT found
// within the authorization field on the incoming requests metadata.
func JWTAuthenticationStreamInterceptor(logger *zap.Logger, validator method.JWTValidator, o ...containers.Option[InterceptorOptions]) grpc.StreamServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		// skip auth for any preconfigured servers
		if skipped(ctx, srv, opts) {
			logger.Debug("skipping authentication for server", zap.String("method", info.FullMethod))
			return handler(srv, stream)
		}

		ctx, err := authenticateJWT(ctx, logger, validator)
		if err != nil {
			return err
		}

		// wrappedServerStream is a helper that allows modifying the context of the server stream
		return handler(srv, &grpcmiddleware.WrappedServerStream{
			ServerStream:   stream,
			WrappedContext: ctx,
		})
	}
}

// authenticateJWT authenticates a JWT found in the incoming request metadata and returns a new context with the authenticated authentication instance.
func authenticateJWT(ctx context.Context, logger *zap.Logger, validator method.JWTValidator) (context.Context, error) {
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

	jwtClaims, err := validator.Validate(ctx, token)
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
			userClaims, ok := v.(map[string]any)
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

	return ContextWithAuthentication(ctx, &authrpc.Authentication{
		Method:   authrpc.Method_METHOD_JWT,
		Metadata: metadata,
	}), nil
}

// ClientTokenInterceptorSelector is a selector.Matcher which selects requests
// which contain a client token in the authorization field on the incoming requests metadata.
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

// ClientTokenAuthenticationUnaryInterceptor is a grpc.UnaryServerInterceptor which extracts a clientToken found
// within the authorization field on the incoming requests metadata.
// The fields value is expected to be in the form "Bearer <clientToken>".
func ClientTokenAuthenticationUnaryInterceptor(logger *zap.Logger, authenticator ClientTokenAuthenticator, o ...containers.Option[InterceptorOptions]) grpc.UnaryServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// skip auth for any preconfigured servers
		if skipped(ctx, info.Server, opts) {
			logger.Debug("skipping authentication for server", zap.String("method", info.FullMethod))
			return handler(ctx, req)
		}

		ctx, err := authenticateClientToken(ctx, logger, authenticator)
		if err != nil {
			return nil, err
		}

		return handler(ctx, req)
	}
}

// ClientTokenStreamInterceptor is a grpc.StreamServerInterceptor which extracts a clientToken found
// within the authorization field on the incoming requests metadata.
// The fields value is expected to be in the form "Bearer <clientToken>".
func ClientTokenStreamInterceptor(logger *zap.Logger, authenticator ClientTokenAuthenticator, o ...containers.Option[InterceptorOptions]) grpc.StreamServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		// skip auth for any preconfigured servers
		if skipped(ctx, srv, opts) {
			logger.Debug("skipping authentication for server", zap.String("method", info.FullMethod))
			return handler(srv, stream)
		}

		ctx, err := authenticateClientToken(ctx, logger, authenticator)
		if err != nil {
			return err
		}

		// wrappedServerStream is a helper that allows modifying the context of the server stream
		return handler(srv, &grpcmiddleware.WrappedServerStream{
			ServerStream:   stream,
			WrappedContext: ctx,
		})
	}
}

// authenticateClientToken authenticates a client token found in the incoming request metadata and returns a new context with the authenticated authentication instance.
func authenticateClientToken(ctx context.Context, logger *zap.Logger, authenticator ClientTokenAuthenticator) (context.Context, error) {
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

	return ContextWithAuthentication(ctx, auth), nil
}

// EmailMatchingUnaryInterceptor is a grpc.UnaryServerInterceptor only used in the case where the user is using OIDC
// and wants to whitelist a group of users issuing operations against the Flipt server.
func EmailMatchingUnaryInterceptor(logger *zap.Logger, rgxs []*regexp.Regexp, o ...containers.Option[InterceptorOptions]) grpc.UnaryServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
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

// clientTokenFromMetadata extracts a client token found in the incoming request metadata
// and returns the client token.
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

// cookieFromMetadata extracts a cookie found in the incoming request metadata
// and returns the cookie.
func cookieFromMetadata(md metadata.MD, key string) (*http.Cookie, error) {
	// sadly net/http does not expose cookie parsing
	// outside of http.Request.
	// so instead we fabricate a request around the cookie
	// in order to extract it appropriately.
	return (&http.Request{
		Header: http.Header{"Cookie": md.Get(cookieHeaderKey)},
	}).Cookie(key)
}

// jwtFromMetadata extracts a JWT found in the incoming request metadata
// and returns the JWT.
func jwtFromMetadata(md metadata.MD) (string, error) {
	if authenticationHeader := md.Get(authenticationHeaderKey); len(authenticationHeader) > 0 {
		return fromAuthorization(authenticationHeader[0], authenticationSchemeJWT)
	}

	return "", errUnauthenticated
}

// fromAuthorization extracts a token from an authorization header
// and returns the token.
func fromAuthorization(auth string, scheme authenticationScheme) (string, error) {
	// Ensure auth is prefixed with the scheme
	if a, ok := strings.CutPrefix(auth, scheme.String()+" "); ok {
		return a, nil
	}

	return "", errUnauthenticated
}
