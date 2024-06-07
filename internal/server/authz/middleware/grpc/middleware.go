package grpc_middleware

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/internal/server/authn"
	authmiddlewaregrpc "go.flipt.io/flipt/internal/server/authn/middleware/grpc"
	"go.flipt.io/flipt/internal/server/authz"
	"go.flipt.io/flipt/rpc/flipt"
	"go.opentelemetry.io/otel/trace"
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

// InterceptorOptions configure the basic AuthzUnaryInterceptors
type InterceptorOptions struct {
	skippedServers []any
}

var (
	// methods which should always skip authorization
	skippedMethods = map[string]any{
		"/flipt.auth.AuthenticationService/GetAuthenticationSelf":    struct{}{},
		"/flipt.auth.AuthenticationService/ExpireAuthenticationSelf": struct{}{},
	}
)

func skipped(ctx context.Context, info *grpc.UnaryServerInfo, o InterceptorOptions) bool {
	// if we skip authentication then we must skip authorization
	if skipSrv, ok := info.Server.(authmiddlewaregrpc.SkipsAuthenticationServer); ok && skipSrv.SkipsAuthentication(ctx) {
		return true
	}

	if skipSrv, ok := info.Server.(SkipsAuthorizationServer); ok && skipSrv.SkipsAuthorization(ctx) {
		return true
	}

	// skip authz for any preconfigured methods
	if _, ok := skippedMethods[info.FullMethod]; ok {
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

// WithServerSkipsAuthorization can be used to configure an auth unary interceptor
// which skips authorization when the provided server instance matches the intercepted
// calls parent server instance.
// This allows the caller to registers servers which explicitly skip authorization (e.g. OIDC).
func WithServerSkipsAuthorization(server any) containers.Option[InterceptorOptions] {
	return func(o *InterceptorOptions) {
		o.skippedServers = append(o.skippedServers, server)
	}
}

func AuthorizationRequiredInterceptor(logger *zap.Logger, policyVerifier authz.Verifier, eventPairChecker audit.EventPairChecker, o ...containers.Option[InterceptorOptions]) grpc.UnaryServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// skip authz for any preconfigured servers
		if skipped(ctx, info, opts) {
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

		request := requester.Request()

		allowed, err := policyVerifier.IsAllowed(ctx, map[string]interface{}{
			"request":        request,
			"authentication": auth,
		})

		var event *audit.Event

		actor := authn.ActorFromContext(ctx)

		defer func() {
			if event != nil {
				eventPair := fmt.Sprintf("%s:%s", event.Type, event.Action)

				exists := eventPairChecker.Check(eventPair)
				if exists {
					span := trace.SpanFromContext(ctx)
					span.AddEvent("event", trace.WithAttributes(event.DecodeToAttributes()...))
				}
			}
		}()

		if err != nil {
			logger.Error("unauthorized", zap.Error(err))
			request.Status = flipt.StatusDenied
			event = audit.NewEvent(request, actor, nil)
			return ctx, errUnauthorized
		}

		if !allowed {
			logger.Error("unauthorized", zap.String("reason", "permission denied"))
			event = audit.NewEvent(request, actor, nil)
			return ctx, errUnauthorized
		}

		return handler(ctx, req)
	}
}
