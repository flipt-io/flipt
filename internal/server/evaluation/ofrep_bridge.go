package evaluation

import (
	"context"
	"errors"
	"strconv"

	"github.com/google/uuid"

	"go.flipt.io/flipt/internal/server/ofrep"

	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"

	fliptotel "go.flipt.io/flipt/internal/server/otel"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.opentelemetry.io/otel/trace"
)

const ofrepCtxTargetingKey = "targetingKey"

func (s *Server) OFREPEvaluationBridge(ctx context.Context, input ofrep.EvaluationBridgeInput) (ofrep.EvaluationBridgeOutput, error) {
	flag, err := s.store.GetFlag(ctx, storage.NewResource(input.NamespaceKey, input.FlagKey))
	if err != nil {
		return ofrep.EvaluationBridgeOutput{}, err
	}

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(
		fliptotel.AttributeNamespace.String(input.NamespaceKey),
		fliptotel.AttributeFlag.String(input.FlagKey),
		fliptotel.AttributeProviderName,
	)

	req := &rpcevaluation.EvaluationRequest{
		NamespaceKey: input.NamespaceKey,
		FlagKey:      input.FlagKey,
		EntityId:     uuid.NewString(),
		Context:      input.Context,
	}

	// https://openfeature.dev/docs/reference/concepts/evaluation-context/#targeting-key
	if targetingKey, ok := input.Context[ofrepCtxTargetingKey]; ok {
		req.EntityId = targetingKey
	}

	switch flag.Type {
	case flipt.FlagType_VARIANT_FLAG_TYPE:
		resp, err := s.variant(ctx, flag, req)
		if err != nil {
			return ofrep.EvaluationBridgeOutput{}, err
		}

		span.SetAttributes(
			fliptotel.AttributeMatch.Bool(resp.Match),
			fliptotel.AttributeValue.String(resp.VariantKey),
			fliptotel.AttributeReason.String(resp.Reason.String()),
			fliptotel.AttributeSegments.StringSlice(resp.SegmentKeys),
			fliptotel.AttributeFlagKey(resp.FlagKey),
			fliptotel.AttributeFlagVariant(resp.VariantKey),
		)

		return ofrep.EvaluationBridgeOutput{
			FlagKey: resp.FlagKey,
			Reason:  resp.Reason,
			Variant: resp.VariantKey,
			Value:   resp.VariantKey,
		}, nil
	case flipt.FlagType_BOOLEAN_FLAG_TYPE:
		resp, err := s.boolean(ctx, flag, req)
		if err != nil {
			return ofrep.EvaluationBridgeOutput{}, err
		}

		span.SetAttributes(
			fliptotel.AttributeValue.Bool(resp.Enabled),
			fliptotel.AttributeReason.String(resp.Reason.String()),
			fliptotel.AttributeFlagVariant(strconv.FormatBool(resp.Enabled)),
		)

		return ofrep.EvaluationBridgeOutput{
			FlagKey: resp.FlagKey,
			Variant: strconv.FormatBool(resp.Enabled),
			Reason:  resp.Reason,
			Value:   resp.Enabled,
		}, nil
	default:
		return ofrep.EvaluationBridgeOutput{}, errors.New("unsupported flag type for ofrep bridge")
	}
}
