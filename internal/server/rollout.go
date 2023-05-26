package server

import (
	"context"
	"encoding/base64"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	empty "google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) GetRolloutRule(ctx context.Context, r *flipt.GetRolloutRuleRequest) (*flipt.RolloutRule, error) {
	s.logger.Debug("get rollout rule", zap.Stringer("request", r))
	rollout, err := s.store.GetRolloutRule(ctx, r.NamespaceKey, r.Id)
	s.logger.Debug("get rollout rule", zap.Stringer("response", rollout))
	return rollout, err
}

func (s *Server) ListRolloutRules(ctx context.Context, r *flipt.ListRolloutRuleRequest) (*flipt.RolloutRuleList, error) {
	s.logger.Debug("list rollout rules", zap.Stringer("request", r))

	opts := []storage.QueryOption{storage.WithLimit(uint64(r.GetLimit()))}

	if r.GetPageToken() != "" {
		tok, err := base64.StdEncoding.DecodeString(r.GetPageToken())
		if err != nil {
			return nil, errors.ErrInvalidf("pageToken is not valid: %q", r.GetPageToken())
		}

		opts = append(opts, storage.WithPageToken(string(tok)))
	}

	results, err := s.store.ListRolloutRules(ctx, r.NamespaceKey, r.FlagKey, opts...)
	if err != nil {
		return nil, err
	}

	resp := flipt.RolloutRuleList{
		Rules: results.Results,
	}

	total, err := s.store.CountRules(ctx, r.NamespaceKey)
	if err != nil {
		return nil, err
	}

	resp.TotalCount = int32(total)
	resp.NextPageToken = base64.StdEncoding.EncodeToString([]byte(results.NextPageToken))

	s.logger.Debug("list rollout rules", zap.Stringer("response", &resp))
	return &resp, nil
}

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
