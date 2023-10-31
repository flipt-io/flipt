package evaluation

import (
	"context"
	"errors"
	"fmt"
	"hash/crc32"
	"time"

	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/server/metrics"
	fliptotel "go.flipt.io/flipt/internal/server/otel"
	"go.flipt.io/flipt/rpc/flipt"
	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Variant evaluates a request for a multi-variate flag and entity.
// It adapts the 'v2' evaluation API and proxies the request to the 'v1' evaluation API.
func (s *Server) Variant(ctx context.Context, r *rpcevaluation.EvaluationRequest) (*rpcevaluation.VariantEvaluationResponse, error) {
	flag, err := s.store.GetFlag(ctx, r.NamespaceKey, r.FlagKey)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("variant", zap.Stringer("request", r))

	resp, err := s.variant(ctx, flag, r)
	if err != nil {
		return nil, err
	}

	spanAttrs := []attribute.KeyValue{
		fliptotel.AttributeNamespace.String(r.NamespaceKey),
		fliptotel.AttributeFlag.String(r.FlagKey),
		fliptotel.AttributeEntityID.String(r.EntityId),
		fliptotel.AttributeRequestID.String(r.RequestId),
		fliptotel.AttributeMatch.Bool(resp.Match),
		fliptotel.AttributeValue.String(resp.VariantKey),
		fliptotel.AttributeReason.String(resp.Reason.String()),
		fliptotel.AttributeSegments.StringSlice(resp.SegmentKeys),
	}

	// add otel attributes to span
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(spanAttrs...)

	s.logger.Debug("variant", zap.Stringer("response", resp))
	return resp, nil
}

func (s *Server) variant(ctx context.Context, flag *flipt.Flag, r *rpcevaluation.EvaluationRequest) (*rpcevaluation.VariantEvaluationResponse, error) {
	resp, err := s.evaluator.Evaluate(ctx, flag, &flipt.EvaluationRequest{
		RequestId:    r.RequestId,
		FlagKey:      r.FlagKey,
		EntityId:     r.EntityId,
		Context:      r.Context,
		NamespaceKey: r.NamespaceKey,
	})
	if err != nil {
		return nil, err
	}

	var reason rpcevaluation.EvaluationReason

	switch resp.Reason {
	case flipt.EvaluationReason_MATCH_EVALUATION_REASON:
		reason = rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON
	case flipt.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON:
		reason = rpcevaluation.EvaluationReason_FLAG_DISABLED_EVALUATION_REASON
	default:
		reason = rpcevaluation.EvaluationReason_UNKNOWN_EVALUATION_REASON
	}

	ver := &rpcevaluation.VariantEvaluationResponse{
		RequestId:         r.RequestId,
		Match:             resp.Match,
		Reason:            reason,
		VariantKey:        resp.Value,
		VariantAttachment: resp.Attachment,
		FlagKey:           resp.FlagKey,
	}

	if len(resp.SegmentKeys) > 0 {
		ver.SegmentKeys = resp.SegmentKeys
	}

	return ver, nil
}

// Boolean evaluates a request for a boolean flag and entity.
func (s *Server) Boolean(ctx context.Context, r *rpcevaluation.EvaluationRequest) (*rpcevaluation.BooleanEvaluationResponse, error) {
	flag, err := s.store.GetFlag(ctx, r.NamespaceKey, r.FlagKey)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("boolean", zap.Stringer("request", r))

	if flag.Type != flipt.FlagType_BOOLEAN_FLAG_TYPE {
		return nil, errs.ErrInvalidf("flag type %s invalid", flag.Type)
	}

	resp, err := s.boolean(ctx, flag, r)
	if err != nil {
		return nil, err
	}

	spanAttrs := []attribute.KeyValue{
		fliptotel.AttributeNamespace.String(r.NamespaceKey),
		fliptotel.AttributeFlag.String(r.FlagKey),
		fliptotel.AttributeEntityID.String(r.EntityId),
		fliptotel.AttributeRequestID.String(r.RequestId),
		fliptotel.AttributeValue.Bool(resp.Enabled),
		fliptotel.AttributeReason.String(resp.Reason.String()),
	}

	// add otel attributes to span
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(spanAttrs...)

	s.logger.Debug("boolean", zap.Stringer("response", resp))
	return resp, nil
}

func (s *Server) boolean(ctx context.Context, flag *flipt.Flag, r *rpcevaluation.EvaluationRequest) (*rpcevaluation.BooleanEvaluationResponse, error) {
	rollouts, err := s.store.GetEvaluationRollouts(ctx, r.NamespaceKey, flag.Key)
	if err != nil {
		return nil, err
	}

	var (
		resp = &rpcevaluation.BooleanEvaluationResponse{
			RequestId: r.RequestId,
		}
		lastRank int32
	)

	var (
		startTime     = time.Now().UTC()
		namespaceAttr = metrics.AttributeNamespace.String(r.NamespaceKey)
		flagAttr      = metrics.AttributeFlag.String(r.FlagKey)
	)

	metrics.EvaluationsTotal.Add(ctx, 1, metric.WithAttributeSet(attribute.NewSet(namespaceAttr, flagAttr)))

	defer func() {
		if err == nil {
			metrics.EvaluationResultsTotal.Add(ctx, 1,
				metric.WithAttributeSet(
					attribute.NewSet(
						namespaceAttr,
						flagAttr,
						metrics.AttributeValue.Bool(resp.Enabled),
						metrics.AttributeReason.String(resp.Reason.String()),
						metrics.AttributeType.String("boolean"),
					),
				),
			)
		} else {
			metrics.EvaluationErrorsTotal.Add(ctx, 1, metric.WithAttributeSet(attribute.NewSet(namespaceAttr, flagAttr)))
		}

		metrics.EvaluationLatency.Record(
			ctx,
			float64(time.Since(startTime).Nanoseconds())/1e6,
			metric.WithAttributeSet(
				attribute.NewSet(
					namespaceAttr,
					flagAttr,
				),
			),
		)
	}()

	for _, rollout := range rollouts {
		if rollout.Rank < lastRank {
			return nil, fmt.Errorf("rollout rank: %d detected out of order", rollout.Rank)
		}

		lastRank = rollout.Rank

		if rollout.Threshold != nil {
			// consistent hashing based on the entity id and flag key.
			hash := crc32.ChecksumIEEE([]byte(r.EntityId + r.FlagKey))

			normalizedValue := float32(int(hash) % 100)

			// if this case does not hold, fall through to the next rollout.
			if normalizedValue < rollout.Threshold.Percentage {
				resp.Enabled = rollout.Threshold.Value
				resp.Reason = rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON
				resp.FlagKey = flag.Key
				s.logger.Debug("threshold based matched", zap.Int("rank", int(rollout.Rank)), zap.String("rollout_type", "threshold"))
				return resp, nil
			}
		} else if rollout.Segment != nil {

			var (
				segmentMatches = 0
				segmentKeys    = []string{}
			)

			for k, v := range rollout.Segment.Segments {
				segmentKeys = append(segmentKeys, k)
				matched, reason, err := matchConstraints(r.Context, v.Constraints, v.MatchType)
				if err != nil {
					return nil, err
				}

				// if we don't match the segment, fall through to the next rollout.
				if matched {
					s.logger.Debug(reason)
					segmentMatches++
				}
			}

			switch rollout.Segment.SegmentOperator {
			case flipt.SegmentOperator_OR_SEGMENT_OPERATOR:
				if segmentMatches < 1 {
					s.logger.Debug("did not match ANY segments")
					continue
				}
			case flipt.SegmentOperator_AND_SEGMENT_OPERATOR:
				if len(rollout.Segment.Segments) != segmentMatches {
					s.logger.Debug("did not match ALL segments")
					continue
				}
			}

			resp.Enabled = rollout.Segment.Value
			resp.Reason = rpcevaluation.EvaluationReason_MATCH_EVALUATION_REASON
			resp.FlagKey = flag.Key

			s.logger.Debug("segment based matched", zap.Int("rank", int(rollout.Rank)), zap.Strings("segments", segmentKeys))
			return resp, nil
		}
	}

	// If we have exhausted all rollouts and we still don't have a match, return flag enabled value.
	resp.Reason = rpcevaluation.EvaluationReason_DEFAULT_EVALUATION_REASON
	resp.Enabled = flag.Enabled
	resp.FlagKey = flag.Key

	s.logger.Debug("default rollout matched", zap.Bool("enabled", flag.Enabled))
	return resp, nil
}

// Batch takes in a list of *evaluation.EvaluationRequest and returns their respective responses.
func (s *Server) Batch(ctx context.Context, b *rpcevaluation.BatchEvaluationRequest) (*rpcevaluation.BatchEvaluationResponse, error) {
	resp := &rpcevaluation.BatchEvaluationResponse{
		Responses: make([]*rpcevaluation.EvaluationResponse, 0, len(b.Requests)),
	}

	for _, req := range b.GetRequests() {
		f, err := s.store.GetFlag(ctx, req.NamespaceKey, req.FlagKey)
		if err != nil {
			var errnf errs.ErrNotFound
			if errors.As(err, &errnf) {
				eresp := &rpcevaluation.EvaluationResponse{
					Type: rpcevaluation.EvaluationResponseType_ERROR_EVALUATION_RESPONSE_TYPE,
					Response: &rpcevaluation.EvaluationResponse_ErrorResponse{
						ErrorResponse: &rpcevaluation.ErrorEvaluationResponse{
							FlagKey:      req.FlagKey,
							NamespaceKey: req.NamespaceKey,
							Reason:       rpcevaluation.ErrorEvaluationReason_NOT_FOUND_ERROR_EVALUATION_REASON,
						},
					},
				}

				resp.Responses = append(resp.Responses, eresp)
				continue
			}

			return nil, err
		}

		switch f.Type {
		case flipt.FlagType_BOOLEAN_FLAG_TYPE:
			res, err := s.boolean(ctx, f, req)
			if err != nil {
				return nil, err
			}

			eresp := &rpcevaluation.EvaluationResponse{
				Type: rpcevaluation.EvaluationResponseType_BOOLEAN_EVALUATION_RESPONSE_TYPE,
				Response: &rpcevaluation.EvaluationResponse_BooleanResponse{
					BooleanResponse: res,
				},
			}

			resp.Responses = append(resp.Responses, eresp)
		case flipt.FlagType_VARIANT_FLAG_TYPE:
			res, err := s.variant(ctx, f, req)
			if err != nil {
				return nil, err
			}
			eresp := &rpcevaluation.EvaluationResponse{
				Type: rpcevaluation.EvaluationResponseType_VARIANT_EVALUATION_RESPONSE_TYPE,
				Response: &rpcevaluation.EvaluationResponse_VariantResponse{
					VariantResponse: res,
				},
			}

			resp.Responses = append(resp.Responses, eresp)
		default:
			return nil, errs.ErrInvalidf("unknown flag type: %s", f.Type)
		}
	}

	return resp, nil
}
