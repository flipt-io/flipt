package server

import (
	"context"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	flipt "github.com/markphelps/flipt/proto"
	uuid "github.com/satori/go.uuid"
)

// GetRule gets a rule
func (s *Server) GetRule(ctx context.Context, req *flipt.GetRuleRequest) (*flipt.Rule, error) {
	return s.RuleStore.GetRule(ctx, req)
}

// ListRules lists all rules
func (s *Server) ListRules(ctx context.Context, req *flipt.ListRuleRequest) (*flipt.RuleList, error) {
	if req.FlagKey == "" {
		return nil, EmptyFieldError("flagKey")
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
		return nil, EmptyFieldError("flagKey")
	}
	if req.SegmentKey == "" {
		return nil, EmptyFieldError("segmentKey")
	}
	if req.Rank <= 0 {
		return nil, InvalidFieldError("rank", "must be greater than 0")
	}
	return s.RuleStore.CreateRule(ctx, req)
}

// UpdateRule updates an existing rule
func (s *Server) UpdateRule(ctx context.Context, req *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	if req.Id == "" {
		return nil, EmptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, EmptyFieldError("flagKey")
	}
	if req.SegmentKey == "" {
		return nil, EmptyFieldError("segmentKey")
	}
	return s.RuleStore.UpdateRule(ctx, req)
}

// DeleteRule deletes a rule
func (s *Server) DeleteRule(ctx context.Context, req *flipt.DeleteRuleRequest) (*empty.Empty, error) {
	if req.Id == "" {
		return nil, EmptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, EmptyFieldError("flagKey")
	}

	if err := s.RuleStore.DeleteRule(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// OrderRules orders rules
func (s *Server) OrderRules(ctx context.Context, req *flipt.OrderRulesRequest) (*empty.Empty, error) {
	if req.FlagKey == "" {
		return nil, EmptyFieldError("flagKey")
	}
	if len(req.RuleIds) < 2 {
		return nil, InvalidFieldError("ruleIds", "must contain atleast 2 elements")
	}

	if err := s.RuleStore.OrderRules(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// CreateDistribution creates a distribution
func (s *Server) CreateDistribution(ctx context.Context, req *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	if req.FlagKey == "" {
		return nil, EmptyFieldError("flagKey")
	}
	if req.RuleId == "" {
		return nil, EmptyFieldError("ruleId")
	}
	if req.VariantId == "" {
		return nil, EmptyFieldError("variantId")
	}
	if req.Rollout < 0 {
		return nil, InvalidFieldError("rollout", "must be greater than or equal to '0'")
	}
	if req.Rollout > 100 {
		return nil, InvalidFieldError("rollout", "must be less than or equal to '100'")
	}
	return s.RuleStore.CreateDistribution(ctx, req)
}

// UpdateDistribution updates an existing distribution
func (s *Server) UpdateDistribution(ctx context.Context, req *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	if req.Id == "" {
		return nil, EmptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, EmptyFieldError("flagKey")
	}
	if req.RuleId == "" {
		return nil, EmptyFieldError("ruleId")
	}
	if req.VariantId == "" {
		return nil, EmptyFieldError("variantId")
	}
	if req.Rollout < 0 {
		return nil, InvalidFieldError("rollout", "must be greater than or equal to '0'")
	}
	if req.Rollout > 100 {
		return nil, InvalidFieldError("rollout", "must be less than or equal to '100'")
	}
	return s.RuleStore.UpdateDistribution(ctx, req)
}

// DeleteDistribution deletes a distribution
func (s *Server) DeleteDistribution(ctx context.Context, req *flipt.DeleteDistributionRequest) (*empty.Empty, error) {
	if req.Id == "" {
		return nil, EmptyFieldError("id")
	}
	if req.FlagKey == "" {
		return nil, EmptyFieldError("flagKey")
	}
	if req.RuleId == "" {
		return nil, EmptyFieldError("ruleId")
	}
	if req.VariantId == "" {
		return nil, EmptyFieldError("variantId")
	}

	if err := s.RuleStore.DeleteDistribution(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// Evaluate evaluates a request for a given flag and entity
func (s *Server) Evaluate(ctx context.Context, req *flipt.EvaluationRequest) (*flipt.EvaluationResponse, error) {
	if req.FlagKey == "" {
		return nil, EmptyFieldError("flagKey")
	}
	if req.EntityId == "" {
		return nil, EmptyFieldError("entityId")
	}

	startTime := time.Now()

	// set request ID if not present
	if req.RequestId == "" {
		req.RequestId = uuid.NewV4().String()
	}

	resp, err := s.RuleStore.Evaluate(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp != nil {
		resp.RequestDurationMillis = float64(time.Since(startTime)) / float64(time.Millisecond)
	}

	return resp, nil
}
