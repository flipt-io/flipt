package grpc_middleware

import (
	"context"
	"slices"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/containers"
	authmiddlewaregrpc "go.flipt.io/flipt/internal/server/authn/middleware/grpc"
	"go.flipt.io/flipt/internal/server/authz"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/ofrep"
	"go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// SkipsAuthorizationServer is a grpc.Server which should always skip authentication.
type SkipsAuthorizationServer interface {
	SkipsAuthorization(ctx context.Context) bool
}

// InterceptorOptions configure the basic AuthzUnaryInterceptors
type InterceptorOptions struct {
	skippedServers []any
}

// methods which should always skip authorization
var skippedMethods = map[string]any{}

func skipped(ctx context.Context, info *grpc.UnaryServerInfo, o InterceptorOptions) bool {
	if skippedServer(ctx, info.Server) {
		return true
	}

	// skip authz for any preconfigured methods
	if _, ok := skippedMethods[info.FullMethod]; ok {
		return true
	}

	// TODO: refactor to remove this check
	return slices.Contains(o.skippedServers, info.Server)
}

func skippedServer(ctx context.Context, server any) bool {
	// if we skip authentication then we must skip authorization
	if skipSrv, ok := server.(authmiddlewaregrpc.SkipsAuthenticationServer); ok && skipSrv.SkipsAuthentication(ctx) {
		return true
	}

	if skipSrv, ok := server.(SkipsAuthorizationServer); ok && skipSrv.SkipsAuthorization(ctx) {
		return true
	}

	return false
}

func skippedStream(ctx context.Context, srv any, info *grpc.StreamServerInfo, o InterceptorOptions) bool {
	if skippedServer(ctx, srv) {
		return true
	}

	if _, ok := skippedMethods[info.FullMethod]; ok {
		return true
	}

	return slices.Contains(o.skippedServers, srv)
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

var errUnauthorized = errors.ErrUnauthorizedf("permission denied")

// authorizationRequest derives the action from the RPC method because create and
// update RPCs share the same protobuf request messages.
func authorizationRequest(fullMethod string, request flipt.Request) flipt.Request {
	switch fullMethod {
	case environments.EnvironmentsService_CreateNamespace_FullMethodName,
		environments.EnvironmentsService_CreateResource_FullMethodName:
		request.Action = flipt.ActionCreate
	}

	return request
}

func AuthorizationRequiredInterceptor(logger *zap.Logger, policyVerifier authz.Verifier, o ...containers.Option[InterceptorOptions]) grpc.UnaryServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		// skip authz for any preconfigured servers
		if skipped(ctx, info, opts) {
			logger.Debug("skipping authorization for server", zap.String("method", info.FullMethod))
			return handler(ctx, req)
		}

		ctx, err := authorize(ctx, logger, policyVerifier, info.FullMethod, req)
		if err != nil {
			return ctx, err
		}

		return handler(ctx, req)
	}
}

// AuthorizationRequiredStreamInterceptor enforces authorization for streaming
// RPCs (e.g. EvaluationSnapshotNamespaceStream). The request message is only
// available via RecvMsg, so authorization runs on the first received message.
func AuthorizationRequiredStreamInterceptor(logger *zap.Logger, policyVerifier authz.Verifier, o ...containers.Option[InterceptorOptions]) grpc.StreamServerInterceptor {
	var opts InterceptorOptions
	containers.ApplyAll(&opts, o...)

	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if skippedStream(stream.Context(), srv, info, opts) {
			logger.Debug("skipping authorization for server", zap.String("method", info.FullMethod))
			return handler(srv, stream)
		}

		return handler(srv, &authorizeStream{
			ServerStream:   stream,
			logger:         logger,
			policyVerifier: policyVerifier,
			fullMethod:     info.FullMethod,
		})
	}
}

// authorizeStream authorizes the first message received on a server stream.
type authorizeStream struct {
	grpc.ServerStream
	logger         *zap.Logger
	policyVerifier authz.Verifier
	fullMethod     string
	ctx            context.Context
	authorized     bool
}

