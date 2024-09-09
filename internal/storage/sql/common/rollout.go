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
)

const (
	tableRollouts                 = "rollouts"
	tableRolloutPercentages       = "rollout_thresholds"
	tableRolloutSegments          = "rollout_segments"
	tableRolloutSegmentReferences = "rollout_segment_references"
)

func (s *Store) GetRollout(ctx context.Context, ns storage.NamespaceRequest, id string) (*flipt.Rollout, error) {
	return getRollout(ctx, s.builder, ns, id)
}

func getRollout(ctx context.Context, builder sq.StatementBuilderType, ns storage.NamespaceRequest, id string) (*flipt.Rollout, error) {
	var (
		createdAt fliptsql.Timestamp
		updatedAt fliptsql.Timestamp

		rollout = &flipt.Rollout{}

		err = builder.Select("id, namespace_key, flag_key, \"type\", \"rank\", description, created_at, updated_at").
			From(tableRollouts).
			Where(sq.Eq{"id": id, "namespace_key": ns.Namespace()}).
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
			return nil, errs.ErrNotFoundf(`rollout "%s/%s"`, ns.Namespace(), id)
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

		var (
			value            bool
			rolloutSegmentId string
			segmentOperator  flipt.SegmentOperator
		)
		if err := builder.Select("id, \"value\", segment_operator").
			From(tableRolloutSegments).
			Where(sq.Eq{"rollout_id": rollout.Id}).
			Limit(1).
			QueryRowContext(ctx).
			Scan(&rolloutSegmentId, &value, &segmentOperator); err != nil {
			return nil, err
		}

		segmentRule.Segment.Value = value
		segmentRule.Segment.SegmentOperator = segmentOperator

		rows, err := builder.Select("segment_key").
			From(tableRolloutSegmentReferences).
			Where(sq.Eq{"rollout_segment_id": rolloutSegmentId, "namespace_key": rollout.NamespaceKey}).
			QueryContext(ctx)
		if err != nil {
			return nil, err
		}

		defer func() {
			if cerr := rows.Close(); cerr != nil && err == nil {
				err = cerr
			}
		}()

		var segmentKeys = []string{}

		for rows.Next() {
			var (
				segmentKey string
			)

			if err := rows.Scan(&segmentKey); err != nil {
				return nil, err
			}

			segmentKeys = append(segmentKeys, segmentKey)
		}

		if err := rows.Err(); err != nil {
			return nil, err
		}

		if len(segmentKeys) == 1 {
			segmentRule.Segment.SegmentKey = segmentKeys[0]
		} else {
			segmentRule.Segment.SegmentKeys = segmentKeys
		}

		rollout.Rule = segmentRule
	case flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE:
		var thresholdRule = &flipt.Rollout_Threshold{
			Threshold: &flipt.RolloutThreshold{},
		}

		if err := builder.Select("percentage, \"value\"").
			From(tableRolloutPercentages).
			Where(sq.Eq{"rollout_id": rollout.Id, "namespace_key": rollout.NamespaceKey}).
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

func (s *Store) ListRollouts(ctx context.Context, req *storage.ListRequest[storage.ResourceRequest]) (storage.ResultSet[*flipt.Rollout], error) {
	var (
		rollouts []*flipt.Rollout
		results  = storage.ResultSet[*flipt.Rollout]{}

		query = s.builder.Select("id, namespace_key, flag_key, \"type\", \"rank\", description, created_at, updated_at").
			From(tableRollouts).
			Where(sq.Eq{"flag_key": req.Predicate.Key, "namespace_key": req.Predicate.Namespace()}).
			OrderBy(fmt.Sprintf("\"rank\" %s", req.QueryParams.Order))
	)

	if req.QueryParams.Limit > 0 {
		query = query.Limit(req.QueryParams.Limit + 1)
	}

	var offset uint64

	if req.QueryParams.PageToken != "" {
		token, err := decodePageToken(s.logger, req.QueryParams.PageToken)
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

		rows, err := s.builder.Select("rs.rollout_id, rs.\"value\", rs.segment_operator, rsr.segment_key").
			From("rollout_segments AS rs").
			Join("rollout_segment_references AS rsr ON (rs.id = rsr.rollout_segment_id)").
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

		type intermediateValues struct {
			segmentKeys     []string
			segmentOperator flipt.SegmentOperator
			value           bool
		}

		intermediate := make(map[string]*intermediateValues)

		for rows.Next() {
			var (
				rolloutId       string
				segmentKey      string
				value           bool
				segmentOperator flipt.SegmentOperator
			)

			if err := rows.Scan(&rolloutId, &value, &segmentOperator, &segmentKey); err != nil {
				return results, err
			}

			rs, ok := intermediate[rolloutId]
			if ok {
				rs.segmentKeys = append(rs.segmentKeys, segmentKey)
			} else {
				intermediate[rolloutId] = &intermediateValues{
					segmentKeys:     []string{segmentKey},
					segmentOperator: segmentOperator,
					value:           value,
				}
			}
		}

		for k, v := range intermediate {
			rollout := rolloutsById[k]
			rs := &flipt.RolloutSegment{}

			if len(v.segmentKeys) == 1 {
				rs.SegmentKey = v.segmentKeys[0]
			} else {
				rs.SegmentKeys = v.segmentKeys
			}

			rs.Value = v.value
			rs.SegmentOperator = v.segmentOperator

			rollout.Rule = &flipt.Rollout_Segment{Segment: rs}
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

	if len(rollouts) > int(req.QueryParams.Limit) && req.QueryParams.Limit > 0 {
		next = rollouts[len(rollouts)-1]
		rollouts = rollouts[:req.QueryParams.Limit]
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
func (s *Store) CountRollouts(ctx context.Context, flag storage.ResourceRequest) (uint64, error) {
	var count uint64

	if err := s.builder.Select("COUNT(*)").
		From(tableRollouts).
		Where(sq.Eq{"namespace_key": flag.Namespace(), "flag_key": flag.Key}).
		QueryRowContext(ctx).
		Scan(&count); err != nil {
		return 0, err
	}

	return count, nil
}

func (s *Store) CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (_ *flipt.Rollout, err error) {
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.NamespaceKey)
		}
	}()

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
		now     = flipt.Now()
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
		Values(rollout.Id, rollout.NamespaceKey, rollout.FlagKey, int32(rollout.Type), rollout.Rank, rollout.Description,
			&fliptsql.Timestamp{Timestamp: rollout.CreatedAt},
			&fliptsql.Timestamp{Timestamp: rollout.UpdatedAt},
		).ExecContext(ctx); err != nil {
		return nil, err
	}

	switch r.GetRule().(type) {
	case *flipt.CreateRolloutRequest_Segment:
		rollout.Type = flipt.RolloutType_SEGMENT_ROLLOUT_TYPE
		rolloutSegmentId := uuid.Must(uuid.NewV4()).String()

		var segmentRule = r.GetSegment()

		segmentKeys := sanitizeSegmentKeys(segmentRule.GetSegmentKey(), segmentRule.GetSegmentKeys())

		var segmentOperator = segmentRule.SegmentOperator
		if len(segmentKeys) == 1 {
			segmentOperator = flipt.SegmentOperator_OR_SEGMENT_OPERATOR
		}

		if _, err := s.builder.Insert(tableRolloutSegments).
			RunWith(tx).
			Columns("id", "rollout_id", "\"value\"", "segment_operator").
			Values(rolloutSegmentId, rollout.Id, segmentRule.Value, int32(segmentOperator)).
			ExecContext(ctx); err != nil {
			return nil, err
		}

		for _, segmentKey := range segmentKeys {
			if _, err := s.builder.Insert(tableRolloutSegmentReferences).
				RunWith(tx).
				Columns("rollout_segment_id", "namespace_key", "segment_key").
				Values(rolloutSegmentId, rollout.NamespaceKey, segmentKey).
				ExecContext(ctx); err != nil {
				return nil, err
			}
		}

		innerSegment := &flipt.RolloutSegment{
			Value:           segmentRule.Value,
			SegmentOperator: segmentOperator,
		}

		if len(segmentKeys) == 1 {
			innerSegment.SegmentKey = segmentKeys[0]
		} else {
			innerSegment.SegmentKeys = segmentKeys
		}

		rollout.Rule = &flipt.Rollout_Segment{
			Segment: innerSegment,
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
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.NamespaceKey)
		}
	}()

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

	ns := storage.NewNamespace(r.NamespaceKey)
	// get current state for rollout
	rollout, err := getRollout(ctx, s.builder.RunWith(tx), ns, r.Id)
	if err != nil {
		return nil, err
	}

	whereClause := sq.And{sq.Eq{"id": r.Id}, sq.Eq{"flag_key": r.FlagKey}, sq.Eq{"namespace_key": r.NamespaceKey}}

	query := s.builder.Update(tableRollouts).
		RunWith(tx).
		Set("description", r.Description).
		Set("updated_at", &fliptsql.Timestamp{Timestamp: flipt.Now()}).
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

		segmentKeys := sanitizeSegmentKeys(segmentRule.GetSegmentKey(), segmentRule.GetSegmentKeys())

		var segmentOperator = segmentRule.SegmentOperator
		if len(segmentKeys) == 1 {
			segmentOperator = flipt.SegmentOperator_OR_SEGMENT_OPERATOR
		}

		if _, err := s.builder.Update(tableRolloutSegments).
			RunWith(tx).
			Set("segment_operator", segmentOperator).
			Set("value", segmentRule.Value).
			Where(sq.Eq{"rollout_id": r.Id}).ExecContext(ctx); err != nil {
			return nil, err
		}

		// Delete and reinsert rollout_segment_references.
		row := s.builder.Select("id").
			RunWith(tx).
			From(tableRolloutSegments).
			Where(sq.Eq{"rollout_id": r.Id}).
			Limit(1).
			QueryRowContext(ctx)

		var rolloutSegmentId string

		if err := row.Scan(&rolloutSegmentId); err != nil {
			return nil, err
		}

		if _, err := s.builder.Delete(tableRolloutSegmentReferences).
			RunWith(tx).
			Where(sq.And{sq.Eq{"rollout_segment_id": rolloutSegmentId}, sq.Eq{"namespace_key": r.NamespaceKey}}).
			ExecContext(ctx); err != nil {
			return nil, err
		}

		for _, segmentKey := range segmentKeys {
			if _, err := s.builder.
				Insert(tableRolloutSegmentReferences).
				RunWith(tx).
				Columns("rollout_segment_id", "namespace_key", "segment_key").
				Values(
					rolloutSegmentId,
					r.NamespaceKey,
					segmentKey,
				).
				ExecContext(ctx); err != nil {
				return nil, err
			}
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

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	rollout, err = getRollout(ctx, s.builder, ns, r.Id)
	if err != nil {
		return nil, err
	}

	return rollout, nil
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

func (s *Store) DeleteRollout(ctx context.Context, r *flipt.DeleteRolloutRequest) (err error) {
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.NamespaceKey)
		}
	}()

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
func (s *Store) OrderRollouts(ctx context.Context, r *flipt.OrderRolloutsRequest) (err error) {
	defer func() {
		if err == nil {
			err = s.setVersion(ctx, r.NamespaceKey)
		}
	}()

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
	updatedAt := flipt.Now()

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
