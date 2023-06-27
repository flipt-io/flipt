package server

import (
	"context"
	"errors"

	errs "go.flipt.io/flipt/errors"
	fliptotel "go.flipt.io/flipt/internal/server/otel"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// Evaluate evaluates a request for a given flag and entity
func (s *Server) Evaluate(ctx context.Context, r *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
	s.logger.Debug("evaluate", zap.Stringer("request", r))

	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	resp, err := s.mvEvaluator.Evaluate(ctx, r)
	if err != nil {
		return resp, err
	}

	spanAttrs := []attribute.KeyValue{
		fliptotel.AttributeNamespace.String(r.NamespaceKey),
		fliptotel.AttributeFlag.String(r.FlagKey),
		fliptotel.AttributeEntityID.String(r.EntityId),
		fliptotel.AttributeRequestID.String(r.RequestId),
	}

	if resp != nil {
		spanAttrs = append(spanAttrs,
			fliptotel.AttributeMatch.Bool(resp.Match),
			fliptotel.AttributeSegment.String(resp.SegmentKey),
			fliptotel.AttributeValue.String(resp.Value),
			fliptotel.AttributeReason.String(resp.Reason.String()),
		)
	}

	// add otel attributes to span
	span := trace.SpanFromContext(ctx)
	span.SetAttributes(spanAttrs...)

	s.logger.Debug("evaluate", zap.Stringer("response", resp))
	return resp, nil
}

// BatchEvaluate evaluates a request for multiple flags and entities
func (s *Server) BatchEvaluate(ctx context.Context, r *flipt.BatchEvaluationRequest) (*flipt.BatchEvaluationResponse, error) {
	s.logger.Debug("batch-evaluate", zap.Stringer("request", r))

	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	resp, err := s.batchEvaluate(ctx, r)
	if err != nil {
		return nil, err
	}
	s.logger.Debug("batch-evaluate", zap.Stringer("response", resp))
	return resp, nil
}

func (s *Server) batchEvaluate(ctx context.Context, r *flipt.BatchEvaluationRequest) (*flipt.BatchEvaluationResponse, error) {
	res := flipt.BatchEvaluationResponse{
		Responses: make([]*flipt.EvaluationResponse, 0, len(r.GetRequests())),
	}

	// TODO: we should change this to a native batch query instead of looping through
	// each request individually
	for _, req := range r.GetRequests() {
		// ensure all requests have the same namespace
		if req.NamespaceKey == "" {
			req.NamespaceKey = r.NamespaceKey
		} else if req.NamespaceKey != r.NamespaceKey {
			return &res, errs.InvalidFieldError("namespace_key", "must be the same for all requests if specified")
		}

		// TODO: we also need to validate each request, we should likely do this in the validation middleware
		f, err := s.mvEvaluator.Evaluate(ctx, req)
		if err != nil {
			var errnf errs.ErrNotFound
			if r.GetExcludeNotFound() && errors.As(err, &errnf) {
				continue
			}
			return &res, err
		}
		f.RequestId = ""
		res.Responses = append(res.Responses, f)
	}

	return &res, nil
}
