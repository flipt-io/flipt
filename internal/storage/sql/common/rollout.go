package common

import (
	"context"
	"database/sql"
	"encoding/base64"
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

const (
	tableRollouts           = "rollouts"
	tableRolloutPercentages = "rollout_thresholds"
	tableRolloutSegments    = "rollout_segments"
)

func (s *Store) GetRollout(ctx context.Context, namespaceKey, id string) (*flipt.Rollout, error) {
	return getRollout(ctx, s.builder, namespaceKey, id)
}

func getRollout(ctx context.Context, builder sq.StatementBuilderType, namespaceKey, id string) (*flipt.Rollout, error) {
	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}

	var (
		createdAt fliptsql.Timestamp
		updatedAt fliptsql.Timestamp

		rollout = &flipt.Rollout{}

		err = builder.Select("id, namespace_key, flag_key, \"type\", \"rank\", description, created_at, updated_at").
			From(tableRollouts).
			Where(sq.And{sq.Eq{"id": id}, sq.Eq{"namespace_key": namespaceKey}}).
			QueryRowContext(ctx).
			Scan(
				&rollout.Id,
				&rollout.NamespaceKey,
				&rollout.FlagKey,
				&rollout.Type,
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

	switch rollout.Type {
	case flipt.RolloutType_SEGMENT_ROLLOUT_TYPE:
		var segmentRule = &flipt.Rollout_Segment{
			Segment: &flipt.RolloutSegment{},
		}

		if err := builder.Select("segment_key, \"value\"").
			From(tableRolloutSegments).
			Where(sq.And{sq.Eq{"rollout_id": rollout.Id}, sq.Eq{"namespace_key": rollout.NamespaceKey}}).
			Limit(1).
			QueryRowContext(ctx).
			Scan(
				&segmentRule.Segment.SegmentKey,
				&segmentRule.Segment.Value); err != nil {
			return nil, err
		}

		rollout.Rule = segmentRule
	case flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE:
		var thresholdRule = &flipt.Rollout_Threshold{
			Threshold: &flipt.RolloutThreshold{},
		}

		if err := builder.Select("percentage, \"value\"").
			From(tableRolloutPercentages).
			Where(sq.And{sq.Eq{"rollout_id": rollout.Id}, sq.Eq{"namespace_key": rollout.NamespaceKey}}).
			Limit(1).
			QueryRowContext(ctx).
			Scan(
				&thresholdRule.Threshold.Percentage,
				&thresholdRule.Threshold.Value); err != nil {
			return nil, err
		}

		rollout.Rule = thresholdRule

	default:
		return nil, fmt.Errorf("unknown rollout type %v", rollout.Type)
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
		token, err := decodePageToken(s.logger, params.PageToken)
		if err != nil {
			return results, err
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
		rolloutsByType = map[flipt.RolloutType][]*flipt.Rollout{}
	)

	for rows.Next() {
		var (
			rollout    = &flipt.Rollout{}
			rCreatedAt fliptsql.Timestamp
			rUpdatedAt fliptsql.Timestamp
		)

		if err := rows.Scan(
			&rollout.Id,
			&rollout.NamespaceKey,
			&rollout.FlagKey,
			&rollout.Type,
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
		rolloutsByType[rollout.Type] = append(rolloutsByType[rollout.Type], rollout)
	}

	if err := rows.Err(); err != nil {
		return results, err
	}

	if err := rows.Close(); err != nil {
		return results, err
	}

	// get all rules from rollout_segment_rules table
	if len(rolloutsByType[flipt.RolloutType_SEGMENT_ROLLOUT_TYPE]) > 0 {
		allRuleIds := make([]string, 0, len(rolloutsByType[flipt.RolloutType_SEGMENT_ROLLOUT_TYPE]))
		for _, rollout := range rolloutsByType[flipt.RolloutType_SEGMENT_ROLLOUT_TYPE] {
			allRuleIds = append(allRuleIds, rollout.Id)
		}

		rows, err := s.builder.Select("rollout_id, segment_key, \"value\"").
			From(tableRolloutSegments).
			Where(sq.Eq{"rollout_id": allRuleIds}).
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

		if err := rows.Err(); err != nil {
			return results, err
		}
	}

	// get all rules from rollout_percentage_rules table
	if len(rolloutsByType[flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE]) > 0 {
		allRuleIds := make([]string, 0, len(rolloutsByType[flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE]))
		for _, rollout := range rolloutsByType[flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE] {
			allRuleIds = append(allRuleIds, rollout.Id)
		}

		rows, err := s.builder.Select("rollout_id, percentage, \"value\"").
			From(tableRolloutPercentages).
			Where(sq.Eq{"rollout_id": allRuleIds}).
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
				rule      = &flipt.RolloutThreshold{}
			)

			if err := rows.Scan(&rolloutId, &rule.Percentage, &rule.Value); err != nil {
				return results, err
			}

			rollout := rolloutsById[rolloutId]
			rollout.Rule = &flipt.Rollout_Threshold{Threshold: rule}
		}

		if err := rows.Err(); err != nil {
			return results, err
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
		results.NextPageToken = base64.StdEncoding.EncodeToString(out)
	}

	return results, nil
}

// CountRollouts counts all rollouts
func (s *Store) CountRollouts(ctx context.Context, namespaceKey, flagKey string) (uint64, error) {
	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}

	var count uint64

	if err := s.builder.Select("COUNT(*)").
		From(tableRollouts).
		Where(sq.And{sq.Eq{"namespace_key": namespaceKey}, sq.Eq{"flag_key": flagKey}}).
		QueryRowContext(ctx).
		Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Store) CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (_ *flipt.Rollout, err error) {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	var count uint64

	if err := s.builder.Select("COUNT(*)").
		From(tableRollouts).
		Where(sq.And{sq.Eq{"namespace_key": r.NamespaceKey}, sq.Eq{"flag_key": r.FlagKey}, sq.Eq{"\"rank\"": r.Rank}}).
		QueryRowContext(ctx).
		Scan(&count); err != nil {
		return nil, err
	}

	if count > 0 {
		return nil, errs.ErrInvalidf("rank number: %d already exists", r.Rank)
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

	switch r.GetRule().(type) {
	case *flipt.CreateRolloutRequest_Segment:
		rollout.Type = flipt.RolloutType_SEGMENT_ROLLOUT_TYPE
	case *flipt.CreateRolloutRequest_Threshold:
		rollout.Type = flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE
	case nil:
		return nil, errs.ErrInvalid("rollout rule is missing")
	default:
		return nil, errs.ErrInvalidf("invalid rollout rule type %T", r.GetRule())
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err := s.builder.Insert(tableRollouts).
		RunWith(tx).
		Columns("id", "namespace_key", "flag_key", "\"type\"", "\"rank\"", "description", "created_at", "updated_at").
		Values(rollout.Id, rollout.NamespaceKey, rollout.FlagKey, rollout.Type, rollout.Rank, rollout.Description,
			&fliptsql.Timestamp{Timestamp: rollout.CreatedAt},
			&fliptsql.Timestamp{Timestamp: rollout.UpdatedAt},
		).ExecContext(ctx); err != nil {
		return nil, err
	}

	switch r.GetRule().(type) {
	case *flipt.CreateRolloutRequest_Segment:
		rollout.Type = flipt.RolloutType_SEGMENT_ROLLOUT_TYPE

		var segmentRule = r.GetSegment()

		if _, err := s.builder.Insert(tableRolloutSegments).
			RunWith(tx).
			Columns("id", "rollout_id", "namespace_key", "segment_key", "\"value\"").
			Values(uuid.Must(uuid.NewV4()).String(), rollout.Id, rollout.NamespaceKey, segmentRule.SegmentKey, segmentRule.Value).
			ExecContext(ctx); err != nil {
			return nil, err
		}

		rollout.Rule = &flipt.Rollout_Segment{
			Segment: segmentRule,
		}
	case *flipt.CreateRolloutRequest_Threshold:
		rollout.Type = flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE

		var thresholdRule = r.GetThreshold()

		if _, err := s.builder.Insert(tableRolloutPercentages).
			RunWith(tx).
			Columns("id", "rollout_id", "namespace_key", "percentage", "\"value\"").
			Values(uuid.Must(uuid.NewV4()).String(), rollout.Id, rollout.NamespaceKey, thresholdRule.Percentage, thresholdRule.Value).
			ExecContext(ctx); err != nil {
			return nil, err
		}

		rollout.Rule = &flipt.Rollout_Threshold{
			Threshold: thresholdRule,
		}
	default:
		return nil, fmt.Errorf("invalid rollout rule type %v", rollout.Type)
	}

	return rollout, tx.Commit()
}

func (s *Store) UpdateRollout(ctx context.Context, r *flipt.UpdateRolloutRequest) (_ *flipt.Rollout, err error) {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	if r.Id == "" {
		return nil, errs.ErrInvalid("rollout ID not supplied")
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// get current state for rollout
	rollout, err := getRollout(ctx, s.builder.RunWith(tx), r.NamespaceKey, r.Id)
	if err != nil {
		return nil, err
	}

	whereClause := sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}, sq.Eq{"namespace_key": r.NamespaceKey}}

	query := s.builder.Update(tableRollouts).
		RunWith(tx).
		Set("description", r.Description).
		Set("updated_at", &fliptsql.Timestamp{Timestamp: timestamppb.Now()}).
		Where(whereClause)

	res, err := query.ExecContext(ctx)
	if err != nil {
		return nil, err
	}

	count, err := res.RowsAffected()
	if err != nil {
		return nil, err
	}

	if count != 1 {
		return nil, errs.ErrNotFoundf(`rollout "%s/%s"`, r.NamespaceKey, r.Id)
	}

	switch r.Rule.(type) {
	case *flipt.UpdateRolloutRequest_Segment:
		// enforce that rollout type is consistent with the DB
		if err := ensureRolloutType(rollout, flipt.RolloutType_SEGMENT_ROLLOUT_TYPE); err != nil {
			return nil, err
		}

		var segmentRule = r.GetSegment()

		if _, err := s.builder.Update(tableRolloutSegments).
			RunWith(tx).
			Set("segment_key", segmentRule.SegmentKey).
			Set("value", segmentRule.Value).
			Where(sq.Eq{"rollout_id": r.Id}).ExecContext(ctx); err != nil {
			return nil, err
		}
	case *flipt.UpdateRolloutRequest_Threshold:
		// enforce that rollout type is consistent with the DB
		if err := ensureRolloutType(rollout, flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE); err != nil {
			return nil, err
		}

		var thresholdRule = r.GetThreshold()

		if _, err := s.builder.Update(tableRolloutPercentages).
			RunWith(tx).
			Set("percentage", thresholdRule.Percentage).
			Set("value", thresholdRule.Value).
			Where(sq.Eq{"rollout_id": r.Id}).ExecContext(ctx); err != nil {
			return nil, err
		}
	default:
		return nil, errs.InvalidFieldError("rule", "invalid rollout rule type")
	}

	rollout, err = getRollout(ctx, s.builder.RunWith(tx), r.NamespaceKey, r.Id)
	if err != nil {
		return nil, err
	}

	return rollout, tx.Commit()
}

func ensureRolloutType(rollout *flipt.Rollout, typ flipt.RolloutType) error {
	if rollout.Type == typ {
		return nil
	}

	return errs.ErrInvalidf(
		"cannot change type of rollout: have %q attempted %q",
		rollout.Type,
		typ,
	)
}

func (s *Store) DeleteRollout(ctx context.Context, r *flipt.DeleteRolloutRequest) error {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	_, err = s.builder.Delete(tableRollouts).
		RunWith(tx).
		Where(sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}, sq.Eq{"namespace_key": r.NamespaceKey}}).
		ExecContext(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// reorder existing rollouts after deletion
	rows, err := s.builder.Select("id").
		RunWith(tx).
		From(tableRollouts).
		Where(sq.And{sq.Eq{"namespace_key": r.NamespaceKey}, sq.Eq{"flag_key": r.FlagKey}}).
		OrderBy("\"rank\" ASC").
		QueryContext(ctx)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			_ = tx.Rollback()
			err = cerr
		}
	}()

	var rolloutIDs []string

	for rows.Next() {
		var rolloutID string

		if err := rows.Scan(&rolloutID); err != nil {
			_ = tx.Rollback()
			return err
		}

		rolloutIDs = append(rolloutIDs, rolloutID)
	}

	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err := s.orderRollouts(ctx, tx, r.NamespaceKey, r.FlagKey, rolloutIDs); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// OrderRollouts orders rollouts
func (s *Store) OrderRollouts(ctx context.Context, r *flipt.OrderRolloutsRequest) error {
	if r.NamespaceKey == "" {
		r.NamespaceKey = storage.DefaultNamespace
	}

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	if err := s.orderRollouts(ctx, tx, r.NamespaceKey, r.FlagKey, r.RolloutIds); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *Store) orderRollouts(ctx context.Context, runner sq.BaseRunner, namespaceKey, flagKey string, rolloutIDs []string) error {
	updatedAt := timestamppb.Now()

	for i, id := range rolloutIDs {
		_, err := s.builder.Update(tableRollouts).
			RunWith(runner).
			Set("\"rank\"", i+1).
			Set("updated_at", &fliptsql.Timestamp{Timestamp: updatedAt}).
			Where(sq.And{sq.Eq{"id": id}, sq.Eq{"namespace_key": namespaceKey}, sq.Eq{"flag_key": flagKey}}).
			ExecContext(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}
