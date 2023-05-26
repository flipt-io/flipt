package common

import (
	"context"

	"go.flipt.io/flipt/rpc/flipt"
)

func (s *Store) GetRolloutStrategy(ctx context.Context, namespaceKey, flagKey string) (*flipt.RolloutStrategy, error) {
	panic("not implemented")
}

func (s *Store) CreateRolloutStrategy(ctx context.Context, r *flipt.CreateRolloutStrategyRequest) (*flipt.RolloutStrategy, error) {
	panic("not implemented")
}

func (s *Store) UpdateRolloutStrategy(ctx context.Context, r *flipt.UpdateRolloutStrategyRequest) (*flipt.RolloutStrategy, error) {
	panic("not implemented")
}

func (s *Store) DeleteRolloutStrategy(ctx context.Context, r *flipt.DeleteRolloutStrategyRequest) error {
	panic("not implemented")
}
