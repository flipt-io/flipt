package auth

import (
	"context"
	"strings"
	"time"

	authrpc "go.flipt.io/flipt/rpc/flipt/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const authenticationHeaderKey = "authorization"

var errUnauthenticated = status.Error(codes.Unauthenticated, "request was not authenticated")

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

// UnaryInterceptor is a grpc.UnaryServerInterceptor which extracts a clientToken found
// within the authorization field on the incoming requests metadata.
// The fields value is expected to be in the form "Bearer <clientToken>".
func UnaryInterceptor(logger *zap.Logger, authenticator Authenticator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Error("unauthenticated", zap.String("reason", "metadata not found on context"))
			return ctx, errUnauthenticated
		}

		authenticationHeader := md.Get(authenticationHeaderKey)
		if len(authenticationHeader) < 1 {
			logger.Error("unauthenticated", zap.String("reason", "no authorization provided"))
			return ctx, errUnauthenticated
		}

		clientToken := strings.TrimPrefix(authenticationHeader[0], "Bearer ")
		// ensure token was prefixed with "Bearer "
		if authenticationHeader[0] == clientToken {
			logger.Error("unauthenticated", zap.String("reason", "authorization malformed"))
			return ctx, errUnauthenticated
		}

		auth, err := authenticator.GetAuthenticationByClientToken(ctx, clientToken)
		if err != nil {
			logger.Error("unauthenticated",
				zap.String("reason", "error retrieving authentication for client token"),
				zap.Error(err))
			return ctx, errUnauthenticated
		}

		if auth.ExpiresAt != nil && auth.ExpiresAt.AsTime().Before(time.Now()) {
			logger.Error("unauthenticated",
				zap.String("reason", "authorization expired"),
				zap.String("authentication_id", auth.Id),
			)
			return ctx, errUnauthenticated
		}

		return handler(context.WithValue(ctx, authenticationContextKey{}, auth), req)
	}
}
