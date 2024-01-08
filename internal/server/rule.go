package server

import (
	"context"

	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

// GetRule gets a rule
func (s *Server) GetRule(ctx context.Context, r *flipt.GetRuleRequest) (*flipt.Rule, error) {
	s.logger.Debug("get rule", zap.Stringer("request", r))
	rule, err := s.store.GetRule(ctx,
		storage.NewNamespace(r.NamespaceKey, storage.WithReference(r.Reference)),
		r.Id)
	s.logger.Debug("get rule", zap.Stringer("response", rule))
	return rule, err
}

// ListRules lists all rules for a flag
func (s *Server) ListRules(ctx context.Context, r *flipt.ListRuleRequest) (*flipt.RuleList, error) {
	s.logger.Debug("list rules", zap.Stringer("request", r))

	flag := storage.NewResource(r.NamespaceKey, r.FlagKey, storage.WithReference(r.Reference))
	results, err := s.store.ListRules(ctx, storage.ListWithParameters(flag, r))
	if err != nil {
		return nil, err
	}

	resp := flipt.RuleList{
		Rules: results.Results,
	}

	total, err := s.store.CountRules(ctx, flag)
	if err != nil {
		return nil, err
	}

	resp.TotalCount = int32(total)
	resp.NextPageToken = results.NextPageToken

	s.logger.Debug("list rules", zap.Stringer("response", &resp))
	return &resp, nil
}

// CreateRule creates a rule
func (s *Server) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	s.logger.Debug("create rule", zap.Stringer("request", r))
	rule, err := s.store.CreateRule(ctx, r)
	s.logger.Debug("create rule", zap.Stringer("response", rule))
	return rule, err
}

// UpdateRule updates an existing rule
func (s *Server) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	s.logger.Debug("update rule", zap.Stringer("request", r))
	rule, err := s.store.UpdateRule(ctx, r)
	s.logger.Debug("update rule", zap.Stringer("response", rule))
	return rule, err
}

// DeleteRule deletes a rule
func (s *Server) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) (*empty.Empty, error) {
	s.logger.Debug("delete rule", zap.Stringer("request", r))
	if err := s.store.DeleteRule(ctx, r); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

// OrderRules orders rules
func (s *Server) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) (*empty.Empty, error) {
	s.logger.Debug("order rules", zap.Stringer("request", r))
	if err := s.store.OrderRules(ctx, r); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

// CreateDistribution creates a distribution
func (s *Server) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	s.logger.Debug("create distribution", zap.Stringer("request", r))
	distribution, err := s.store.CreateDistribution(ctx, r)
	s.logger.Debug("create distribution", zap.Stringer("response", distribution))
	return distribution, err
}

// UpdateDistribution updates an existing distribution
func (s *Server) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	s.logger.Debug("update distribution", zap.Stringer("request", r))
	distribution, err := s.store.UpdateDistribution(ctx, r)
	s.logger.Debug("update distribution", zap.Stringer("response", distribution))
	return distribution, err
}

// DeleteDistribution deletes a distribution
func (s *Server) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) (*empty.Empty, error) {
	s.logger.Debug("delete distribution", zap.Stringer("request", r))
	if err := s.store.DeleteDistribution(ctx, r); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}
