package grpc_middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	grpcinterceptors "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	grpclogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"

	errs "go.flipt.io/flipt/errors"
	cctx "go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/server/common"
	"go.flipt.io/flipt/internal/server/metrics"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ValidationUnaryInterceptor validates incoming requests
func ValidationUnaryInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	if v, ok := req.(flipt.Validator); ok {
		if err := v.Validate(); err != nil {
			return nil, err
		}
	}

	return handler(ctx, req)
}

// ErrorUnaryInterceptor intercepts known errors and returns the appropriate GRPC status code
func ErrorUnaryInterceptor(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	resp, err = handler(ctx, req)
	if err != nil {
		return resp, handleError(ctx, err)
	}

	return resp, nil
}

// ErrorStreamInterceptor intercepts known errors and returns the appropriate GRPC status code
func ErrorStreamInterceptor(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	err := handler(srv, stream)
	if err != nil {
		return handleError(stream.Context(), err)
	}

	return nil
}

func handleError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}

	metrics.ErrorsTotal.Add(ctx, 1)

	// given already a *status.Error then forward unchanged
	if _, ok := status.FromError(err); ok {
		return err
	}

	if errors.Is(err, context.Canceled) {
		return status.Error(codes.Canceled, err.Error())
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return status.Error(codes.DeadlineExceeded, err.Error())
	}

	code := codes.Internal
	switch {
	case errs.AsMatch[errs.ErrNotFound](err):
		code = codes.NotFound
	case errs.AsMatch[errs.ErrInvalid](err),
		errs.AsMatch[errs.ErrValidation](err):
		code = codes.InvalidArgument
	case errs.AsMatch[errs.ErrUnauthenticated](err):
		code = codes.Unauthenticated
	case errs.AsMatch[errs.ErrUnauthorized](err):
		code = codes.PermissionDenied
	case errs.AsMatch[errs.ErrAlreadyExists](err):
		code = codes.AlreadyExists
	case errs.AsMatch[errs.ErrConflict](err):
		code = codes.Aborted
	case errs.AsMatch[errs.ErrNotImplemented](err):
		code = codes.Unimplemented
	case errs.AsMatch[errs.ErrNotModified](err):
		// special case: only supported via HTTP / Gateway
		// we set the response to OK, but override the http status code to 304 (Not Modified)
		code = codes.OK
		_ = grpc.SetHeader(ctx, metadata.Pairs("x-http-code", "304"))
	}

	return status.Error(code, err.Error())
}

type RequestIdentifiable interface {
	// SetRequestIDIfNotBlank attempts to set the provided ID on the instance
	// If the ID was blank, it returns the ID provided to this call.
	// If the ID was not blank, it returns the ID found on the instance.
	SetRequestIDIfNotBlank(id string) string
}

type ResponseDurationRecordable interface {
	// SetTimestamps records the start and end times on the target instance.
	SetTimestamps(start, end time.Time)
}

// FliptHeadersUnaryInterceptor intercepts incoming requests and adds the flipt environment and namespace to the context.
func FliptHeadersUnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Debug("no metadata found in context")
			return handler(ctx, req)
		}

		ctx = contextWithMetadata(ctx, md, logger)
		return handler(ctx, req)
	}
}

// FliptHeadersStreamInterceptor intercepts incoming requests and adds the flipt environment and namespace to the context.
func FliptHeadersStreamInterceptor(logger *zap.Logger) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(stream.Context())
		if !ok {
			logger.Debug("no metadata found in context")
			return handler(srv, stream)
		}

		ctx := contextWithMetadata(stream.Context(), md, logger)
		return handler(srv, &grpcmiddleware.WrappedServerStream{
			ServerStream:   stream,
			WrappedContext: ctx,
		})
	}
}

// contextWithMetadata adds the flipt environment and namespace to the context if they are present in the metadata.
func contextWithMetadata(ctx context.Context, md metadata.MD, logger *zap.Logger) context.Context {
	if fliptEnvironment := md.Get(common.HeaderFliptEnvironment); len(fliptEnvironment) > 0 {
		environment := fliptEnvironment[0]
		if environment != "" {
			logger.Debug("setting flipt environment in request context", zap.String("environment", environment))
			ctx = cctx.WithFliptEnvironment(ctx, environment)
		}
	}

	if fliptNamespace := md.Get(common.HeaderFliptNamespace); len(fliptNamespace) > 0 {
		namespace := fliptNamespace[0]
		if namespace != "" {
			logger.Debug("setting flipt namespace in request context", zap.String("namespace", namespace))
			ctx = cctx.WithFliptNamespace(ctx, namespace)
		}
	}

	return ctx
}

