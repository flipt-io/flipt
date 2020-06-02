package server

import (
	"context"

	"github.com/golang/protobuf/ptypes/empty"
	flipt "github.com/markphelps/flipt/rpc"
	"github.com/markphelps/flipt/storage"
)

// GetRule gets a rule
func (s *Server) GetRule(ctx context.Context, r *flipt.GetRuleRequest) (*flipt.Rule, error) {
	s.logger.WithField("request", r).Debug("get rule")
	rule, err := s.RuleStore.GetRule(ctx, r.Id)
	s.logger.WithField("response", rule).Debug("get rule")
	return rule, err
}

// ListRules lists all rules for a flag
func (s *Server) ListRules(ctx context.Context, r *flipt.ListRuleRequest) (*flipt.RuleList, error) {
	s.logger.WithField("request", r).Debug("list rules")
	rules, err := s.RuleStore.ListRules(ctx, r.FlagKey, storage.WithLimit(uint64(r.Limit)), storage.WithOffset(uint64(r.Offset)))
	if err != nil {
		return nil, err
	}

	var resp flipt.RuleList

	for i := range rules {
		resp.Rules = append(resp.Rules, rules[i])
	}

	s.logger.WithField("response", &resp).Debug("list rules")
	return &resp, nil
}

// CreateRule creates a rule
func (s *Server) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	s.logger.WithField("request", r).Debug("create rule")
	rule, err := s.RuleStore.CreateRule(ctx, r)
	s.logger.WithField("response", rule).Debug("create rule")
	return rule, err
}

// UpdateRule updates an existing rule
func (s *Server) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	s.logger.WithField("request", r).Debug("update rule")
	rule, err := s.RuleStore.UpdateRule(ctx, r)
	s.logger.WithField("response", rule).Debug("update rule")
	return rule, err
}

// DeleteRule deletes a rule
func (s *Server) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) (*empty.Empty, error) {
	s.logger.WithField("request", r).Debug("delete rule")
	if err := s.RuleStore.DeleteRule(ctx, r); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

// OrderRules orders rules
func (s *Server) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) (*empty.Empty, error) {
	s.logger.WithField("request", r).Debug("order rules")
	if err := s.RuleStore.OrderRules(ctx, r); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

// CreateDistribution creates a distribution
func (s *Server) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	s.logger.WithField("request", r).Debug("create distribution")
	distribution, err := s.RuleStore.CreateDistribution(ctx, r)
	s.logger.WithField("response", distribution).Debug("create distribution")
	return distribution, err
}

// UpdateDistribution updates an existing distribution
func (s *Server) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	s.logger.WithField("request", r).Debug("update distribution")
	distribution, err := s.RuleStore.UpdateDistribution(ctx, r)
	s.logger.WithField("response", distribution).Debug("update distribution")
	return distribution, err
}

// DeleteDistribution deletes a distribution
func (s *Server) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) (*empty.Empty, error) {
	s.logger.WithField("request", r).Debug("delete distribution")
	if err := s.RuleStore.DeleteDistribution(ctx, r); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}
