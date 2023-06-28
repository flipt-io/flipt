package common

import (
	"context"
	"database/sql"

	sq "github.com/Masterminds/squirrel"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
)

func (s *Store) GetEvaluationRules(ctx context.Context, namespaceKey, flagKey string) ([]*storage.EvaluationRule, error) {
	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}
	// get all rules for flag with their constraints if any
	rows, err := s.builder.Select("r.id, r.namespace_key, r.flag_key, r.segment_key, s.match_type, r.\"rank\", c.id, c.type, c.property, c.operator, c.value").
		From("rules r").
		Join("segments s ON (r.segment_key = s.\"key\" AND r.namespace_key = s.namespace_key)").
		LeftJoin("constraints c ON (s.\"key\" = c.segment_key AND s.namespace_key = c.namespace_key)").
		Where(sq.And{sq.Eq{"r.flag_key": flagKey}, sq.Eq{"r.namespace_key": namespaceKey}}).
		OrderBy("r.\"rank\" ASC").
		GroupBy("r.id, c.id, s.match_type").
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
			tempRule           storage.EvaluationRule
			optionalConstraint optionalConstraint
		)

		if err := rows.Scan(
			&tempRule.ID,
			&tempRule.NamespaceKey,
			&tempRule.FlagKey,
			&tempRule.SegmentKey,
			&tempRule.SegmentMatchType,
			&tempRule.Rank,
			&optionalConstraint.Id,
			&optionalConstraint.Type,
			&optionalConstraint.Property,
			&optionalConstraint.Operator,
			&optionalConstraint.Value); err != nil {
			return rules, err
		}

		if existingRule, ok := uniqueRules[tempRule.ID]; ok {
			// current rule we know about
			if optionalConstraint.Id.Valid {
				constraint := storage.EvaluationConstraint{
					ID:       optionalConstraint.Id.String,
					Type:     flipt.ComparisonType(optionalConstraint.Type.Int32),
					Property: optionalConstraint.Property.String,
					Operator: optionalConstraint.Operator.String,
					Value:    optionalConstraint.Value.String,
				}
				existingRule.Constraints = append(existingRule.Constraints, constraint)
			}
		} else {
			// haven't seen this rule before
			newRule := &storage.EvaluationRule{
				ID:               tempRule.ID,
				NamespaceKey:     tempRule.NamespaceKey,
				FlagKey:          tempRule.FlagKey,
				SegmentKey:       tempRule.SegmentKey,
				SegmentMatchType: tempRule.SegmentMatchType,
				Rank:             tempRule.Rank,
			}

			if optionalConstraint.Id.Valid {
				constraint := storage.EvaluationConstraint{
					ID:       optionalConstraint.Id.String,
					Type:     flipt.ComparisonType(optionalConstraint.Type.Int32),
					Property: optionalConstraint.Property.String,
					Operator: optionalConstraint.Operator.String,
					Value:    optionalConstraint.Value.String,
				}
				newRule.Constraints = append(newRule.Constraints, constraint)
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

func (s *Store) GetEvaluationRollouts(ctx context.Context, flagKey, namespaceKey string) ([]*storage.EvaluationRollout, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT r.id, r.namespace_key, r."type", r."rank", rp.percentage, rp.value, rss.segment_key, rss.match_type, rss.constraint_type, rss.constraint_property, rss.constraint_operator, rss.constraint_value
		FROM rollouts r
		LEFT JOIN rollout_percentages rp ON (r.id = rp.rollout_id)
		LEFT JOIN (
			SELECT rs.rollout_id, rs.segment_key, s.match_type, rs.value AS rollout_segment_value, c."type" AS constraint_type, c.property AS constraint_property, c.operator AS constraint_operator, c.value AS constraint_value
			FROM rollout_segments rs
			JOIN segments s ON (rs.segment_key = s."key")
			JOIN constraints c ON (rs.segment_key = c.segment_key)
		) rss ON (r.id = rss.rollout_id)
		WHERE r.flag_key = $1 AND r.namespace_key = $2
		ORDER BY r."rank" ASC
		`, flagKey, namespaceKey)
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
			rpPercentageNumber   sql.NullFloat64
			rpPercentageValue    sql.NullBool
			rsSegmentKey         sql.NullString
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
			&rpPercentageNumber,
			&rpPercentageValue,
			&rsSegmentKey,
			&rsMatchType,
			&rsConstraintType,
			&rsConstraintProperty,
			&rsConstraintOperator,
			&rsConstraintValue,
		); err != nil {
			return rollouts, err
		}

		if rpPercentageNumber.Valid && rpPercentageValue.Valid {
			storagePercentage := &storage.RolloutPercentage{
				Percentage: float32(rpPercentageNumber.Float64),
				Value:      rpPercentageValue.Bool,
			}

			evaluationRollout.Percentage = storagePercentage
		}

		if rsSegmentKey.Valid &&
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
				SegmentKey:       rsSegmentKey.String,
				SegmentMatchType: flipt.MatchType(rsMatchType.Int32),
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
