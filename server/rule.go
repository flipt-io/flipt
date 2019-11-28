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
	return s.RuleStore.CreateRule(ctx, req)
}

// UpdateRule updates an existing rule
func (s *Server) UpdateRule(ctx context.Context, req *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	return s.RuleStore.UpdateRule(ctx, req)
}

// DeleteRule deletes a rule
func (s *Server) DeleteRule(ctx context.Context, req *flipt.DeleteRuleRequest) (*empty.Empty, error) {
	if err := s.RuleStore.DeleteRule(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// OrderRules orders rules
func (s *Server) OrderRules(ctx context.Context, req *flipt.OrderRulesRequest) (*empty.Empty, error) {
	if err := s.RuleStore.OrderRules(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

// CreateDistribution creates a distribution
func (s *Server) CreateDistribution(ctx context.Context, req *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	return s.RuleStore.CreateDistribution(ctx, req)
}

// UpdateDistribution updates an existing distribution
func (s *Server) UpdateDistribution(ctx context.Context, req *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	return s.RuleStore.UpdateDistribution(ctx, req)
}

// DeleteDistribution deletes a distribution
func (s *Server) DeleteDistribution(ctx context.Context, req *flipt.DeleteDistributionRequest) (*empty.Empty, error) {
	if err := s.RuleStore.DeleteDistribution(ctx, req); err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}
