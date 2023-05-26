package server

import (
	"context"

	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) CreateRolloutRule(ctx context.Context, r *flipt.CreateRolloutRuleRequest) (*flipt.RolloutRule, error) {
	s.logger.Debug("create rollout rule", zap.Stringer("request", r))
	rollout, err := s.store.CreateRolloutRule(ctx, r)
	s.logger.Debug("create rollout rule", zap.Stringer("response", rollout))
	return rollout, err
}

func (s *Server) UpdateRolloutRule(ctx context.Context, r *flipt.UpdateRolloutRuleRequest) (*flipt.RolloutRule, error) {
	s.logger.Debug("update rollout rule", zap.Stringer("request", r))
	rollout, err := s.store.UpdateRolloutRule(ctx, r)
	s.logger.Debug("update rollout rule", zap.Stringer("response", rollout))
	return rollout, err
}

func (s *Server) DeleteRolloutRule(ctx context.Context, r *flipt.DeleteRolloutRuleRequest) (*empty.Empty, error) {
	s.logger.Debug("delete rollout rule", zap.Stringer("request", r))
	err := s.store.DeleteRolloutRule(ctx, r)
	s.logger.Debug("delete rollout rule", zap.Error(err))
	return &empty.Empty{}, err
}
