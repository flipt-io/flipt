package grpc_middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/blang/semver/v4"
	"github.com/gofrs/uuid"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/cache"
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
	"google.golang.org/protobuf/proto"
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

var (
	legacyEvalCachePrefix evaluationCacheKey[*flipt.EvaluationRequest]      = "ev1"
	newEvalCachePrefix    evaluationCacheKey[*evaluation.EvaluationRequest] = "ev2"
)

// CacheUnaryInterceptor caches the response of a request if the request is cacheable.
// TODO: we could clean this up by using generics in 1.18+ to avoid the type switch/duplicate code.
func CacheUnaryInterceptor(cache cache.Cacher, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if cache == nil {
			return handler(ctx, req)
		}

		switch r := req.(type) {
		case *flipt.EvaluationRequest:
			key, err := legacyEvalCachePrefix.Key(r)
			if err != nil {
				logger.Error("getting cache key", zap.Error(err))
				return handler(ctx, req)
			}

			cached, ok, err := cache.Get(ctx, key)
			if err != nil {
				// if error, log and without cache
				logger.Error("getting from cache", zap.Error(err))
				return handler(ctx, req)
			}

			if ok {
				resp := &flipt.EvaluationResponse{}
				if err := proto.Unmarshal(cached, resp); err != nil {
					logger.Error("unmarshalling from cache", zap.Error(err))
					return handler(ctx, req)
				}

				logger.Debug("evaluate cache hit", zap.Stringer("response", resp))
				return resp, nil
			}

			logger.Debug("evaluate cache miss")
			resp, err := handler(ctx, req)
			if err != nil {
				return resp, err
			}

			// marshal response
			data, merr := proto.Marshal(resp.(*flipt.EvaluationResponse))
			if merr != nil {
				logger.Error("marshalling for cache", zap.Error(err))
				return resp, err
			}

			// set in cache
			if cerr := cache.Set(ctx, key, data); cerr != nil {
				logger.Error("setting in cache", zap.Error(err))
			}

			return resp, err

		case *flipt.GetFlagRequest:
			key := flagCacheKey(r.GetNamespaceKey(), r.GetKey())

			cached, ok, err := cache.Get(ctx, key)
			if err != nil {
				// if error, log and continue without cache
				logger.Error("getting from cache", zap.Error(err))
				return handler(ctx, req)
			}

			if ok {
				// if cached, return it
				flag := &flipt.Flag{}
				if err := proto.Unmarshal(cached, flag); err != nil {
					logger.Error("unmarshalling from cache", zap.Error(err))
					return handler(ctx, req)
				}

				logger.Debug("flag cache hit", zap.Stringer("flag", flag))
				return flag, nil
			}

			logger.Debug("flag cache miss")
			resp, err := handler(ctx, req)
			if err != nil {
				return nil, err
			}

			// marshal response
			data, merr := proto.Marshal(resp.(*flipt.Flag))
			if merr != nil {
				logger.Error("marshalling for cache", zap.Error(err))
				return resp, err
			}

			// set in cache
			if cerr := cache.Set(ctx, key, data); cerr != nil {
				logger.Error("setting in cache", zap.Error(err))
			}

			return resp, err

		case *flipt.UpdateFlagRequest, *flipt.DeleteFlagRequest:
			// need to do this assertion because the request type is not known in this block
			keyer := r.(flagKeyer)
			// delete from cache
			if err := cache.Delete(ctx, flagCacheKey(keyer.GetNamespaceKey(), keyer.GetKey())); err != nil {
				logger.Error("deleting from cache", zap.Error(err))
			}
		case *flipt.CreateVariantRequest, *flipt.UpdateVariantRequest, *flipt.DeleteVariantRequest:
			// need to do this assertion because the request type is not known in this block
			keyer := r.(variantFlagKeyger)
			// delete from cache
			if err := cache.Delete(ctx, flagCacheKey(keyer.GetNamespaceKey(), keyer.GetFlagKey())); err != nil {
				logger.Error("deleting from cache", zap.Error(err))
			}
		case *evaluation.EvaluationRequest:
			key, err := newEvalCachePrefix.Key(r)
			if err != nil {
				logger.Error("getting cache key", zap.Error(err))
				return handler(ctx, req)
			}

			cached, ok, err := cache.Get(ctx, key)
			if err != nil {
				// if error, log and without cache
				logger.Error("getting from cache", zap.Error(err))
				return handler(ctx, req)
			}

			if ok {
				resp := &evaluation.EvaluationResponse{}
				if err := proto.Unmarshal(cached, resp); err != nil {
					logger.Error("unmarshalling from cache", zap.Error(err))
					return handler(ctx, req)
				}

				logger.Debug("evaluate cache hit", zap.Stringer("response", resp))
				switch r := resp.Response.(type) {
				case *evaluation.EvaluationResponse_VariantResponse:
					return r.VariantResponse, nil
				case *evaluation.EvaluationResponse_BooleanResponse:
					return r.BooleanResponse, nil
				default:
					logger.Error("unexpected eval cache response type", zap.String("type", fmt.Sprintf("%T", resp.Response)))
				}

				return handler(ctx, req)
			}

			logger.Debug("evaluate cache miss")
			resp, err := handler(ctx, req)
			if err != nil {
				return resp, err
			}

			evalResponse := &evaluation.EvaluationResponse{}
			switch r := resp.(type) {
			case *evaluation.VariantEvaluationResponse:
				evalResponse.Type = evaluation.EvaluationResponseType_VARIANT_EVALUATION_RESPONSE_TYPE
				evalResponse.Response = &evaluation.EvaluationResponse_VariantResponse{
					VariantResponse: r,
				}
			case *evaluation.BooleanEvaluationResponse:
				evalResponse.Type = evaluation.EvaluationResponseType_BOOLEAN_EVALUATION_RESPONSE_TYPE
				evalResponse.Response = &evaluation.EvaluationResponse_BooleanResponse{
					BooleanResponse: r,
				}
			}

			// marshal response
			data, merr := proto.Marshal(evalResponse)
			if merr != nil {
				logger.Error("marshalling for cache", zap.Error(err))
				return resp, err
			}

			// set in cache
			if cerr := cache.Set(ctx, key, data); cerr != nil {
				logger.Error("setting in cache", zap.Error(err))
			}

			return resp, err
		}

		return handler(ctx, req)
	}
}

