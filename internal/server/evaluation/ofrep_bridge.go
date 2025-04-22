package evaluation

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	flipterrors "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/common"
	"go.flipt.io/flipt/internal/server/tracing"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/core"
	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.flipt.io/flipt/rpc/flipt/ofrep"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
)

const (
	metaKeyAttachment = "attachment"
	metaKeySegments   = "segments"
)

func (s *Server) OFREPFlagEvaluation(ctx context.Context, r *ofrep.EvaluateFlagRequest) (*ofrep.EvaluationResponse, error) {
	env := s.store.GetFromContext(ctx)

	store, err := env.EvaluationStore()
	if err != nil {
		return nil, err
	}

	if r.Key == "" {
		return nil, newFlagMissingError()
	}

	var (
		namespaceKey = getNamespace(ctx)
		entityId     = getTargetingKey(r.Context)
	)

	flag, err := store.GetFlag(ctx, storage.NewResource(namespaceKey, r.Key))
	if err != nil {
		return nil, transformError(r.Key, err)
	}
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		tracing.AttributeEnvironment.String(env.Key()),
		tracing.AttributeNamespace.String(namespaceKey),
		tracing.AttributeFlag.String(r.Key),
		tracing.AttributeProviderName,
	)

	req := &rpcevaluation.EvaluationRequest{
		EnvironmentKey: env.Key(),
		NamespaceKey:   namespaceKey,
		FlagKey:        r.Key,
		EntityId:       entityId,
		Context:        r.Context,
	}

	switch flag.Type {
	case core.FlagType_VARIANT_FLAG_TYPE:
		resp, err := s.variant(ctx, store, env, flag, req)
		if err != nil {
			return nil, transformError(r.Key, err)
		}

		span.SetAttributes(
			tracing.AttributeMatch.Bool(resp.Match),
			tracing.AttributeValue.String(resp.VariantKey),
			tracing.AttributeReason.String(resp.Reason.String()),
			tracing.AttributeSegments.StringSlice(resp.SegmentKeys),
			tracing.AttributeFlagKey(resp.FlagKey),
			tracing.AttributeFlagVariant(resp.VariantKey),
		)

		mm := map[string]any{}
		if len(resp.SegmentKeys) > 0 {
			mm[metaKeySegments] = strings.Join(resp.SegmentKeys, ",")
		}
		if resp.VariantAttachment != "" {
			mm[metaKeyAttachment] = resp.VariantAttachment
		}

		value, err := structpb.NewValue(resp.VariantKey)
		if err != nil {
			return nil, err
		}
		metadata, err := structpb.NewStruct(mm)
		if err != nil {
			return nil, err
		}
		return &ofrep.EvaluationResponse{
			Key:      resp.FlagKey,
			Reason:   transformReason(resp.Reason),
			Variant:  resp.VariantKey,
			Value:    value,
			Metadata: metadata,
		}, nil

	case core.FlagType_BOOLEAN_FLAG_TYPE:
		resp, err := s.boolean(ctx, store, env, flag, req)
		if err != nil {
			return nil, transformError(r.Key, err)
		}

		span.SetAttributes(
			tracing.AttributeValue.Bool(resp.Enabled),
			tracing.AttributeReason.String(resp.Reason.String()),
			tracing.AttributeFlagVariant(strconv.FormatBool(resp.Enabled)),
		)

		value, err := structpb.NewValue(resp.Enabled)
		if err != nil {
			return nil, err
		}

		return &ofrep.EvaluationResponse{
			Key:     resp.FlagKey,
			Reason:  transformReason(resp.Reason),
			Variant: strconv.FormatBool(resp.Enabled),
			Value:   value,
		}, nil
	default:
		return nil, errors.New("unsupported flag type for ofrep bridge")
	}
}

