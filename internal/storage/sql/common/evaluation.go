package common

import (
	"context"
	"database/sql"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
)

func (s *Store) GetEvaluationRules(ctx context.Context, namespaceKey, flagKey string) ([]*storage.EvaluationRule, error) {
	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}

	rows, err := s.builder.Select(`
    	r.id,
    	r.namespace_key,
    	r.flag_key,
    	rss.segment_key,
    	rss.segment_match_type,
		r.rule_segment_operator,
    	r.rank,
    	rss.constraint_id,
    	rss.constraint_type,
    	rss.constraint_property,
    	rss.constraint_operator,
    	rss.constraint_value`,
	).
		From("rules AS r").
		LeftJoin(`(
		SELECT
			rs.rule_id,
			rs.segment_key,
			s.match_type AS segment_match_type,
			c.id AS constraint_id,
			c.type AS constraint_type,
			c.property AS constraint_property,
			c.operator AS constraint_operator,
			c.value AS constraint_value
		FROM rule_segments AS rs
		JOIN segments AS s ON rs.segment_key = s.key
		LEFT JOIN constraints AS c ON (s.key = c.segment_key AND s.namespace_key = c.namespace_key)
	) rss ON (r.id = rss.rule_id)`).
		Where(sq.And{sq.Eq{"r.flag_key": flagKey}, sq.Eq{"r.namespace_key": namespaceKey}}).
		OrderBy("r.\"rank\" ASC").
		QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var (
		uniqueRules = make(map[string]*storage.EvaluationRule)
		rules       = []*storage.EvaluationRule{}
	)

	for rows.Next() {
		var (
			intermediateStorageRule struct {
				ID                  string
				NamespaceKey        string
				FlagKey             string
				SegmentKey          string
				SegmentMatchType    flipt.MatchType
				RuleSegmentOperator flipt.RuleSegmentOperator
				Rank                int32
			}
			optionalConstraint optionalConstraint
		)

		if err := rows.Scan(
			&intermediateStorageRule.ID,
			&intermediateStorageRule.NamespaceKey,
			&intermediateStorageRule.FlagKey,
			&intermediateStorageRule.SegmentKey,
			&intermediateStorageRule.SegmentMatchType,
			&intermediateStorageRule.RuleSegmentOperator,
			&intermediateStorageRule.Rank,
			&optionalConstraint.Id,
			&optionalConstraint.Type,
			&optionalConstraint.Property,
			&optionalConstraint.Operator,
			&optionalConstraint.Value); err != nil {
			return rules, err
		}

		fmt.Println("***RULE SEGMENT OPERATOR***: ", intermediateStorageRule.RuleSegmentOperator)

		if existingRule, ok := uniqueRules[intermediateStorageRule.ID]; ok {
			var constraint storage.EvaluationConstraint
			if optionalConstraint.Id.Valid {
				constraint = storage.EvaluationConstraint{
					ID:       optionalConstraint.Id.String,
					Type:     flipt.ComparisonType(optionalConstraint.Type.Int32),
					Property: optionalConstraint.Property.String,
					Operator: optionalConstraint.Operator.String,
					Value:    optionalConstraint.Value.String,
				}
			}

			segment, ok := existingRule.Segments[intermediateStorageRule.SegmentKey]
			if !ok {
				existingRule.Segments[intermediateStorageRule.SegmentKey] = &storage.EvaluationSegment{
					SegmentKey:  intermediateStorageRule.SegmentKey,
					MatchType:   intermediateStorageRule.SegmentMatchType,
					Constraints: []storage.EvaluationConstraint{constraint},
				}
			} else {
				segment.Constraints = append(segment.Constraints, constraint)
			}

			// Append to constraints if segment exists.
		} else {
			// haven't seen this rule before
			newRule := &storage.EvaluationRule{
				ID:                  intermediateStorageRule.ID,
				NamespaceKey:        intermediateStorageRule.NamespaceKey,
				FlagKey:             intermediateStorageRule.FlagKey,
				Rank:                intermediateStorageRule.Rank,
				RuleSegmentOperator: intermediateStorageRule.RuleSegmentOperator,
				Segments:            make(map[string]*storage.EvaluationSegment),
			}

			newRule.Segments[intermediateStorageRule.SegmentKey] = &storage.EvaluationSegment{
				SegmentKey: intermediateStorageRule.SegmentKey,
				MatchType:  intermediateStorageRule.SegmentMatchType,
			}

			if optionalConstraint.Id.Valid {
				constraint := storage.EvaluationConstraint{
					ID:       optionalConstraint.Id.String,
					Type:     flipt.ComparisonType(optionalConstraint.Type.Int32),
					Property: optionalConstraint.Property.String,
					Operator: optionalConstraint.Operator.String,
					Value:    optionalConstraint.Value.String,
				}

				if segment, ok := newRule.Segments[intermediateStorageRule.SegmentKey]; ok {
					segment.Constraints = []storage.EvaluationConstraint{constraint}
				}
			}

			uniqueRules[newRule.ID] = newRule
			rules = append(rules, newRule)
		}
	}

	if err := rows.Err(); err != nil {
		return rules, err
	}

	if err := rows.Close(); err != nil {
		return rules, err
	}

	return rules, nil
}