// EventPairChecker is the middleware side contract for checking if an event pair exists.
type EventPairChecker interface {
	Check(eventPair string) bool
}

// AuditEventUnaryInterceptor captures events and adds them to the trace span to be consumed downstream.
func AuditEventUnaryInterceptor(logger *zap.Logger, eventPairChecker EventPairChecker) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var request flipt.Request
		r, ok := req.(flipt.Request)

		if !ok {
			return handler(ctx, req)
		}

		request = r

		resp, err := handler(ctx, req)
		if err != nil {
			return resp, err
		}

		var event *audit.Event

		actor := authn.ActorFromContext(ctx)

		defer func() {
			if event != nil {
				ts := string(event.Type)
				as := string(event.Action)
				eventPair := fmt.Sprintf("%s:%s", ts, as)

				exists := eventPairChecker.Check(eventPair)
				if exists {
					span := trace.SpanFromContext(ctx)
					span.AddEvent("event", trace.WithAttributes(event.DecodeToAttributes()...))
				}
			}
		}()

		// Delete request(s) have to be handled separately because they do not
		// return the concrete type but rather an *empty.Empty response.
		switch r := req.(type) {
		case *flipt.DeleteFlagRequest:
			event = audit.NewEvent(request, actor, r)
		case *flipt.DeleteVariantRequest:
			event = audit.NewEvent(request, actor, r)
		case *flipt.DeleteSegmentRequest:
			event = audit.NewEvent(request, actor, r)
		case *flipt.DeleteDistributionRequest:
			event = audit.NewEvent(request, actor, r)
		case *flipt.DeleteConstraintRequest:
			event = audit.NewEvent(request, actor, r)
		case *flipt.DeleteNamespaceRequest:
			event = audit.NewEvent(request, actor, r)
		case *flipt.OrderRulesRequest:
			event = audit.NewEvent(request, actor, r)
		case *flipt.DeleteRuleRequest:
			event = audit.NewEvent(request, actor, r)
		case *flipt.OrderRolloutsRequest:
			event = audit.NewEvent(request, actor, r)
		case *flipt.DeleteRolloutRequest:
			event = audit.NewEvent(request, actor, r)
		}

		// Short circuiting the middleware here since we have a non-nil event from
		// detecting a delete.
		if event != nil {
			return resp, err
		}

		switch r := resp.(type) {
		case *flipt.Flag:
			event = audit.NewEvent(request, actor, audit.NewFlag(r))
		case *flipt.Variant:
			event = audit.NewEvent(request, actor, audit.NewVariant(r))
		case *flipt.Segment:
			event = audit.NewEvent(request, actor, audit.NewSegment(r))
		case *flipt.Distribution:
			event = audit.NewEvent(request, actor, audit.NewDistribution(r))
		case *flipt.Constraint:
			event = audit.NewEvent(request, actor, audit.NewConstraint(r))
		case *flipt.Namespace:
			event = audit.NewEvent(request, actor, audit.NewNamespace(r))
		case *flipt.Rollout:
			event = audit.NewEvent(request, actor, audit.NewRollout(r))
		case *flipt.Rule:
			event = audit.NewEvent(request, actor, audit.NewRule(r))
		case *auth.CreateTokenResponse:
			event = audit.NewEvent(request, actor, r.Authentication.Metadata)
		}

		return resp, err
	}
}

