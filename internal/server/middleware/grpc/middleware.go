package grpc_middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	errs "go.flipt.io/flipt/errors"
	cctx "go.flipt.io/flipt/internal/common"
	"go.flipt.io/flipt/internal/server/analytics"
	"go.flipt.io/flipt/internal/server/common"
	"go.flipt.io/flipt/internal/server/metrics"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	return handler(srv, stream)
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
func EvaluationUnaryInterceptor(analyticsEnabled bool) grpc.UnaryServerInterceptor {
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

			if analyticsEnabled {
				span := trace.SpanFromContext(ctx)

				switch r := resp.(type) {
				case *evaluation.VariantEvaluationResponse:
					// This "should" always be an evalution request under these circumstances.
					if evaluationRequest, ok := req.(*evaluation.EvaluationRequest); ok {
						var variantKey *string = nil
						if r.GetVariantKey() != "" {
							variantKey = &r.VariantKey
						}

						evaluationResponses := []*analytics.EvaluationResponse{
							{
								EnvironmentKey:  evaluationRequest.GetEnvironmentKey(),
								NamespaceKey:    evaluationRequest.GetNamespaceKey(),
								FlagKey:         r.GetFlagKey(),
								FlagType:        evaluation.EvaluationFlagType_VARIANT_FLAG_TYPE.String(),
								Match:           &r.Match,
								Reason:          r.GetReason().String(),
								Timestamp:       r.GetTimestamp().AsTime(),
								EvaluationValue: variantKey,
								EntityId:        evaluationRequest.EntityId,
							},
						}

						if evaluationResponsesBytes, err := json.Marshal(evaluationResponses); err == nil {
							keyValue := []attribute.KeyValue{
								{
									Key:   "flipt.evaluation.response",
									Value: attribute.StringValue(string(evaluationResponsesBytes)),
								},
							}
							span.AddEvent("evaluation_response", trace.WithAttributes(keyValue...))
						}
					}
				case *evaluation.BooleanEvaluationResponse:
					if evaluationRequest, ok := req.(*evaluation.EvaluationRequest); ok {
						evaluationValue := fmt.Sprint(r.GetEnabled())
						evaluationResponses := []*analytics.EvaluationResponse{
							{
								EnvironmentKey:  evaluationRequest.GetEnvironmentKey(),
								NamespaceKey:    evaluationRequest.GetNamespaceKey(),
								FlagKey:         r.GetFlagKey(),
								FlagType:        evaluation.EvaluationFlagType_BOOLEAN_FLAG_TYPE.String(),
								Reason:          r.GetReason().String(),
								Timestamp:       r.GetTimestamp().AsTime(),
								Match:           nil,
								EvaluationValue: &evaluationValue,
								EntityId:        evaluationRequest.EntityId,
							},
						}

						if evaluationResponsesBytes, err := json.Marshal(evaluationResponses); err == nil {
							keyValue := []attribute.KeyValue{
								{
									Key:   "flipt.evaluation.response",
									Value: attribute.StringValue(string(evaluationResponsesBytes)),
								},
							}
							span.AddEvent("evaluation_response", trace.WithAttributes(keyValue...))
						}
					}
				case *evaluation.BatchEvaluationResponse:
					if batchEvaluationRequest, ok := req.(*evaluation.BatchEvaluationRequest); ok {
						evaluationResponses := make([]*analytics.EvaluationResponse, 0, len(r.GetResponses()))
						for idx, response := range r.GetResponses() {
							switch response.GetType() {
							case evaluation.EvaluationResponseType_VARIANT_EVALUATION_RESPONSE_TYPE:
								variantResponse := response.GetVariantResponse()
								var variantKey *string = nil
								if variantResponse.GetVariantKey() != "" {
									variantKey = &variantResponse.VariantKey
								}

								evaluationResponses = append(evaluationResponses, &analytics.EvaluationResponse{
									EnvironmentKey:  batchEvaluationRequest.Requests[idx].GetEnvironmentKey(),
									NamespaceKey:    batchEvaluationRequest.Requests[idx].GetNamespaceKey(),
									FlagKey:         variantResponse.GetFlagKey(),
									FlagType:        evaluation.EvaluationFlagType_VARIANT_FLAG_TYPE.String(),
									Match:           &variantResponse.Match,
									Reason:          variantResponse.GetReason().String(),
									Timestamp:       variantResponse.Timestamp.AsTime(),
									EvaluationValue: variantKey,
									EntityId:        batchEvaluationRequest.Requests[idx].EntityId,
								})
							case evaluation.EvaluationResponseType_BOOLEAN_EVALUATION_RESPONSE_TYPE:
								booleanResponse := response.GetBooleanResponse()
								evaluationValue := fmt.Sprint(booleanResponse.GetEnabled())
								evaluationResponses = append(evaluationResponses, &analytics.EvaluationResponse{
									EnvironmentKey:  batchEvaluationRequest.Requests[idx].GetEnvironmentKey(),
									NamespaceKey:    batchEvaluationRequest.Requests[idx].GetNamespaceKey(),
									FlagKey:         booleanResponse.GetFlagKey(),
									FlagType:        evaluation.EvaluationFlagType_BOOLEAN_FLAG_TYPE.String(),
									Reason:          booleanResponse.GetReason().String(),
									Timestamp:       booleanResponse.Timestamp.AsTime(),
									Match:           nil,
									EvaluationValue: &evaluationValue,
									EntityId:        batchEvaluationRequest.Requests[idx].EntityId,
								})
							}
						}

						if evaluationResponsesBytes, err := json.Marshal(evaluationResponses); err == nil {
							keyValue := []attribute.KeyValue{
								{
									Key:   "flipt.evaluation.response",
									Value: attribute.StringValue(string(evaluationResponsesBytes)),
								},
							}
							span.AddEvent("evaluation_response", trace.WithAttributes(keyValue...))
						}
					}
				}
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