func (s *authorizeStream) RecvMsg(m any) error {
	if err := s.ServerStream.RecvMsg(m); err != nil {
		return err
	}

	if s.authorized {
		return nil
	}

	ctx, err := authorize(s.ServerStream.Context(), s.logger, s.policyVerifier, s.fullMethod, m)
	if err != nil {
		return err
	}

	s.ctx = ctx
	s.authorized = true
	return nil
}

func (s *authorizeStream) Context() context.Context {
	if s.ctx != nil {
		return s.ctx
	}
	return s.ServerStream.Context()
}

// requestsFor derives authorization requests for req. OFREP requests carry no
// namespace/environment fields; their scope comes from the Flipt headers
// already placed on ctx by FliptHeadersInterceptor.
func requestsFor(ctx context.Context, req any) ([]flipt.Request, bool) {
	if requester, ok := req.(flipt.Requester); ok {
		return requester.Request(), true
	}

	switch req.(type) {
	case *ofrep.EvaluateFlagRequest, *ofrep.EvaluateBulkRequest:
		env, _ := common.FliptEnvironmentFromContext(ctx)
		ns, _ := common.FliptNamespaceFromContext(ctx)
		return []flipt.Request{
			flipt.NewRequest(flipt.ScopeNamespace, flipt.ActionEvaluate,
				flipt.WithEnvironment(env),
				flipt.WithNamespace(ns)),
		}, true
	}

	return nil, false
}

func authorize(ctx context.Context, logger *zap.Logger, policyVerifier authz.Verifier, fullMethod string, req any) (context.Context, error) {
	ctx = authz.ContextWithAuthorizationRequired(ctx)

	requests, ok := requestsFor(ctx, req)
	if !ok {
		logger.Error("unsupported request type for authorization", zap.String("method", fullMethod))
		return ctx, errUnauthorized
	}

	auth := authmiddlewaregrpc.GetAuthenticationFrom(ctx)
	if auth == nil {
		logger.Error("unauthorized", zap.String("reason", "authentication required"))
		return ctx, errUnauthorized
	}

	for _, request := range requests {
		request = authorizationRequest(fullMethod, request)

		// Convert the request and authentication to plain Go values once so
		// the policy engine can consume them without converting structs on
		// every evaluation.
		input := authz.Input(request, auth)

		allowed, err := policyVerifier.IsAllowed(ctx, input)
		if err != nil {
			logger.Error("unauthorized", zap.Error(err))
			return ctx, errUnauthorized
		}

		switch fullMethod {
		case environments.EnvironmentsService_ListEnvironments_FullMethodName:
			viewableEnvironments, err := policyVerifier.ViewableEnvironments(ctx, input)
			if err != nil {
				logger.Error("unauthorized", zap.Error(err))
				return ctx, errUnauthorized
			}

			// A non-empty viewable scope authorizes a partial list even when the
			// request itself is not allowed (there is no environment on a list
			// request to evaluate). An empty scope is still passed to the handler
			// so the endpoint can return an empty list rather than fail open.
			// Policies that do not define the optional scope retain the historical
			// unrestricted list behavior.
			if viewableEnvironments == nil {
				viewableEnvironments = []string{"*"}
			}
			ctx = context.WithValue(ctx, authz.EnvironmentsKey, viewableEnvironments)
			continue
		case environments.EnvironmentsService_ListNamespaces_FullMethodName:
			viewableNamespaces, err := policyVerifier.ViewableNamespaces(ctx, *request.Environment, input)
			if err != nil {
				logger.Error("unauthorized", zap.Error(err))
				return ctx, errUnauthorized
			}

			logger.Debug("policy namespaces evaluation", zap.Any("namespaces", viewableNamespaces))
			// As with environments, preserve partial access but always attach a
			// scope, including an empty one, so the endpoint cannot fail open.
			// Policies that do not define the optional scope retain the historical
			// unrestricted list behavior.
			if viewableNamespaces == nil {
				viewableNamespaces = []string{"*"}
			}
			ctx = context.WithValue(ctx, authz.NamespacesKey, viewableNamespaces)
			continue
		}

		if !allowed {
			logger.Error("unauthorized", zap.String("reason", "permission denied"))
			return ctx, errUnauthorized
		}
	}

	return ctx, nil
}