func (s *Store) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	rows, err := s.builder.Select("d.id, d.rule_id, d.variant_id, d.rollout, v.\"key\", v.attachment").
		From("distributions d").
		Join("variants v ON (d.variant_id = v.id)").
		Where(sq.Eq{"d.rule_id": ruleID}).
		OrderBy("d.created_at ASC").
		QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var distributions []*storage.EvaluationDistribution

	for rows.Next() {
		var (
			d          storage.EvaluationDistribution
			attachment sql.NullString
		)

		if err := rows.Scan(
			&d.ID, &d.RuleID, &d.VariantID, &d.Rollout, &d.VariantKey, &attachment,
		); err != nil {
			return distributions, err
		}

		if attachment.Valid {
			attachmentString, err := compactJSONString(attachment.String)
			if err != nil {
				return distributions, err
			}
			d.VariantAttachment = attachmentString
		}

		distributions = append(distributions, &d)
	}

	if err := rows.Err(); err != nil {
		return distributions, err
	}

	if err := rows.Close(); err != nil {
		return distributions, err
	}

	return distributions, nil
}

func (s *Store) GetEvaluationRollouts(ctx context.Context, namespaceKey, flagKey string) ([]*storage.EvaluationRollout, error) {
	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}

	rows, err := s.builder.Select(`
		r.id,
		r.namespace_key,
		r."type",
		r."rank",
		rt.percentage,
		rt.value,
		rss.segment_key,
		rss.rollout_segment_value,
		rss.match_type,
		rss.constraint_type,
		rss.constraint_property,
		rss.constraint_operator,
		rss.constraint_value
	`).
		From("rollouts r").
		LeftJoin("rollout_thresholds rt ON (r.id = rt.rollout_id)").
		LeftJoin(`(
		SELECT
			rs.rollout_id,
			rs.segment_key,
			s.match_type,
			rs.value AS rollout_segment_value,
			c."type" AS constraint_type,
			c.property AS constraint_property,
			c.operator AS constraint_operator,
			c.value AS constraint_value
		FROM rollout_segments rs
		JOIN segments s ON (rs.segment_key = s."key")
		JOIN constraints c ON (rs.segment_key = c.segment_key)
	) rss ON (r.id = rss.rollout_id)
	`).
		Where(sq.And{sq.Eq{"r.namespace_key": namespaceKey}, sq.Eq{"r.flag_key": flagKey}}).
		OrderBy(`r."rank" ASC`).
		QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := rows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var (
		uniqueSegmentedRollouts = make(map[string]*storage.EvaluationRollout)
		rollouts                = []*storage.EvaluationRollout{}
	)

	for rows.Next() {
		var (
			rolloutId            string
			evaluationRollout    storage.EvaluationRollout
			rtPercentageNumber   sql.NullFloat64
			rtPercentageValue    sql.NullBool
			rsSegmentKey         sql.NullString
			rsSegmentValue       sql.NullBool
			rsMatchType          sql.NullInt32
			rsConstraintType     sql.NullInt32
			rsConstraintProperty sql.NullString
			rsConstraintOperator sql.NullString
			rsConstraintValue    sql.NullString
		)

		if err := rows.Scan(
			&rolloutId,
			&evaluationRollout.NamespaceKey,
			&evaluationRollout.RolloutType,
			&evaluationRollout.Rank,
			&rtPercentageNumber,
			&rtPercentageValue,
			&rsSegmentKey,
			&rsSegmentValue,
			&rsMatchType,
			&rsConstraintType,
			&rsConstraintProperty,
			&rsConstraintOperator,
			&rsConstraintValue,
		); err != nil {
			return rollouts, err
		}

		if rtPercentageNumber.Valid && rtPercentageValue.Valid {
			storageThreshold := &storage.RolloutThreshold{
				Percentage: float32(rtPercentageNumber.Float64),
				Value:      rtPercentageValue.Bool,
			}

			evaluationRollout.Threshold = storageThreshold
		} else if rsSegmentKey.Valid &&
			rsSegmentValue.Valid &&
			rsMatchType.Valid &&
			rsConstraintType.Valid &&
			rsConstraintProperty.Valid &&
			rsConstraintOperator.Valid && rsConstraintValue.Valid {
			c := storage.EvaluationConstraint{
				Type:     flipt.ComparisonType(rsConstraintType.Int32),
				Property: rsConstraintProperty.String,
				Operator: rsConstraintOperator.String,
				Value:    rsConstraintValue.String,
			}

			if existingSegment, ok := uniqueSegmentedRollouts[rolloutId]; ok {
				existingSegment.Segment.Constraints = append(existingSegment.Segment.Constraints, c)
				continue
			}

			storageSegment := &storage.RolloutSegment{
				Key:       rsSegmentKey.String,
				Value:     rsSegmentValue.Bool,
				MatchType: flipt.MatchType(rsMatchType.Int32),
			}

			storageSegment.Constraints = append(storageSegment.Constraints, c)

			evaluationRollout.Segment = storageSegment
			uniqueSegmentedRollouts[rolloutId] = &evaluationRollout
		}

		rollouts = append(rollouts, &evaluationRollout)
	}

	if err := rows.Err(); err != nil {
		return rollouts, err
	}

	return rollouts, nil
}
