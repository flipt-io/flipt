package evaluation

import (
	"context"

	"go.flipt.io/flipt/rpc/flipt"
	rpcEvaluation "go.flipt.io/flipt/rpc/flipt/evaluation"
)

// Variant evaluates a request for a multi-variate flag and entity.
func (e *EvaluateServer) Variant(ctx context.Context, v *rpcEvaluation.EvaluationRequest) (*rpcEvaluation.VariantEvaluationResponse, error) {
	resp, err := e.mvEvaluator.Evaluate(ctx, &flipt.EvaluationRequest{
		RequestId:    v.RequestId,
		FlagKey:      v.FlagKey,
		EntityId:     v.EntityId,
		Context:      v.Context,
		NamespaceKey: v.NamespaceKey,
	})
	if err != nil {
		return nil, err
	}

	ver := &rpcEvaluation.VariantEvaluationResponse{
		Match:             resp.Match,
		SegmentKey:        resp.SegmentKey,
		Reason:            rpcEvaluation.EvaluationReason(resp.Reason),
		VariantKey:        resp.Value,
		VariantAttachment: resp.Attachment,
	}

	return ver, nil
}
