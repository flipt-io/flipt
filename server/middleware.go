package server

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/sirupsen/logrus"
	errs "go.flipt.io/flipt/errors"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/server/cache"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
	timestamp "google.golang.org/protobuf/types/known/timestamppb"
)

// ValidationUnaryInterceptor validates incomming requests
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

	errorsTotal.Inc()

	var errnf errs.ErrNotFound
	if errors.As(err, &errnf) {
		err = status.Error(codes.NotFound, err.Error())
		return
	}

	var errin errs.ErrInvalid
	if errors.As(err, &errin) {
		err = status.Error(codes.InvalidArgument, err.Error())
		return
	}

	var errv errs.ErrValidation
	if errors.As(err, &errv) {
		err = status.Error(codes.InvalidArgument, err.Error())
		return
	}

	err = status.Error(codes.Internal, err.Error())
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
func CacheUnaryInterceptor(cache cache.Cacher, logger logrus.FieldLogger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if cache == nil {
			return handler(ctx, req)
		}

		switch r := req.(type) {
		case *flipt.EvaluationRequest:
			key, err := evaluationCacheKey(r)
			if err != nil {
				logger.WithError(err).Error("getting cache key")
				return handler(ctx, req)
			}

			cached, ok, err := cache.Get(ctx, key)
			if err != nil {
				// if error, log and without cache
				logger.WithError(err).Error("getting from cache")
				return handler(ctx, req)
			}

			if ok {
				resp := &flipt.EvaluationResponse{}
				if err := proto.Unmarshal(cached, resp); err != nil {
					logger.WithError(err).Error("unmarshalling from cache")
					return handler(ctx, req)
				}

				logger.Debugf("evaluate cache hit: %+v", resp)
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
				logger.WithError(merr).Error("marshalling for cache")
				return resp, err
			}

			// set in cache
			if cerr := cache.Set(ctx, key, data); cerr != nil {
				logger.WithError(cerr).Error("setting in cache")
			}

			return resp, err

		case *flipt.GetFlagRequest:
			key := flagCacheKey(r.GetKey())

			cached, ok, err := cache.Get(ctx, key)
			if err != nil {
				// if error, log and continue without cache
				logger.WithError(err).Error("getting from cache")
				return handler(ctx, req)
			}

			if ok {
				// if cached, return it
				flag := &flipt.Flag{}
				if err := proto.Unmarshal(cached, flag); err != nil {
					logger.WithError(err).Error("unmarshalling from cache")
					return handler(ctx, req)
				}

				logger.Debugf("flag cache hit: %+v", flag)
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
				logger.WithError(merr).Error("marshalling for cache")
				return resp, err
			}

			// set in cache
			if cerr := cache.Set(ctx, key, data); cerr != nil {
				logger.WithError(cerr).Error("setting in cache")
			}

			return resp, err

		case *flipt.UpdateFlagRequest, *flipt.DeleteFlagRequest:
			// need to do this assertion because the request type is not known in this block
			keyer := r.(flagKeyer)
			// delete from cache
			if err := cache.Delete(ctx, flagCacheKey(keyer.GetKey())); err != nil {
				logger.WithError(err).Error("deleting from cache")
			}
		case *flipt.CreateVariantRequest, *flipt.UpdateVariantRequest, *flipt.DeleteVariantRequest:
			// need to do this assertion because the request type is not known in this block
			keyer := r.(variantFlagKeyger)
			// delete from cache
			if err := cache.Delete(ctx, flagCacheKey(keyer.GetFlagKey())); err != nil {
				logger.WithError(err).Error("deleting from cache")
			}
		}

		return handler(ctx, req)
	}
}

type flagKeyer interface {
	GetKey() string
}

type variantFlagKeyger interface {
	GetFlagKey() string
}

func flagCacheKey(key string) string {
	k := fmt.Sprintf("f:%s", key)
	return fmt.Sprintf("flipt:%x", md5.Sum([]byte(k))) //nolint:gosec
}

func evaluationCacheKey(r *flipt.EvaluationRequest) (string, error) {
	out, err := json.Marshal(r.GetContext())
	if err != nil {
		return "", fmt.Errorf("marshalling req to json: %w", err)
	}

	k := fmt.Sprintf("e:%s:%s:%s", r.GetFlagKey(), r.GetEntityId(), out)
	return fmt.Sprintf("flipt:%x", md5.Sum([]byte(k))), nil //nolint:gosec
}
