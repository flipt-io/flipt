package evaluation

import (
	"context"
	"errors"

	errs "go.flipt.io/flipt/errors"
	fliptotel "go.flipt.io/flipt/internal/server/otel"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	rpcevaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Variant evaluates a request for a multi-variate flag and entity.
func (s *Server) Variant(ctx context.Context, v *rpcevaluation.EvaluationRequest) (*rpcevaluation.VariantEvaluationResponse, error) {
	flag, err := s.store.GetFlag(ctx, v.NamespaceKey, v.FlagKey)
	if err != nil {
		return nil, err
	}

	s.logger.Debug("variant", zap.Stringer("request", v))

	if v.NamespaceKey == "" {
		v.NamespaceKey = storage.DefaultNamespace
	}

	ver, err := s.variant(ctx, flag, v)
	if err != nil {
		return nil, err
	}

	spanAttrs := []attribute.KeyValue{
		fliptotel.AttributeNamespace.String(v.NamespaceKey),
		fliptotel.AttributeFlag.String(v.FlagKey),
		fliptotel.AttributeEntityID.String(v.EntityId),
		fliptotel.AttributeRequestID.String(v.RequestId),
	}

	if ver != nil {
		spanAttrs = append(spanAttrs,
			fliptotel.AttributeMatch.Bool(ver.Match),
			fliptotel.AttributeValue.String(ver.VariantKey),
			fliptotel.AttributeReason.String(ver.Reason.String()),
			fliptotel.AttributeSegment.String(ver.SegmentKey),
		)
	}

	// add otel attributes to span
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(spanAttrs...)

	s.logger.Debug("variant", zap.Stringer("response", ver))
	return ver, nil
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

	ver := &rpcevaluation.VariantEvaluationResponse{
		Match:             resp.Match,
		SegmentKey:        resp.SegmentKey,
		Reason:            rpcevaluation.EvaluationReason(resp.Reason),
		VariantKey:        resp.Value,
		VariantAttachment: resp.Attachment,
	}

	ver.Match = resp.Match
	ver.SegmentKey = resp.SegmentKey
	ver.Reason = rpcevaluation.EvaluationReason(resp.Reason)
	ver.VariantKey = resp.Value
	ver.VariantAttachment = resp.Attachment

	return ver, nil
}

// Boolean evaluates a request for a boolean flag and entity.
func (s *Server) Boolean(ctx context.Context, r *rpcevaluation.EvaluationRequest) (*rpcevaluation.BooleanEvaluationResponse, error) {
	flag, err := s.store.GetFlag(ctx, r.NamespaceKey, r.FlagKey)
	if err != nil {
		return nil, err
	}

	if flag.Type != flipt.FlagType_BOOLEAN_FLAG_TYPE {
		return nil, errs.ErrInvalidf("flag type %s invalid", flag.Type)
	}

	return s.boolean(ctx, flag, r)
}

func (s *Server) boolean(ctx context.Context, flag *flipt.Flag, r *rpcevaluation.EvaluationRequest) (*rpcevaluation.BooleanEvaluationResponse, error) {
	rollouts, err := s.store.GetEvaluationRollouts(ctx, r.NamespaceKey, flag.Key)
	if err != nil {
		return nil, err
	}

	resp, err := s.evaluator.booleanMatch(r, flag.Enabled, rollouts)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// Batch takes ina list of *evaluation.EvaluationRequest and returns their respective responses.
func (s *Server) Batch(ctx context.Context, b *rpcevaluation.BatchEvaluationRequest) (*rpcevaluation.BatchEvaluationResponse, error) {
	resp := &rpcevaluation.BatchEvaluationResponse{
		Responses: []*rpcevaluation.EvaluationResponse{},
	}

	for _, req := range b.GetRequests() {
		f, err := s.store.GetFlag(ctx, req.NamespaceKey, req.FlagKey)
		if err != nil {
			var errnf errs.ErrNotFound
			if b.GetExcludeNotFound() && errors.As(err, &errnf) {
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
				FlagType: rpcevaluation.EvaluationFlagType(flipt.FlagType_BOOLEAN_FLAG_TYPE),
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
				FlagType: rpcevaluation.EvaluationFlagType(flipt.FlagType_VARIANT_FLAG_TYPE),
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
