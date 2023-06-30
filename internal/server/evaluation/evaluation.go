package evaluation

import (
	"context"

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
		return nil, err
	}

	s.logger.Debug("variant", zap.Stringer("request", v))
	if v.NamespaceKey == "" {
		v.NamespaceKey = storage.DefaultNamespace
	}

	resp, err := s.evaluator.Evaluate(ctx, flag, &flipt.EvaluationRequest{
		RequestId:    v.RequestId,
		FlagKey:      v.FlagKey,
		EntityId:     v.EntityId,
		Context:      v.Context,
		NamespaceKey: v.NamespaceKey,
	})
	if err != nil {
		return nil, err
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

// Boolean evaluates a request for a boolean flag and entity.
func (s *Server) Boolean(ctx context.Context, r *rpcEvaluation.EvaluationRequest) (*rpcEvaluation.BooleanEvaluationResponse, error) {
	flag, err := s.store.GetFlag(ctx, r.NamespaceKey, r.FlagKey)
	if err != nil {
		return nil, err
	}

	if flag.Type != flipt.FlagType_BOOLEAN_FLAG_TYPE {
		return nil, errs.ErrInvalidf("flag type %s invalid", flag.Type)
	}

	rollouts, err := s.store.GetEvaluationRollouts(ctx, r.NamespaceKey, r.FlagKey)
	if err != nil {
		return nil, err
	}

	resp, err := s.evaluator.booleanMatch(r, flag.Enabled, rollouts)
	if err != nil {
		return nil, err
	}

	return resp, nil
}
