package server

import (
	"context"

	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) GetRolloutStrategy(ctx context.Context, r *flipt.GetRolloutStrategyRequest) (*flipt.RolloutStrategy, error) {
	s.logger.Debug("get rollout strategy", zap.Stringer("request", r))
	rollout, err := s.store.GetRolloutStrategy(ctx, r.NamespaceKey, r.FlagKey)
	s.logger.Debug("get rollout strategy", zap.Stringer("response", rollout))
	return rollout, err
}

func (s *Server) CreateRolloutStrategy(ctx context.Context, r *flipt.CreateRolloutStrategyRequest) (*flipt.RolloutStrategy, error) {
	s.logger.Debug("create rollout strategy", zap.Stringer("request", r))
	rollout, err := s.store.CreateRolloutStrategy(ctx, r)
	s.logger.Debug("create rollout strategy", zap.Stringer("response", rollout))
	return rollout, err
}

func (s *Server) UpdateRolloutStrategy(ctx context.Context, r *flipt.UpdateRolloutStrategyRequest) (*flipt.RolloutStrategy, error) {
	s.logger.Debug("update rollout strategy", zap.Stringer("request", r))
	rollout, err := s.store.UpdateRolloutStrategy(ctx, r)
	s.logger.Debug("update rollout strategy", zap.Stringer("response", rollout))
	return rollout, err
}

func (s *Server) DeleteRolloutStrategy(ctx context.Context, r *flipt.DeleteRolloutStrategyRequest) (*empty.Empty, error) {
	s.logger.Debug("delete rollout strategy", zap.Stringer("request", r))
	err := s.store.DeleteRolloutStrategy(ctx, r)
	s.logger.Debug("delete rollout strategy", zap.Error(err))
	return &empty.Empty{}, err
}