// EvaluationUnaryInterceptor sets required request/response fields.
// Note: this should be added before any caching interceptor to ensure the request id/response fields are unique.
// Note: this should be added after the FliptHeadersInterceptor to ensure the environment and namespace are set in the context.
func EvaluationUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		startTime := time.Now().UTC()

		if r, ok := req.(*evaluation.EvaluationRequest); ok {
			environmentKey, _ := cctx.FliptEnvironmentFromContext(ctx)
			namespaceKey, _ := cctx.FliptNamespaceFromContext(ctx)
			r.SetEnvironmentKeyIfNotBlank(environmentKey)
			r.SetNamespaceKeyIfNotBlank(namespaceKey)
		}

		// set request ID if not present
		requestID := uuid.NewString()
		if r, ok := req.(RequestIdentifiable); ok {
			requestID = r.SetRequestIDIfNotBlank(requestID)

			resp, err = handler(ctx, req)
			if err != nil {
				return resp, err
			}

			// set request ID on response
			if r, ok := resp.(RequestIdentifiable); ok {
				_ = r.SetRequestIDIfNotBlank(requestID)
			}

			// record start, end, duration on response types
			if r, ok := resp.(ResponseDurationRecordable); ok {
				r.SetTimestamps(startTime, time.Now().UTC())
			}

			return resp, nil
		}

		return handler(ctx, req)
	}
}

// ForwardFliptEnvironment extracts the "x-flipt-environment" header from an HTTP request
// and forwards them as grpc metadata entries.
func ForwardFliptEnvironment(ctx context.Context, req *http.Request) metadata.MD {
	return forwardHeader(ctx, req, common.HeaderFliptEnvironment)
}

// ForwardFliptNamespace extracts the "x-flipt-namespace" header from an HTTP request
// and forwards them as grpc metadata entries.
func ForwardFliptNamespace(ctx context.Context, req *http.Request) metadata.MD {
	return forwardHeader(ctx, req, common.HeaderFliptNamespace)
}

func forwardHeader(ctx context.Context, req *http.Request, headerKey string) metadata.MD {
	headerKey = strings.ToLower(headerKey)
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	values := req.Header.Values(headerKey)
	if len(values) > 0 {
		md[headerKey] = values
	}
	return md
}

// ZapInterceptorLogger wraps a zap.Logger to conform to the grpclogging.Logger interface.
// It translates grpclogging log levels to zap log levels and properly formats log fields.
func ZapInterceptorLogger(l *zap.Logger) grpclogging.Logger {
	l = l.WithOptions(zap.AddCallerSkip(1))
	return grpclogging.LoggerFunc(func(ctx context.Context, lvl grpclogging.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.With(f...)

		switch lvl {
		case grpclogging.LevelDebug:
			logger.Debug(msg)
		case grpclogging.LevelInfo:
			logger.Info(msg)
		case grpclogging.LevelWarn:
			logger.Warn(msg)
		case grpclogging.LevelError:
			logger.Error(msg)
		default:
			logger.Info(msg)
		}
	})
}

// LoggingSelector is a selector.Matcher that returns false for health check calls,
// excluding them from gRPC logging. This prevents noisy health check logs.
func LoggingSelector(_ context.Context, c grpcinterceptors.CallMeta) bool {
	return c.FullMethod() != "/grpc.health.v1.Health/Check"
}

// ChainUnaryServer creates a single interceptor out of a chain of many interceptors.
//
// Execution is done in left-to-right order, including passing of context.
// For example ChainUnaryServer([]{one, two, three}) will execute one before two before three, and three
// will see context changes of one and two.
//
// While this can be useful in some scenarios, it is generally advisable to use
// google.golang.org/grpc.ChainUnaryInterceptor directly.
func ChainUnaryServer(interceptors []grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return interceptors[0](ctx, req, info, getChainUnaryHandler(interceptors, 0, info, handler))
	}
}

// getChainUnaryHandler is a recursive helper that builds the interceptor chain handler.
// It returns the final handler when reaching the last interceptor, otherwise wraps
// the next interceptor in a closure.
func getChainUnaryHandler(interceptors []grpc.UnaryServerInterceptor, curr int, info *grpc.UnaryServerInfo, finalHandler grpc.UnaryHandler) grpc.UnaryHandler {
	if curr == len(interceptors)-1 {
		return finalHandler
	}
	return func(ctx context.Context, req any) (any, error) {
		return interceptors[curr+1](ctx, req, info, getChainUnaryHandler(interceptors, curr+1, info, finalHandler))
	}
}

// ChainStreamServer creates a single interceptor out of a chain of many stream interceptors.
//
// Execution is done in left-to-right order, including passing of context.
// For example ChainStreamServer([]{one, two, three}) will execute one before two before three.
// If you want to pass context between interceptors, use WrapServerStream.
//
// While this can be useful in some scenarios, it is generally advisable to use
// google.golang.org/grpc.ChainStreamInterceptor directly.
func ChainStreamServer(interceptors []grpc.StreamServerInterceptor) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return interceptors[0](srv, ss, info, getChainStreamHandler(interceptors, 0, info, handler))
	}
}

// getChainStreamHandler is a recursive helper that builds the stream interceptor chain handler.
// It returns the final handler when reaching the last interceptor, otherwise wraps
// the next interceptor in a closure.
func getChainStreamHandler(interceptors []grpc.StreamServerInterceptor, curr int, info *grpc.StreamServerInfo, finalHandler grpc.StreamHandler) grpc.StreamHandler {
	if curr == len(interceptors)-1 {
		return finalHandler
	}
	return func(srv any, ss grpc.ServerStream) error {
		return interceptors[curr+1](srv, ss, info, getChainStreamHandler(interceptors, curr+1, info, finalHandler))
	}
}