func (s *Server) OFREPFlagEvaluationBulk(ctx context.Context, r *ofrep.EvaluateBulkRequest) (*ofrep.BulkEvaluationResponse, error) {
	env := s.store.GetFromContext(ctx)

	store, err := env.EvaluationStore()
	if err != nil {
		return nil, err
	}

	var (
		namespaceKey = getNamespace(ctx)
		flagKeys, ok = r.Context["flags"]
		keys         = strings.Split(flagKeys, ",")
	)

	if !ok {
		flags, err := store.ListFlags(ctx, storage.ListWithOptions(storage.NewNamespace(namespaceKey)))
		if err != nil {
			return nil, err
		}

		keys = make([]string, 0, len(flags.Results))
		for _, flag := range flags.Results {
			switch flag.Type {
			case core.FlagType_BOOLEAN_FLAG_TYPE:
				keys = append(keys, flag.Key)
			case core.FlagType_VARIANT_FLAG_TYPE:
				if flag.Enabled {
					keys = append(keys, flag.Key)
				}
			}
		}
	}

	responses := make([]*ofrep.EvaluationResponse, 0, len(keys))
	for _, flagKey := range keys {
		resp, err := s.OFREPFlagEvaluation(ctx, &ofrep.EvaluateFlagRequest{
			Key:     flagKey,
			Context: r.Context,
		})
		if err != nil {
			return nil, err
		}
		responses = append(responses, resp)
	}

	return &ofrep.BulkEvaluationResponse{
		Flags: responses,
	}, nil
}

const ofrepCtxTargetingKey = "targetingKey"

func getTargetingKey(context map[string]string) string {
	// https://openfeature.dev/docs/reference/concepts/evaluation-context/#targeting-key
	if targetingKey, ok := context[ofrepCtxTargetingKey]; ok {
		return targetingKey
	}
	return uuid.NewString()
}

func getNamespace(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return flipt.DefaultNamespace
	}

	namespace := md.Get(common.HeaderFliptNamespace)
	if len(namespace) == 0 {
		return flipt.DefaultNamespace
	}

	return namespace[0]
}

func transformReason(reason rpcevaluation.EvaluationReason) ofrep.EvaluateReason {
	switch reason {
	case rpcevaluation.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON:
		return ofrep.EvaluateReason_DISABLED
	case rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON:
		return ofrep.EvaluateReason_TARGETING_MATCH
	case rpcevaluation.EvaluationReason_DEFAULT_EVALUATION_REASON:
		return ofrep.EvaluateReason_DEFAULT
	default:
		return ofrep.EvaluateReason_UNKNOWN
	}
}

func transformError(key string, err error) error {
	switch {
	case flipterrors.AsMatch[flipterrors.ErrInvalid](err):
		return newBadRequestError(key, err)
	case flipterrors.AsMatch[flipterrors.ErrValidation](err):
		return newBadRequestError(key, err)
	case flipterrors.AsMatch[flipterrors.ErrNotFound](err):
		return newFlagNotFoundError(key)
	}
	return err
}

const statusFlagKeyPointer = "ofrep-flag-key"

func statusWithKey(st *status.Status, key string) (*status.Status, error) {
	return st.WithDetails(&structpb.Struct{
		Fields: map[string]*structpb.Value{
			statusFlagKeyPointer: structpb.NewStringValue(key),
		},
	})
}

func newBadRequestError(key string, err error) error {
	v := status.New(codes.InvalidArgument, err.Error())
	v, derr := statusWithKey(v, key)
	if derr != nil {
		return status.Errorf(codes.Internal, "failed to encode not bad request error")
	}
	return v.Err()
}

func newFlagNotFoundError(key string) error {
	v := status.New(codes.NotFound, fmt.Sprintf("flag was not found %s", key))
	v, derr := statusWithKey(v, key)
	if derr != nil {
		return status.Errorf(codes.Internal, "failed to encode not found error")
	}
	return v.Err()
}

func newFlagMissingError() error {
	return status.Error(codes.InvalidArgument, "flag key was not provided")
}
