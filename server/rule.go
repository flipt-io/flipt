package server

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	"github.com/markphelps/flipt"
	uuid "github.com/satori/go.uuid"
)

func (s *Server) GetRule(ctx context.Context, req *flipt.GetRuleRequest) (*flipt.Rule, error) {
	return s.RuleService.Rule(ctx, req)
}

func (s *Server) ListRules(ctx context.Context, req *flipt.ListRuleRequest) (*flipt.RuleList, error) {
	if req.FlagKey == "" {
		return nil, flipt.EmptyFieldError("flagKey")
	}

	rules, err := s.RuleService.Rules(ctx, req)
	if err != nil {
		return nil, err
	}

	var resp flipt.RuleList

	for i := range rules {
		resp.Rules = append(resp.Rules, rules[i])
	}

	return &resp, nil
}

func (s *Server) CreateRule(ctx context.Context, req *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	if req.FlagKey == "" {
		return nil, flipt.EmptyFieldError("flagKey")
	}
	if req.SegmentKey == "" {
		return nil, flipt.EmptyFieldError("segmentKey")
	}
	if req.Rank <= 0 {
		return nil, flipt.InvalidFieldError("rank", "must be greater than 0")
	}
	return s.RuleService.CreateRule(ctx, req)
}

func (s *Server) UpdateRule(ctx context.Context, req *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	if req.Id == "" {
		return nil, flipt.EmptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, flipt.EmptyFieldError("flagKey")
	}
	if req.SegmentKey == "" {
		return nil, flipt.EmptyFieldError("segmentKey")
	}
	return s.RuleService.UpdateRule(ctx, req)
}

func (s *Server) DeleteRule(ctx context.Context, req *flipt.DeleteRuleRequest) (*empty.Empty, error) {
	if req.Id == "" {
		return nil, flipt.EmptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, flipt.EmptyFieldError("flagKey")
	}

	if err := s.RuleService.DeleteRule(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) OrderRules(ctx context.Context, req *flipt.OrderRulesRequest) (*empty.Empty, error) {
	if req.FlagKey == "" {
		return nil, flipt.EmptyFieldError("flagKey")
	}
	if len(req.RuleIds) < 2 {
		return nil, flipt.InvalidFieldError("ruleIds", "must contain atleast 2 elements")
	}

	if err := s.RuleService.OrderRules(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) CreateDistribution(ctx context.Context, req *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	if req.FlagKey == "" {
		return nil, flipt.EmptyFieldError("flagKey")
	}
	if req.RuleId == "" {
		return nil, flipt.EmptyFieldError("ruleId")
	}
	if req.VariantId == "" {
		return nil, flipt.EmptyFieldError("variantId")
	}
	if req.Rollout < 0 {
		return nil, flipt.InvalidFieldError("rollout", "must be greater than or equal to '0'")
	}
	if req.Rollout > 100 {
		return nil, flipt.InvalidFieldError("rollout", "must be less than or equal to '100'")
	}
	return s.RuleService.CreateDistribution(ctx, req)
}

func (s *Server) UpdateDistribution(ctx context.Context, req *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	if req.Id == "" {
		return nil, flipt.EmptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, flipt.EmptyFieldError("flagKey")
	}
	if req.RuleId == "" {
		return nil, flipt.EmptyFieldError("ruleId")
	}
	if req.VariantId == "" {
		return nil, flipt.EmptyFieldError("variantId")
	}
	if req.Rollout < 0 {
		return nil, flipt.InvalidFieldError("rollout", "must be greater than or equal to '0'")
	}
	if req.Rollout > 100 {
		return nil, flipt.InvalidFieldError("rollout", "must be less than or equal to '100'")
	}
	return s.RuleService.UpdateDistribution(ctx, req)
}

func (s *Server) DeleteDistribution(ctx context.Context, req *flipt.DeleteDistributionRequest) (*empty.Empty, error) {
	if req.Id == "" {
		return nil, flipt.EmptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, flipt.EmptyFieldError("flagKey")
	}
	if req.RuleId == "" {
		return nil, flipt.EmptyFieldError("ruleId")
	}
	if req.VariantId == "" {
		return nil, flipt.EmptyFieldError("variantId")
	}

	if err := s.RuleService.DeleteDistribution(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (s *Server) Evaluate(ctx context.Context, req *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
	if req.FlagKey == "" {
		return nil, flipt.EmptyFieldError("flagKey")
	}
	if req.EntityId == "" {
		return nil, flipt.EmptyFieldError("entityId")
	}

	startTime := time.Now()

	// set request ID if not present
	if req.RequestId == "" {
		req.RequestId = uuid.NewV4().String()
	}

	resp, err := s.RuleService.Evaluate(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp != nil {
		resp.RequestDurationMillis = float64(time.Since(startTime)) / float64(time.Millisecond)
	}

	return resp, nil
}
