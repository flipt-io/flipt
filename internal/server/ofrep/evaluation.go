package ofrep

import (
	"context"
	"strings"

	"github.com/google/uuid"
	flipterrors "go.flipt.io/flipt/errors"
	"go.uber.org/zap"

	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.flipt.io/flipt/rpc/flipt/ofrep"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

const ofrepCtxTargetingKey = "targetingKey"

func (s *Server) EvaluateFlag(ctx context.Context, r *ofrep.EvaluateFlagRequest) (*ofrep.EvaluatedFlag, error) {
	s.logger.Debug("ofrep flag", zap.Stringer("request", r))
	if r.Key == "" {
		return nil, newFlagMissingError()
	}
	entityId := getTargetingKey(r.Context)
	output, err := s.bridge.OFREPFlagEvaluation(ctx, EvaluationBridgeInput{
		FlagKey:      r.Key,
		NamespaceKey: getNamespace(ctx),
		EntityId:     entityId,
		Context:      r.Context,
	})
	if err != nil {
		return nil, transformError(r.Key, err)
	}

	resp, err := transformOutput(output)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("ofrep flag", zap.Stringer("response", resp))

	return resp, nil
}

func (s *Server) EvaluateBulk(ctx context.Context, r *ofrep.EvaluateBulkRequest) (*ofrep.BulkEvaluationResponse, error) {
	s.logger.Debug("ofrep bulk", zap.Stringer("request", r))
	entityId := getTargetingKey(r.Context)
	flagKeys, ok := r.Context["flags"]
	if !ok {
		return nil, newFlagsMissingError()
	}
	namespaceKey := getNamespace(ctx)
	keys := strings.Split(flagKeys, ",")
	flags := make([]*ofrep.EvaluatedFlag, 0, len(keys))
	for _, key := range keys {
		key = strings.TrimSpace(key)
		o, err := s.bridge.OFREPFlagEvaluation(ctx, EvaluationBridgeInput{
			FlagKey:      key,
			NamespaceKey: namespaceKey,
			EntityId:     entityId,
			Context:      r.Context,
		})
		if err != nil {
			return nil, transformError(key, err)
		}

		evaluation, err := transformOutput(o)
		if err != nil {
			return nil, transformError(key, err)
		}
		flags = append(flags, evaluation)
	}
	resp := &ofrep.BulkEvaluationResponse{
		Flags: flags,
	}
	s.logger.Debug("ofrep bulk", zap.Stringer("response", resp))
	return resp, nil
}

func transformOutput(output EvaluationBridgeOutput) (*ofrep.EvaluatedFlag, error) {
	value, err := structpb.NewValue(output.Value)
	if err != nil {
		return nil, err
	}
	metadata, err := structpb.NewStruct(output.Metadata)
	if err != nil {
		return nil, err
	}

	return &ofrep.EvaluatedFlag{
		Key:      output.FlagKey,
		Reason:   transformReason(output.Reason),
		Variant:  output.Variant,
		Value:    value,
		Metadata: metadata,
	}, nil
}

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
		return "default"
	}

	namespace := md.Get("x-flipt-namespace")
	if len(namespace) == 0 {
		return "default"
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
