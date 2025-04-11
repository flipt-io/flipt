package grpc_middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/analytics"
	"go.flipt.io/flipt/internal/server/common"
	"go.flipt.io/flipt/internal/server/metrics"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ValidationUnaryInterceptor validates incoming requests
func ValidationUnaryInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	if v, ok := req.(flipt.Validator); ok {
		if err := v.Validate(); err != nil {
			return nil, err
		}
	}

	return handler(ctx, req)
}

// ErrorUnaryInterceptor intercepts known errors and returns the appropriate GRPC status code
func ErrorUnaryInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	resp, err = handler(ctx, req)
	if err == nil {
		return resp, nil
	}

	metrics.ErrorsTotal.Add(ctx, 1)

	// given already a *status.Error then forward unchanged
	if _, ok := status.FromError(err); ok {
		return
	}

	if errors.Is(err, context.Canceled) {
		err = status.Error(codes.Canceled, err.Error())
		return
	}

	if errors.Is(err, context.DeadlineExceeded) {
		err = status.Error(codes.DeadlineExceeded, err.Error())
		return
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

	err = status.Error(code, err.Error())
	return
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

// EvaluationUnaryInterceptor sets required request/response fields.
// Note: this should be added before any caching interceptor to ensure the request id/response fields are unique.
func EvaluationUnaryInterceptor(analyticsEnabled bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		startTime := time.Now().UTC()

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