type namespaceKeyer interface {
	GetNamespaceKey() string
}

type flagKeyer interface {
	namespaceKeyer
	GetKey() string
}

type variantFlagKeyger interface {
	namespaceKeyer
	GetFlagKey() string
}

func flagCacheKey(namespaceKey, key string) string {
	// for backward compatibility
	if namespaceKey != "" {
		return fmt.Sprintf("f:%s:%s", namespaceKey, key)
	}
	return fmt.Sprintf("f:%s", key)
}

type evaluationRequest interface {
	GetNamespaceKey() string
	GetFlagKey() string
	GetEntityId() string
	GetContext() map[string]string
}

type evaluationCacheKey[T evaluationRequest] string

func (e evaluationCacheKey[T]) Key(r T) (string, error) {
	out, err := json.Marshal(r.GetContext())
	if err != nil {
		return "", fmt.Errorf("marshalling req to json: %w", err)
	}

	// for backward compatibility
	if r.GetNamespaceKey() != "" {
		return fmt.Sprintf("%s:%s:%s:%s:%s", string(e), r.GetNamespaceKey(), r.GetFlagKey(), r.GetEntityId(), out), nil
	}

	return fmt.Sprintf("%s:%s:%s:%s", string(e), r.GetFlagKey(), r.GetEntityId(), out), nil
}

// x-flipt-accept-server-version represents the maximum version of the flipt server that the client can handle.
const fliptAcceptServerVersionHeaderKey = "x-flipt-accept-server-version"

type fliptServerVersionContextKey struct{}

// WithFliptAcceptServerVersion sets the flipt version in the context.
func WithFliptAcceptServerVersion(ctx context.Context, version semver.Version) context.Context {
	return context.WithValue(ctx, fliptServerVersionContextKey{}, version)
}

// The last version that does not support the x-flipt-accept-server-version header.
var preFliptAcceptServerVersion = semver.MustParse("1.37.1")

// FliptAcceptServerVersionFromContext returns the flipt-accept-server-version from the context if it exists or the default version.
func FliptAcceptServerVersionFromContext(ctx context.Context) semver.Version {
	v, ok := ctx.Value(fliptServerVersionContextKey{}).(semver.Version)
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
