package common

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
	"go.flipt.io/flipt/rpc/flipt"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type RolloutType uint8

const (
	UnknownRolloutType RolloutType = iota
	SegmentRolloutType
	PercentageRolloutType

	tableRollouts               = "rollouts"
	tableRolloutPercentageRules = "rollout_percentages"
	tableRolloutSegmentRules    = "rollout_segments"
)

func (s *Store) GetRollout(ctx context.Context, namespaceKey, id string) (*flipt.Rollout, error) {
	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}

	var (
		createdAt fliptsql.Timestamp
		updatedAt fliptsql.Timestamp

		ruleType RolloutType
		rollout  = &flipt.Rollout{}

		err = s.builder.Select("id, namespace_key, flag_key, \"type\", \"rank\", description, created_at, updated_at").
			From(tableRollouts).
			Where(sq.And{sq.Eq{"id": id}, sq.Eq{"namespace_key": namespaceKey}}).
			QueryRowContext(ctx).
			Scan(
				&rollout.Id,
				&rollout.NamespaceKey,
				&rollout.FlagKey,
				&ruleType,
				&rollout.Rank,
				&rollout.Description,
				&createdAt,
				&updatedAt)
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errs.ErrNotFoundf(`rollout "%s/%s"`, namespaceKey, id)
		}

		return nil, err
	}

	rollout.CreatedAt = createdAt.Timestamp
	rollout.UpdatedAt = updatedAt.Timestamp

	switch ruleType {
	case SegmentRolloutType:
		var segmentRule = &flipt.Rollout_Segment{
			Segment: &flipt.RolloutSegment{},
		}

		if err := s.builder.Select("segment_key, \"value\"").
			From(tableRolloutSegmentRules).
			Where(sq.And{sq.Eq{"rollout_rule_id": rollout.Id}, sq.Eq{"namespace_key": rollout.NamespaceKey}}).
			Limit(1).
			QueryRowContext(ctx).
			Scan(
				&segmentRule.Segment.SegmentKey,
				&segmentRule.Segment.Value); err != nil {
			// TODO: log error instead?
			return nil, err
		}

		rollout.Rule = segmentRule
	case PercentageRolloutType:
		var percentageRule = &flipt.Rollout_Percentage{
			Percentage: &flipt.RolloutPercentage{},
		}

		if err := s.builder.Select("percentage").
			From(tableRolloutPercentageRules).
			Where(sq.And{sq.Eq{"rollout_rule_id": rollout.Id}, sq.Eq{"namespace_key": rollout.NamespaceKey}}).
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

func (s *Store) ListRollouts(ctx context.Context, namespaceKey, flagKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Rollout], error) {
	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}

	params := &storage.QueryParams{}

	for _, opt := range opts {
		opt(params)
	}

	var (
		rollouts []*flipt.Rollout
		results  = storage.ResultSet[*flipt.Rollout]{}

		query = s.builder.Select("id, namespace_key, flag_key, \"type\", \"rank\", description, created_at, updated_at").
			From(tableRollouts).
			Where(sq.Eq{"flag_key": flagKey, "namespace_key": namespaceKey}).
			OrderBy(fmt.Sprintf("\"rank\" %s", params.Order))
	)

	if params.Limit > 0 {
		query = query.Limit(params.Limit + 1)
	}

	var offset uint64

	if params.PageToken != "" {
		var token PageToken

		if err := json.Unmarshal([]byte(params.PageToken), &token); err != nil {
			return results, fmt.Errorf("decoding page token %w", err)
		}

		offset = token.Offset
		query = query.Offset(offset)
	}

	rows, err := query.QueryContext(ctx)
	if err != nil {
		return results, err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var (
		rolloutsById   = map[string]*flipt.Rollout{}
		rolloutsByType = map[RolloutType][]*flipt.Rollout{}
	)

	for rows.Next() {
		var (
			rollout     = &flipt.Rollout{}
			rolloutType RolloutType
			rCreatedAt  fliptsql.Timestamp
			rUpdatedAt  fliptsql.Timestamp
		)

		if err := rows.Scan(
			&rollout.Id,
			&rollout.NamespaceKey,
			&rollout.FlagKey,
			&rolloutType,
			&rollout.Rank,
			&rollout.Description,
			&rCreatedAt,
			&rUpdatedAt); err != nil {
			return results, err
		}

		rollout.CreatedAt = rCreatedAt.Timestamp
		rollout.UpdatedAt = rUpdatedAt.Timestamp

		rollouts = append(rollouts, rollout)
		rolloutsById[rollout.Id] = rollout
		rolloutsByType[rolloutType] = append(rolloutsByType[rolloutType], rollout)
	}

	if err := rows.Err(); err != nil {
		return results, err
	}

	if err := rows.Close(); err != nil {
		return results, err
	}

	// get all rules from rollout_segment_rules table
	if len(rolloutsByType[SegmentRolloutType]) > 0 {
		allRuleIds := make([]string, len(rolloutsByType[SegmentRolloutType]))
		for _, rollout := range rolloutsByType[SegmentRolloutType] {
			allRuleIds = append(allRuleIds, rollout.Id)
		}

		rows, err := s.builder.Select("rollout_rule_id, segment_key, \"value\"").
			From(tableRolloutSegmentRules).
			Where(sq.Eq{"rollout_rule_id": allRuleIds}).
			QueryContext(ctx)

		if err != nil {
			return results, err
		}

		defer func() {
			if cerr := rows.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()

		for rows.Next() {
			var (
				rolloutId string
				rule      = &flipt.RolloutSegment{}
			)

			if err := rows.Scan(&rolloutId, &rule.SegmentKey, &rule.Value); err != nil {
				return results, err
			}

			rollout := rolloutsById[rolloutId]
			rollout.Rule = &flipt.Rollout_Segment{Segment: rule}
		}
	}
	// get all rules from rollout_percentage_rules table
	if len(rolloutsByType[PercentageRolloutType]) > 0 {
		allRuleIds := make([]string, len(rolloutsByType[PercentageRolloutType]))
		for _, rollout := range rolloutsByType[PercentageRolloutType] {
			allRuleIds = append(allRuleIds, rollout.Id)
		}

		rows, err := s.builder.Select("rollout_rule_id, percentage, \"value\"").
			From(tableRolloutPercentageRules).
			Where(sq.Eq{"rollout_rule_id": allRuleIds}).
			QueryContext(ctx)

		if err != nil {
			return results, err
		}

		defer func() {
			if cerr := rows.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()

		for rows.Next() {
			var (
				rolloutId string
				rule      = &flipt.RolloutPercentage{}
			)

			if err := rows.Scan(&rolloutId, &rule.Percentage, &rule.Value); err != nil {
				return results, err
			}

			rollout := rolloutsById[rolloutId]
			rollout.Rule = &flipt.Rollout_Percentage{Percentage: rule}
		}
	}

	var next *flipt.Rollout

	if len(rollouts) > int(params.Limit) && params.Limit > 0 {
		next = rollouts[len(rollouts)-1]
		rollouts = rollouts[:params.Limit]
	}

	results.Results = rollouts

	if next != nil {
		out, err := json.Marshal(PageToken{Key: next.Id, Offset: offset + uint64(len(rollouts))})
		if err != nil {
			return results, fmt.Errorf("encoding page token %w", err)
		}
		results.NextPageToken = string(out)
	}

	return results, nil
}

func (s *Store) CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (*flipt.Rollout, error) {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	var rule RolloutType

	if r.GetRule() != nil {
		if r.GetSegment() != nil {
			rule = SegmentRolloutType
		} else if r.GetPercentage() != nil {
			rule = PercentageRolloutType
		}
	}

	var (
		now     = timestamppb.Now()
		rollout = &flipt.Rollout{
			Id:           uuid.Must(uuid.NewV4()).String(),
			NamespaceKey: r.NamespaceKey,
			FlagKey:      r.FlagKey,
			Rank:         r.Rank,
			Description:  r.Description,
			CreatedAt:    now,
			UpdatedAt:    now,
		}
	)

	if _, err := s.builder.Insert(tableRollouts).
		Columns("id", "namespace_key", "flag_key", "\"type\"", "rank", "description", "created_at", "updated_at").
		Values(rollout.Id, rollout.NamespaceKey, rollout.FlagKey, rule, rollout.Rank, rollout.Description,
			&fliptsql.Timestamp{Timestamp: rollout.CreatedAt},
			&fliptsql.Timestamp{Timestamp: rollout.UpdatedAt},
		).ExecContext(ctx); err != nil {
		return nil, err
	}

	switch rule {
	case SegmentRolloutType:
		var segmentRule = r.GetSegment()

		if _, err := s.builder.Insert(tableRolloutSegmentRules).
			Columns("id", "rollout_rule_id", "namespace_key", "segment_key", "\"value\"").
			Values(uuid.Must(uuid.NewV4()).String(), rollout.Id, rollout.NamespaceKey, segmentRule.SegmentKey, segmentRule.Value).
			ExecContext(ctx); err != nil {
			return nil, err
		}

		rollout.Rule = &flipt.Rollout_Segment{
			Segment: segmentRule,
		}
	case PercentageRolloutType:
		var percentageRule = r.GetPercentage()

		if _, err := s.builder.Insert(tableRolloutPercentageRules).
			Columns("id", "rollout_rule_id", "namespace_key", "percentage", "\"value\"").
			Values(uuid.Must(uuid.NewV4()).String(), rollout.Id, rollout.NamespaceKey, percentageRule.Percentage, percentageRule.Value).
			ExecContext(ctx); err != nil {
			return nil, err
		}

		rollout.Rule = &flipt.Rollout_Percentage{
			Percentage: percentageRule,
		}
	}

	return rollout, nil
}

func (s *Store) UpdateRollout(ctx context.Context, r *flipt.UpdateRolloutRequest) (*flipt.Rollout, error) {
	panic("not implemented")
}

func (s *Store) DeleteRollout(ctx context.Context, r *flipt.DeleteRolloutRequest) error {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	_, err := s.builder.Delete(tableRollouts).
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"namespace_key": r.NamespaceKey}}).ExecContext(ctx)

	return err
}
