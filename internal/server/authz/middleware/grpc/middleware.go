package grpc_middleware

import (
	"context"

	"go.flipt.io/flipt/internal/containers"
	authmiddlewaregrpc "go.flipt.io/flipt/internal/server/authn/middleware/grpc"
	"go.flipt.io/flipt/internal/server/authz"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var errUnauthorized = status.Error(codes.PermissionDenied, "request was not authorized")

// SkipsAuthorizationServer is a grpc.Server which should always skip authentication.
type SkipsAuthorizationServer interface {
	SkipsAuthorization(ctx context.Context) bool
}

// InterceptorOptions configure the basic AuthUnaryInterceptors
type InterceptorOptions struct {
	skippedServers []any
}

func skipped(ctx context.Context, server any, o InterceptorOptions) bool {
	if skipSrv, ok := server.(SkipsAuthorizationServer); ok && skipSrv.SkipsAuthorization(ctx) {
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

// WithServerSkipsAuthorization can be used to configure an auth unary interceptor
// which skips authorization when the provided server instance matches the intercepted
// calls parent server instance.
// This allows the caller to registers servers which explicitly skip authorization (e.g. OIDC).
func WithServerSkipsAuthorization(server any) containers.Option[InterceptorOptions] {
	return func(o *InterceptorOptions) {
		o.skippedServers = append(o.skippedServers, server)
	}
}

func AuthorizationRequiredInterceptor(logger *zap.Logger, policyVerifier authz.Verifier, o ...containers.Option[InterceptorOptions]) grpc.UnaryServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// skip authz for any preconfigured servers
		if skipped(ctx, info.Server, opts) {
			logger.Debug("skipping authorization for server", zap.String("method", info.FullMethod))
			return handler(ctx, req)
		}

		requester, ok := req.(flipt.Requester)
		if !ok {
			logger.Error("request must implement flipt.Requester", zap.String("method", info.FullMethod))
			return ctx, errUnauthorized
		}

		auth := authmiddlewaregrpc.GetAuthenticationFrom(ctx)
		if auth == nil {
			logger.Error("unauthorized", zap.String("reason", "authentication required"))
			return ctx, errUnauthorized
		}

		role, ok := auth.Metadata["io.flipt.auth.role"]
		if !ok {
			logger.Error("unauthorized", zap.String("reason", "user role required"))
			return ctx, errUnauthorized
		}

		request := requester.Request()
		action := string(request.Action)
		if action == "" {
			logger.Error("unauthorized", zap.String("reason", "request action required"))
			return ctx, errUnauthorized
		}

		scope := string(request.Scope)
		if scope == "" {
			// fallback to subject if scope is not provided
			scope = string(request.Subject)
		}

		if scope == "" {
			logger.Error("unauthorized", zap.String("reason", "request scope or subject required"))
			return ctx, errUnauthorized
		}

		allowed, err := policyVerifier.IsAllowed(ctx, map[string]interface{}{
			"role":   role,
			"action": action,
			"scope":  scope,
		})

		if err != nil {
			logger.Error("unauthorized", zap.Error(err))
			return ctx, errUnauthorized
		}

		if !allowed {
			logger.Error("unauthorized", zap.String("reason", "permission denied"))
			return ctx, errUnauthorized
		}

		return handler(ctx, req)
	}
}
