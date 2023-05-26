package common

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"go.flipt.io/flipt/internal/storage"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	"go.flipt.io/flipt/rpc/flipt"
)

type RolloutStrategyType uint8

const (
	UnknownRolloutStrategyType RolloutStrategyType = iota
	SegmentRolloutStrategyType
	PercentageRolloutStrategyType
)

func (s *Store) GetRolloutStrategy(ctx context.Context, namespaceKey, flagKey string) (*flipt.RolloutStrategy, error) {
	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}

	var (
		createdAt fliptsql.Timestamp
		updatedAt fliptsql.Timestamp

		strategy RolloutStrategyType
		rollout  = &flipt.RolloutStrategy{}

		err = s.builder.Select("id, namespace_key, flag_key, \"type\", name, created_at, updated_at").
			From("rollout_strategies").
			Where(sq.And{sq.Eq{"namespace_key": namespaceKey}, sq.Eq{"\"flag_key\"": flagKey}}).
			Limit(1).
			QueryRowContext(ctx).
			Scan(
				&rollout.NamespaceKey,
				&rollout.FlagKey,
				&strategy,
				&createdAt,
				&updatedAt)
	)

	if err != nil {
		return nil, err
	}

	rollout.CreatedAt = createdAt.Timestamp
	rollout.UpdatedAt = updatedAt.Timestamp

	switch strategy {
	case SegmentRolloutStrategyType:
		var segmentStrategy = &flipt.RolloutStrategy_RolloutStrategySegment{
			RolloutStrategySegment: &flipt.RolloutStrategySegment{},
		}

		if err := s.builder.Select("segment_key, \"value\"").
			From("rollout_strategy_segments").
			Where(sq.Eq{"rollout_strategy_id": rollout.Id}).
			Limit(1).
			QueryRowContext(ctx).
			Scan(
				&segmentStrategy.RolloutStrategySegment.SegmentKey,
				&segmentStrategy.RolloutStrategySegment.Value); err != nil {
			// TODO: log error instead?
			return nil, err
		}

		rollout.Strategy = segmentStrategy
	case PercentageRolloutStrategyType:
		var percentageStrategy = &flipt.RolloutStrategy_RolloutStrategyPercentage{
			RolloutStrategyPercentage: &flipt.RolloutStrategyPercentage{},
		}

		if err := s.builder.Select("percentage").
			From("rollout_strategy_percentages").
			Where(sq.Eq{"rollout_strategy_id": rollout.Id}).
			Limit(1).
			QueryRowContext(ctx).
			Scan(&percentageStrategy.RolloutStrategyPercentage.Percentage); err != nil {
			// TODO: log error instead?
			return nil, err
		}

	default:
		// TODO: log unknown strategy type
	}

	return rollout, nil
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
