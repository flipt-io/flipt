package grpc_middleware

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/audit"
	"go.flipt.io/flipt/internal/server/auth"
	"go.flipt.io/flipt/internal/server/cache"
	"go.flipt.io/flipt/internal/server/metrics"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	timestamp "google.golang.org/protobuf/types/known/timestamppb"
)

const (
	ipKey        = "x-forwarded-for"
	oidcEmailKey = "io.flipt.auth.oidc.email"
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

	// given already a *status.Error then forward unchanged
	if _, ok := status.FromError(err); ok {
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

// EvaluationUnaryInterceptor sets required request/response fields.
// Note: this should be added before any caching interceptor to ensure the request id/response fields are unique.
func EvaluationUnaryInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	switch r := req.(type) {
	case *flipt.EvaluationRequest:
		startTime := time.Now()

		// set request ID if not present
		if r.RequestId == "" {
			r.RequestId = uuid.Must(uuid.NewV4()).String()
		}

		resp, err = handler(ctx, req)
		if err != nil {
			return resp, err
		}

		// set response fields
		if resp != nil {
			if rr, ok := resp.(*flipt.EvaluationResponse); ok {
				rr.RequestId = r.RequestId
				rr.Timestamp = timestamp.New(time.Now().UTC())
				rr.RequestDurationMillis = float64(time.Since(startTime)) / float64(time.Millisecond)
			}
			return resp, nil
		}

	case *flipt.BatchEvaluationRequest:
		startTime := time.Now()

		// set request ID if not present
		if r.RequestId == "" {
			r.RequestId = uuid.Must(uuid.NewV4()).String()
		}

		resp, err = handler(ctx, req)
		if err != nil {
			return resp, err
		}

		// set response fields
		if resp != nil {
			if rr, ok := resp.(*flipt.BatchEvaluationResponse); ok {
				rr.RequestId = r.RequestId
				rr.RequestDurationMillis = float64(time.Since(startTime)) / float64(time.Millisecond)
				return resp, nil
			}
		}
	}

	return handler(ctx, req)
}

// CacheUnaryInterceptor caches the response of a request if the request is cacheable.
// TODO: we could clean this up by using generics in 1.18+ to avoid the type switch/duplicate code.
func CacheUnaryInterceptor(cache cache.Cacher, logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if cache == nil {
			return handler(ctx, req)
		}

		switch r := req.(type) {
		case *flipt.EvaluationRequest:
			key, err := evaluationCacheKey(r)
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
		}

		return handler(ctx, req)
	}
}

// AuditUnaryInterceptor sends audit logs to configured sinks upon successful RPC requests for auditable events.
func AuditUnaryInterceptor(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			return resp, err
		}

		// Identity metadata for audit events. We will always include the IP address in the
		// metadata configuration, and only include the email when it exists from the user logging
		// into the UI via OIDC.
		var author string
		var ipAddress string

		md, _ := metadata.FromIncomingContext(ctx)
		if len(md[ipKey]) > 0 {
			ipAddress = md[ipKey][0]
		}

		auth := auth.GetAuthenticationFrom(ctx)
		if auth != nil {
			author = auth.Metadata[oidcEmailKey]
		}

		var event *audit.Event

		switch r := req.(type) {
		case *flipt.CreateFlagRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Flag, Action: audit.Create, IP: ipAddress, Author: author}, r)
		case *flipt.UpdateFlagRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Flag, Action: audit.Update, IP: ipAddress, Author: author}, r)
		case *flipt.DeleteFlagRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Flag, Action: audit.Delete, IP: ipAddress, Author: author}, r)
		case *flipt.CreateVariantRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Variant, Action: audit.Create, IP: ipAddress, Author: author}, r)
		case *flipt.UpdateVariantRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Variant, Action: audit.Update, IP: ipAddress, Author: author}, r)
		case *flipt.DeleteVariantRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Variant, Action: audit.Delete, IP: ipAddress, Author: author}, r)
		case *flipt.CreateSegmentRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Segment, Action: audit.Create, IP: ipAddress, Author: author}, r)
		case *flipt.UpdateSegmentRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Segment, Action: audit.Update, IP: ipAddress, Author: author}, r)
		case *flipt.DeleteSegmentRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Segment, Action: audit.Delete, IP: ipAddress, Author: author}, r)
		case *flipt.CreateConstraintRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Constraint, Action: audit.Create, IP: ipAddress, Author: author}, r)
		case *flipt.UpdateConstraintRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Constraint, Action: audit.Update, IP: ipAddress, Author: author}, r)
		case *flipt.DeleteConstraintRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Constraint, Action: audit.Delete, IP: ipAddress, Author: author}, r)
		case *flipt.CreateDistributionRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Distribution, Action: audit.Create, IP: ipAddress, Author: author}, r)
		case *flipt.UpdateDistributionRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Distribution, Action: audit.Update, IP: ipAddress, Author: author}, r)
		case *flipt.DeleteDistributionRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Distribution, Action: audit.Delete, IP: ipAddress, Author: author}, r)
		case *flipt.CreateRuleRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Rule, Action: audit.Create, IP: ipAddress, Author: author}, r)
		case *flipt.UpdateRuleRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Rule, Action: audit.Update, IP: ipAddress, Author: author}, r)
		case *flipt.DeleteRuleRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Rule, Action: audit.Delete, IP: ipAddress, Author: author}, r)
		case *flipt.CreateNamespaceRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Namespace, Action: audit.Create, IP: ipAddress, Author: author}, r)
		case *flipt.UpdateNamespaceRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Namespace, Action: audit.Update, IP: ipAddress, Author: author}, r)
		case *flipt.DeleteNamespaceRequest:
			event = audit.NewEvent(audit.Metadata{Type: audit.Namespace, Action: audit.Delete, IP: ipAddress, Author: author}, r)
		}

		if event != nil {
			span := trace.SpanFromContext(ctx)
			span.AddEvent("event", trace.WithAttributes(event.DecodeToAttributes()...))
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
	var k string
	// for backward compatibility
	if namespaceKey != "" {
		k = fmt.Sprintf("f:%s:%s", namespaceKey, key)
	} else {
		k = fmt.Sprintf("f:%s", key)
	}

	return fmt.Sprintf("flipt:%x", md5.Sum([]byte(k)))
}

func evaluationCacheKey(r *flipt.EvaluationRequest) (string, error) {
	out, err := json.Marshal(r.GetContext())
	if err != nil {
		return "", fmt.Errorf("marshalling req to json: %w", err)
	}

	var k string
	// for backward compatibility
	if r.GetNamespaceKey() != "" {
		k = fmt.Sprintf("e:%s:%s:%s:%s", r.GetNamespaceKey(), r.GetFlagKey(), r.GetEntityId(), out)
	} else {
		k = fmt.Sprintf("e:%s:%s:%s", r.GetFlagKey(), r.GetEntityId(), out)
	}

	return fmt.Sprintf("flipt:%x", md5.Sum([]byte(k))), nil
}
