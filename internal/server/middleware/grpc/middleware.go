package grpc_middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/blang/semver/v4"
	"github.com/gofrs/uuid"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/analytics"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/internal/server/authn"
	"go.flipt.io/flipt/internal/server/metrics"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/auth"
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
		requestID := uuid.Must(uuid.NewV4()).String()
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

// AuditEventUnaryInterceptor captures events and adds them to the trace span to be consumed downstream.
func AuditEventUnaryInterceptor(logger *zap.Logger, eventPairChecker audit.EventPairChecker) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var requests []flipt.Request
		r, ok := req.(flipt.Requester)

		if !ok {
			return handler(ctx, req)
		}

		requests = r.Request()

		var events []*audit.Event

		actor := authn.ActorFromContext(ctx)

		defer func() {
			for _, event := range events {
				eventPair := fmt.Sprintf("%s:%s", event.Type, event.Action)

				exists := eventPairChecker.Check(eventPair)
				if exists {
					span := trace.SpanFromContext(ctx)
					span.AddEvent("event", trace.WithAttributes(event.DecodeToAttributes()...))
				}
			}
		}()

		resp, err := handler(ctx, req)
		for _, request := range requests {
			if err != nil {
				var uerr errs.ErrUnauthorized
				if errors.As(err, &uerr) {
					request.Status = flipt.StatusDenied
					events = append(events, audit.NewEvent(request, actor, nil))
				}

				continue
			}

			// Delete and Order request(s) have to be handled separately because they do not
			// return the concrete type but rather an *empty.Empty response.
			if request.Action == flipt.ActionDelete {
				events = append(events, audit.NewEvent(request, actor, r))
				continue
			}

			switch r := req.(type) {
			case *flipt.OrderRulesRequest, *flipt.OrderRolloutsRequest:
				events = append(events, audit.NewEvent(request, actor, r))
				continue
			}

			switch r := resp.(type) {
			case *flipt.Flag:
				events = append(events, audit.NewEvent(request, actor, audit.NewFlag(r)))
			case *flipt.Variant:
				events = append(events, audit.NewEvent(request, actor, audit.NewVariant(r)))
			case *flipt.Segment:
				events = append(events, audit.NewEvent(request, actor, audit.NewSegment(r)))
			case *flipt.Distribution:
				events = append(events, audit.NewEvent(request, actor, audit.NewDistribution(r)))
			case *flipt.Constraint:
				events = append(events, audit.NewEvent(request, actor, audit.NewConstraint(r)))
			case *flipt.Namespace:
				events = append(events, audit.NewEvent(request, actor, audit.NewNamespace(r)))
			case *flipt.Rollout:
				events = append(events, audit.NewEvent(request, actor, audit.NewRollout(r)))
			case *flipt.Rule:
				events = append(events, audit.NewEvent(request, actor, audit.NewRule(r)))
			case *auth.CreateTokenResponse:
				events = append(events, audit.NewEvent(request, actor, r.Authentication.Metadata))
			}
		}

		return resp, err
	}
}

// x-flipt-accept-server-version represents the maximum version of the flipt server that the client can handle.
const fliptAcceptServerVersionHeaderKey = "x-flipt-accept-server-version"

type fliptAcceptServerVersionContextKey struct{}

// WithFliptAcceptServerVersion sets the flipt version in the context.
func WithFliptAcceptServerVersion(ctx context.Context, version semver.Version) context.Context {
	return context.WithValue(ctx, fliptAcceptServerVersionContextKey{}, version)
}

// The last version that does not support the x-flipt-accept-server-version header.
var preFliptAcceptServerVersion = semver.MustParse("1.37.1")

// FliptAcceptServerVersionFromContext returns the flipt-accept-server-version from the context if it exists or the default version.
func FliptAcceptServerVersionFromContext(ctx context.Context) semver.Version {
	v, ok := ctx.Value(fliptAcceptServerVersionContextKey{}).(semver.Version)
	if !ok {
		return preFliptAcceptServerVersion
	}
	return v
}

// FliptAcceptServerVersionUnaryInterceptor is a grpc client interceptor that sets the flipt-accept-server-version in the context if provided in the metadata/header.
func FliptAcceptServerVersionUnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return handler(ctx, req)
		}

		if fliptVersionHeader := md.Get(fliptAcceptServerVersionHeaderKey); len(fliptVersionHeader) > 0 {
			version := fliptVersionHeader[0]
			if version != "" {
				cv, err := semver.ParseTolerant(version)
				if err != nil {
					logger.Warn("parsing x-flipt-accept-server-version header", zap.String("version", version), zap.Error(err))
					return handler(ctx, req)
				}

				logger.Debug("x-flipt-accept-server-version header", zap.String("version", version))
				ctx = WithFliptAcceptServerVersion(ctx, cv)
			}
		}

		return handler(ctx, req)
	}
}

// ForwardFliptAcceptServerVersion extracts the "x-flipt-accept-server-version"" header from an HTTP request
// and forwards them as grpc metadata entries.
func ForwardFliptAcceptServerVersion(ctx context.Context, req *http.Request) metadata.MD {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.MD{}
	}
	values := req.Header.Values(fliptAcceptServerVersionHeaderKey)
	if len(values) > 0 {
		md[fliptAcceptServerVersionHeaderKey] = values
	}
	return md
}
