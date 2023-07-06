package server

import (
	"context"

	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) GetRollout(ctx context.Context, r *flipt.GetRolloutRequest) (*flipt.Rollout, error) {
	s.logger.Debug("get rollout rule", zap.Stringer("request", r))
	rollout, err := s.store.GetRollout(ctx, r.NamespaceKey, r.Id)
	s.logger.Debug("get rollout rule", zap.Stringer("response", rollout))
	return rollout, err
}

func (s *Server) ListRollouts(ctx context.Context, r *flipt.ListRolloutRequest) (*flipt.RolloutList, error) {
	s.logger.Debug("list rollout rules", zap.Stringer("request", r))

	opts := []storage.QueryOption{storage.WithLimit(uint64(r.GetLimit()))}

	results, err := s.store.ListRollouts(ctx, r.NamespaceKey, r.FlagKey, opts...)
	if err != nil {
		return nil, err
	}

	resp := flipt.RolloutList{
		Rules: results.Results,
	}

	total, err := s.store.CountRollouts(ctx, r.NamespaceKey, r.FlagKey)
	if err != nil {
		return nil, err
	}

	resp.TotalCount = int32(total)
	resp.NextPageToken = results.NextPageToken

	s.logger.Debug("list rollout rules", zap.Stringer("response", &resp))
	return &resp, nil
}

func (s *Server) CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (*flipt.Rollout, error) {
	s.logger.Debug("create rollout rule", zap.Stringer("request", r))
	rollout, err := s.store.CreateRollout(ctx, r)
	s.logger.Debug("create rollout rule", zap.Stringer("response", rollout))
	return rollout, err
}

func (s *Server) UpdateRollout(ctx context.Context, r *flipt.UpdateRolloutRequest) (*flipt.Rollout, error) {
	s.logger.Debug("update rollout rule", zap.Stringer("request", r))
	rollout, err := s.store.UpdateRollout(ctx, r)
	s.logger.Debug("update rollout rule", zap.Stringer("response", rollout))
	return rollout, err
}

func (s *Server) DeleteRollout(ctx context.Context, r *flipt.DeleteRolloutRequest) (*empty.Empty, error) {
	s.logger.Debug("delete rollout rule", zap.Stringer("request", r))
	err := s.store.DeleteRollout(ctx, r)
	s.logger.Debug("delete rollout rule", zap.Error(err))
	return &empty.Empty{}, err
}
