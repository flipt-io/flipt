package evaluation

import (
	"context"
	"errors"

	errs "go.flipt.io/flipt/errors"
	fliptotel "go.flipt.io/flipt/internal/server/otel"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	rpcEvaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Variant evaluates a request for a multi-variate flag and entity.
func (s *Server) Variant(ctx context.Context, v *rpcEvaluation.EvaluationRequest) (*rpcEvaluation.VariantEvaluationResponse, error) {
	var ver = &rpcEvaluation.VariantEvaluationResponse{}

	flag, err := s.store.GetFlag(ctx, v.NamespaceKey, v.FlagKey)
	if err != nil {
		var errnf errs.ErrNotFound

		if errors.As(err, &errnf) {
			ver.Reason = rpcEvaluation.EvaluationReason_FLAG_NOT_FOUND_EVALUATION_REASON
			return ver, err
		}

		ver.Reason = rpcEvaluation.EvaluationReason_ERROR_EVALUATION_REASON
		return ver, err
	}

	if flag.Type != flipt.FlagType_VARIANT_FLAG_TYPE {
		ver.Reason = rpcEvaluation.EvaluationReason_ERROR_EVALUATION_REASON
		return ver, errs.ErrInvalidf("flag type %s invalid", flipt.FlagType_name[int32(flag.Type)])
	}

	s.logger.Debug("variant", zap.Stringer("request", v))
	if v.NamespaceKey == "" {
		v.NamespaceKey = storage.DefaultNamespace
	}

	resp, err := s.evaluator.Evaluate(ctx, &flipt.EvaluationRequest{
		RequestId:    v.RequestId,
		FlagKey:      v.FlagKey,
		EntityId:     v.EntityId,
		Context:      v.Context,
		NamespaceKey: v.NamespaceKey,
	})
	if err != nil {
		ver.Reason = rpcEvaluation.EvaluationReason_ERROR_EVALUATION_REASON
		return ver, err
	}

	spanAttrs := []attribute.KeyValue{
		fliptotel.AttributeNamespace.String(v.NamespaceKey),
		fliptotel.AttributeFlag.String(v.FlagKey),
		fliptotel.AttributeEntityID.String(v.EntityId),
		fliptotel.AttributeRequestID.String(v.RequestId),
	}

	if resp != nil {
		spanAttrs = append(spanAttrs,
			fliptotel.AttributeMatch.Bool(resp.Match),
			fliptotel.AttributeValue.String(resp.Value),
			fliptotel.AttributeReason.String(resp.Reason.String()),
			fliptotel.AttributeSegment.String(resp.SegmentKey),
		)
	}

	ver.Match = resp.Match
	ver.SegmentKey = resp.SegmentKey
	ver.Reason = rpcEvaluation.EvaluationReason(resp.Reason)
	ver.VariantKey = resp.Value
	ver.VariantAttachment = resp.Attachment

	// add otel attributes to span
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(spanAttrs...)

	s.logger.Debug("variant", zap.Stringer("response", resp))
	return ver, nil
}
