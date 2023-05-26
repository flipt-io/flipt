package common

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
	"go.flipt.io/flipt/internal/storage"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	"go.flipt.io/flipt/rpc/flipt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RolloutRuleType uint8

const (
	UnknownRolloutRuleType RolloutRuleType = iota
	SegmentRolloutRuleType
	PercentageRolloutRuleType

	tableRolloutRules           = "rollout_rules"
	tableRolloutPercentageRules = "rollout_percent_rules"
	tableRolloutSegmentRules    = "rollout_segment_rules"
)

func (s *Store) GetRolloutRule(ctx context.Context, namespaceKey, flagKey string) (*flipt.RolloutRule, error) {
	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}

	var (
		createdAt fliptsql.Timestamp
		updatedAt fliptsql.Timestamp

		rule    RolloutRuleType
		rollout = &flipt.RolloutRule{}

		err = s.builder.Select("id, namespace_key, flag_key, \"type\", created_at, updated_at").
			From(tableRolloutRules).
			Where(sq.And{sq.Eq{"namespace_key": namespaceKey}, sq.Eq{"\"flag_key\"": flagKey}}).
			Limit(1).
			QueryRowContext(ctx).
			Scan(
				&rollout.Id,
				&rollout.NamespaceKey,
				&rollout.FlagKey,
				&rule,
				&createdAt,
				&updatedAt)
	)

	if err != nil {
		return nil, err
	}

	rollout.CreatedAt = createdAt.Timestamp
	rollout.UpdatedAt = updatedAt.Timestamp

	switch rule {
	case SegmentRolloutRuleType:
		var segmentRule = &flipt.RolloutRule_Segment{
			Segment: &flipt.RolloutRuleSegment{},
		}

		if err := s.builder.Select("segment_key, \"value\"").
			From(tableRolloutSegmentRules).
			Where(sq.Eq{"rollout_rule_id": rollout.Id}).
			Limit(1).
			QueryRowContext(ctx).
			Scan(
				&segmentRule.Segment.SegmentKey,
				&segmentRule.Segment.Value); err != nil {
			// TODO: log error instead?
			return nil, err
		}

		rollout.Rule = segmentRule
	case PercentageRolloutRuleType:
		var percentageRule = &flipt.RolloutRule_Percentage{
			Percentage: &flipt.RolloutRulePercentage{},
		}

		if err := s.builder.Select("percentage").
			From(tableRolloutPercentageRules).
			Where(sq.Eq{"rollout_rule_id": rollout.Id}).
			Limit(1).
			QueryRowContext(ctx).
			Scan(&percentageRule.Percentage.Percentage); err != nil {
			// TODO: log error instead?
			return nil, err
		}

		rollout.Rule = percentageRule

	default:
		// TODO: log unknown rule type
	}

	return rollout, nil
}

func (s *Store) CreateRolloutRule(ctx context.Context, r *flipt.CreateRolloutRuleRequest) (*flipt.RolloutRule, error) {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	var rule RolloutRuleType

	if r.GetRule() != nil {
		if r.GetRule().GetSegment() != nil {
			s.logger.Debug("creating rollout rule segment")
			rule = SegmentRolloutRuleType
		} else if r.GetRule().GetPercentage() != nil {
			s.logger.Debug("creating rollout rule percent")
			rule = PercentageRolloutRuleType
		}
	}

	var (
		now     = timestamppb.Now()
		rollout = &flipt.RolloutRule{
			Id:           uuid.Must(uuid.NewV4()).String(),
			NamespaceKey: r.NamespaceKey,
			FlagKey:      r.FlagKey,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	)

	if _, err := s.builder.Insert(tableRolloutRules).
		Columns("id", "namespace_key", "flag_key", "\"type\"", "rank", "created_at", "updated_at").
		Values(rollout.Id, rollout.NamespaceKey, rollout.FlagKey, rule, rollout.Rank,
			&fliptsql.Timestamp{Timestamp: rollout.CreatedAt},
			&fliptsql.Timestamp{Timestamp: rollout.UpdatedAt},
		).ExecContext(ctx); err != nil {
		return nil, err
	}

	switch rule {
	case SegmentRolloutRuleType:
		var segmentRule = r.GetRule().GetSegment()

		if _, err := s.builder.Insert(tableRolloutSegmentRules).
			Columns("id", "rollout_rule_id", "segment_key", "\"value\"").
			Values(uuid.Must(uuid.NewV4()).String(), rollout.Id, segmentRule.SegmentKey, segmentRule.Value).
			ExecContext(ctx); err != nil {
			return nil, err
		}

		rollout.Rule = &flipt.RolloutRule_Segment{
			Segment: segmentRule,
		}
	case PercentageRolloutRuleType:
		var percentageRule = r.GetRule().GetPercentage()

		if _, err := s.builder.Insert(tableRolloutPercentageRules).
			Columns("id", "rollout_rule_id", "percentage", "\"value\"").
			Values(uuid.Must(uuid.NewV4()).String(), rollout.Id, percentageRule.Percentage, percentageRule.Value).
			ExecContext(ctx); err != nil {
			return nil, err
		}

		rollout.Rule = &flipt.RolloutRule_Percentage{
			Percentage: percentageRule,
		}
	}

	return rollout, nil

}

func (s *Store) UpdateRolloutRule(ctx context.Context, r *flipt.UpdateRolloutRuleRequest) (*flipt.RolloutRule, error) {
	panic("not implemented")
}

func (s *Store) DeleteRolloutRule(ctx context.Context, r *flipt.DeleteRolloutRuleRequest) error {
	panic("not implemented")
}
