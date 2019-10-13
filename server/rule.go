package server

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	flipt "github.com/markphelps/flipt/rpc"
)

// GetRule gets a rule
func (s *Server) GetRule(ctx context.Context, req *flipt.GetRuleRequest) (*flipt.Rule, error) {
	return s.RuleStore.GetRule(ctx, req)
}

// ListRules lists all rules
func (s *Server) ListRules(ctx context.Context, req *flipt.ListRuleRequest) (*flipt.RuleList, error) {
	if req.FlagKey == "" {
		return nil, emptyFieldError("flagKey")
	}

	rules, err := s.RuleStore.ListRules(ctx, req)
	if err != nil {
		return nil, err
	}

	var resp flipt.RuleList

	for i := range rules {
		resp.Rules = append(resp.Rules, rules[i])
	}

	return &resp, nil
}

// CreateRule creates a rule
func (s *Server) CreateRule(ctx context.Context, req *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	if req.FlagKey == "" {
		return nil, emptyFieldError("flagKey")
	}
	if req.SegmentKey == "" {
		return nil, emptyFieldError("segmentKey")
	}
	if req.Rank <= 0 {
		return nil, invalidFieldError("rank", "must be greater than 0")
	}
	return s.RuleStore.CreateRule(ctx, req)
}

// UpdateRule updates an existing rule
func (s *Server) UpdateRule(ctx context.Context, req *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	if req.Id == "" {
		return nil, emptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, emptyFieldError("flagKey")
	}
	if req.SegmentKey == "" {
		return nil, emptyFieldError("segmentKey")
	}
	return s.RuleStore.UpdateRule(ctx, req)
}

// DeleteRule deletes a rule
func (s *Server) DeleteRule(ctx context.Context, req *flipt.DeleteRuleRequest) (*empty.Empty, error) {
	if req.Id == "" {
		return nil, emptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, emptyFieldError("flagKey")
	}

	if err := s.RuleStore.DeleteRule(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// OrderRules orders rules
func (s *Server) OrderRules(ctx context.Context, req *flipt.OrderRulesRequest) (*empty.Empty, error) {
	if req.FlagKey == "" {
		return nil, emptyFieldError("flagKey")
	}
	if len(req.RuleIds) < 2 {
		return nil, invalidFieldError("ruleIds", "must contain atleast 2 elements")
	}

	if err := s.RuleStore.OrderRules(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// CreateDistribution creates a distribution
func (s *Server) CreateDistribution(ctx context.Context, req *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	if req.FlagKey == "" {
		return nil, emptyFieldError("flagKey")
	}
	if req.RuleId == "" {
		return nil, emptyFieldError("ruleId")
	}
	if req.VariantId == "" {
		return nil, emptyFieldError("variantId")
	}
	if req.Rollout < 0 {
		return nil, invalidFieldError("rollout", "must be greater than or equal to '0'")
	}
	if req.Rollout > 100 {
		return nil, invalidFieldError("rollout", "must be less than or equal to '100'")
	}
	return s.RuleStore.CreateDistribution(ctx, req)
}

// UpdateDistribution updates an existing distribution
func (s *Server) UpdateDistribution(ctx context.Context, req *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	if req.Id == "" {
		return nil, emptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, emptyFieldError("flagKey")
	}
	if req.RuleId == "" {
		return nil, emptyFieldError("ruleId")
	}
	if req.VariantId == "" {
		return nil, emptyFieldError("variantId")
	}
	if req.Rollout < 0 {
		return nil, invalidFieldError("rollout", "must be greater than or equal to '0'")
	}
	if req.Rollout > 100 {
		return nil, invalidFieldError("rollout", "must be less than or equal to '100'")
	}
	return s.RuleStore.UpdateDistribution(ctx, req)
}

// DeleteDistribution deletes a distribution
func (s *Server) DeleteDistribution(ctx context.Context, req *flipt.DeleteDistributionRequest) (*empty.Empty, error) {
	if req.Id == "" {
		return nil, emptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, emptyFieldError("flagKey")
	}
	if req.RuleId == "" {
		return nil, emptyFieldError("ruleId")
	}
	if req.VariantId == "" {
		return nil, emptyFieldError("variantId")
	}

	if err := s.RuleStore.DeleteDistribution(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}
