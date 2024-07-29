package ofrep

import (
	"context"

	flipterrors "go.flipt.io/flipt/errors"
	"go.uber.org/zap"

	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.flipt.io/flipt/rpc/flipt/ofrep"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *Server) EvaluateFlag(ctx context.Context, r *ofrep.EvaluateFlagRequest) (*ofrep.EvaluatedFlag, error) {
	s.logger.Debug("ofrep variant", zap.Stringer("request", r))

	if r.Key == "" {
		return nil, NewTargetingKeyMissing()
	}

	output, err := s.bridge.OFREPEvaluationBridge(ctx, EvaluationBridgeInput{
		FlagKey:      r.Key,
		NamespaceKey: getNamespace(ctx),
		Context:      r.Context,
	})
	if err != nil {
		switch {
		case flipterrors.AsMatch[flipterrors.ErrInvalid](err):
			return nil, NewBadRequestError(r.Key, err)
		case flipterrors.AsMatch[flipterrors.ErrValidation](err):
			return nil, NewBadRequestError(r.Key, err)
		case flipterrors.AsMatch[flipterrors.ErrNotFound](err):
			return nil, NewFlagNotFoundError(r.Key)
		case flipterrors.AsMatch[flipterrors.ErrUnauthenticated](err):
			return nil, NewUnauthenticatedError()
		case flipterrors.AsMatch[flipterrors.ErrUnauthorized](err):
			return nil, NewUnauthorizedError()
		}

		return nil, NewInternalServerError(err)
	}

	value, err := structpb.NewValue(output.Value)
	if err != nil {
		return nil, NewInternalServerError(err)
	}

	resp := &ofrep.EvaluatedFlag{
		Key:      output.FlagKey,
		Reason:   transformReason(output.Reason),
		Variant:  output.Variant,
		Value:    value,
		Metadata: &structpb.Struct{Fields: make(map[string]*structpb.Value)},
	}

	s.logger.Debug("ofrep variant", zap.Stringer("response", resp))

	return resp, nil
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
