package evaluation

import (
	"context"
	"hash/crc32"

	"go.flipt.io/flipt/errors"
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
	s.logger.Debug("evaluate", zap.Stringer("request", v))
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

	ver := &rpcEvaluation.VariantEvaluationResponse{
		Match:             resp.Match,
		SegmentKey:        resp.SegmentKey,
		Reason:            rpcEvaluation.EvaluationReason(resp.Reason),
		VariantKey:        resp.Value,
		VariantAttachment: resp.Attachment,
	}

	// add otel attributes to span
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(spanAttrs...)

	s.logger.Debug("evaluate", zap.Stringer("response", resp))
	return ver, nil
}

// Boolean evaluates a request for a boolean flag and entity.
func (s *Server) Boolean(ctx context.Context, r *rpcEvaluation.EvaluationRequest) (*rpcEvaluation.BooleanEvaluationResponse, error) {
	rollouts, err := s.store.GetEvaluationRollouts(ctx, r.FlagKey, r.NamespaceKey)
	if err != nil {
		return nil, err
	}

	resp := &rpcEvaluation.BooleanEvaluationResponse{}

	var lastRank int32

	for _, rollout := range rollouts {
		if rollout.Rank < lastRank {
			resp.Reason = rpcEvaluation.EvaluationReason_ERROR_EVALUATION_REASON
			return resp, errors.ErrInvalidf("rollout rank: %d detected out of order", rollout.Rank)
		}

		lastRank = rollout.Rank

		if rollout.Percentage != nil {
			// consistent hashing based on the entity id and flag key.
			hash := crc32.ChecksumIEEE([]byte(r.EntityId + r.FlagKey))

			normalizedValue := float32(int(hash) % 100)

			// if this case does not hold, fall through to the next rollout.
			if normalizedValue < rollout.Percentage.Percentage {
				resp.Value = rollout.Percentage.Value
				resp.Reason = rpcEvaluation.EvaluationReason_MATCH_EVALUATION_REASON
				s.logger.Debug("percentage based matched", zap.Int("rank", int(rollout.Rank)), zap.String("rollout_type", "percentage"))

				return resp, nil
			}
		} else if rollout.Segment != nil {
			matched, err := doConstraintsMatch(s.logger, r.Context, rollout.Segment.Constraints, rollout.Segment.SegmentMatchType)
			if err != nil {
				resp.Reason = rpcEvaluation.EvaluationReason_ERROR_EVALUATION_REASON
				return resp, err
			}

			// if we don't match the segment, fall through to the next rollout.
			if !matched {
				continue
			}

			resp.Value = rollout.Segment.Value
			resp.Reason = rpcEvaluation.EvaluationReason_MATCH_EVALUATION_REASON

			s.logger.Debug("segment based matched", zap.Int("rank", int(rollout.Rank)), zap.String("segment", rollout.Segment.SegmentKey))

			return resp, nil
		}
	}

	f, err := s.store.GetFlag(ctx, r.NamespaceKey, r.FlagKey)
	if err != nil {
		resp.Reason = rpcEvaluation.EvaluationReason_ERROR_EVALUATION_REASON
		return resp, err
	}

	// If we have exhausted all rollouts and we still don't have a match, return the default value.
	resp.Reason = rpcEvaluation.EvaluationReason_DEFAULT_EVALUATION_REASON
	resp.Value = f.Enabled
	s.logger.Debug("default rollout matched", zap.Bool("value", f.Enabled))

	return resp, nil
}
