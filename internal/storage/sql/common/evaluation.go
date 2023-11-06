package common

import (
	"context"
	"database/sql"
	"sort"

	sq "github.com/Masterminds/squirrel"
	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
)

type rule struct {
	ID              string
	Rank            int32
	SegmentOperator flipt.SegmentOperator
}

type intermediateRule struct {
	ID               string
	NamespaceKey     string
	FlagKey          string
	SegmentKey       string
	SegmentMatchType flipt.MatchType
	SegmentOperator  flipt.SegmentOperator
	Rank             int32
}

func (s *Store) GetEvaluationRules(ctx context.Context, namespaceKey, flagKey string) (_ []*storage.EvaluationRule, err error) {
	if namespaceKey == "" {
		namespaceKey = storage.DefaultNamespace
	}

	ruleRows, err := s.builder.
		Select("id, \"rank\", segment_operator").
		From("rules").
		Where(sq.And{sq.Eq{"flag_key": flagKey}, sq.Eq{"namespace_key": namespaceKey}}).
		QueryContext(ctx)
	if err != nil {
		return nil, err
	}

	defer func() {
		if cerr := ruleRows.Close(); cerr != nil && err == nil {
			err = cerr
		}
	}()

	var (
		uniqueRules = make(map[string]*rule)
		ruleIDs     = make([]string, 0)
	)

	for ruleRows.Next() {
		var r rule

		if err := ruleRows.Scan(&r.ID, &r.Rank, &r.SegmentOperator); err != nil {
			return nil, err
		}

		uniqueRules[r.ID] = &r
		ruleIDs = append(ruleIDs, r.ID)
	}

	if err := ruleRows.Err(); err != nil {
		return nil, err
	}

	if err := ruleRows.Close(); err != nil {
		return nil, err
	}

	rows, err := s.builder.Select(`
		rs.rule_id,
		rs.segment_key,
		s.match_type AS segment_match_type,
		c.id AS constraint_id,
		c."type" AS constraint_type,
		c.property AS constraint_property,
		c.operator AS constraint_operator,
		c.value AS constraint_value
	`).
		From("rule_segments AS rs").
		Join(`segments AS s ON rs.segment_key = s."key"`).
		LeftJoin(`constraints AS c ON (s."key" = c.segment_key AND s.namespace_key = c.namespace_key)`).
		Where(sq.Eq{"rs.rule_id": ruleIDs}).
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
		// ruleID -> rule
		uniqueEvaluationRules = make(map[string]*storage.EvaluationRule)
		// ruleID -> segmentKey -> segment
		uniqueEvaluationRulesSegments = make(map[string]map[string]*storage.EvaluationSegment)
		evaluationRules               = []*storage.EvaluationRule{}
	)

	for rows.Next() {
		var (
			optionalConstraint optionalConstraint
			intermediateRule   intermediateRule
		)

		if err := rows.Scan(
			&intermediateRule.ID,
			&intermediateRule.SegmentKey,
			&intermediateRule.SegmentMatchType,
			&optionalConstraint.Id,
			&optionalConstraint.Type,
			&optionalConstraint.Property,
			&optionalConstraint.Operator,
			&optionalConstraint.Value); err != nil {
			return evaluationRules, err
		}

		r := uniqueRules[intermediateRule.ID]

		intermediateRule.FlagKey = flagKey
		intermediateRule.NamespaceKey = namespaceKey
		intermediateRule.Rank = r.Rank
		intermediateRule.SegmentOperator = r.SegmentOperator

		// check if we've seen this rule before
		if rule, ok := uniqueEvaluationRules[intermediateRule.ID]; ok {
			var constraint *storage.EvaluationConstraint
			if optionalConstraint.Id.Valid {
				constraint = &storage.EvaluationConstraint{
					Id:       optionalConstraint.Id.String,
					Type:     flipt.ComparisonType(optionalConstraint.Type.Int32),
					Property: optionalConstraint.Property.String,
					Operator: optionalConstraint.Operator.String,
					Value:    optionalConstraint.Value.String,
				}
			}

			// check if we've seen this segment before for this rule
			segment, ok := uniqueEvaluationRulesSegments[intermediateRule.ID][intermediateRule.SegmentKey]
			if !ok {
				segment := &storage.EvaluationSegment{
					Key:       intermediateRule.SegmentKey,
					MatchType: intermediateRule.SegmentMatchType,
				}

				if constraint != nil {
					segment.Constraints = []*storage.EvaluationConstraint{constraint}
				}

				if _, ok := uniqueEvaluationRulesSegments[intermediateRule.ID]; !ok {
					uniqueEvaluationRulesSegments[intermediateRule.ID] = make(map[string]*storage.EvaluationSegment)
				}

				uniqueEvaluationRulesSegments[intermediateRule.ID][intermediateRule.SegmentKey] = segment
				rule.Segments = append(rule.Segments, segment)
			} else if constraint != nil {
				segment.Constraints = append(segment.Constraints, constraint)
			}
		} else {
			// haven't seen this rule before
			rule := &storage.EvaluationRule{
				Id:              intermediateRule.ID,
				Rank:            intermediateRule.Rank,
				SegmentOperator: intermediateRule.SegmentOperator,
				Segments:        []*storage.EvaluationSegment{},
			}

			var constraint *storage.EvaluationConstraint
			if optionalConstraint.Id.Valid {
				constraint = &storage.EvaluationConstraint{
					Id:       optionalConstraint.Id.String,
					Type:     flipt.ComparisonType(optionalConstraint.Type.Int32),
					Property: optionalConstraint.Property.String,
					Operator: optionalConstraint.Operator.String,
					Value:    optionalConstraint.Value.String,
				}
			}

			segment := &storage.EvaluationSegment{
				Key:       intermediateRule.SegmentKey,
				MatchType: intermediateRule.SegmentMatchType,
			}

			if constraint != nil {
				segment.Constraints = []*storage.EvaluationConstraint{constraint}
			}

			if _, ok := uniqueEvaluationRulesSegments[intermediateRule.ID]; !ok {
				uniqueEvaluationRulesSegments[intermediateRule.ID] = make(map[string]*storage.EvaluationSegment)
			}

			uniqueEvaluationRulesSegments[intermediateRule.ID][intermediateRule.SegmentKey] = segment
			rule.Segments = append(rule.Segments, segment)

			uniqueEvaluationRules[rule.Id] = rule
			evaluationRules = append(evaluationRules, rule)
		}
	}

	sort.Slice(evaluationRules, func(i, j int) bool {
		return evaluationRules[i].Rank < evaluationRules[j].Rank
	})

	if err := rows.Err(); err != nil {
		return evaluationRules, err
	}

	if err := rows.Close(); err != nil {
		return evaluationRules, err
	}

	return evaluationRules, nil
}

func (s *Store) GetEvaluationDistributions(ctx context.Context, ruleID string) (_ []*storage.EvaluationDistribution, err error) {
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
			&d.Id, &d.RuleId, &d.VariantId, &d.Rollout, &d.VariantKey, &attachment,
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

func (s *Store) GetEvaluationRollouts(ctx context.Context, namespaceKey, flagKey string) (_ []*storage.EvaluationRollout, err error) {
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
		rss.segment_operator,
		rss.match_type,
		rss.constraint_type,
		rss.constraint_property,
		rss.constraint_operator,
		rss.constraint_value
	`).
		From("rollouts AS r").
		LeftJoin("rollout_thresholds AS rt ON (r.id = rt.rollout_id)").
		LeftJoin(`(
		SELECT
			rs.rollout_id,
			rsr.segment_key,
			s.match_type,
			rs.value AS rollout_segment_value,
			rs.segment_operator AS segment_operator,
			c."type" AS constraint_type,
			c.property AS constraint_property,
			c.operator AS constraint_operator,
			c.value AS constraint_value
		FROM rollout_segments AS rs
		JOIN rollout_segment_references AS rsr ON (rs.id = rsr.rollout_segment_id)
		JOIN segments AS s ON (rsr.segment_key = s."key")
		LEFT JOIN constraints AS c ON (rsr.segment_key = c.segment_key)
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
		// rolloutId -> rollout
		uniqueSegmentedEvaluationRollouts = make(map[string]*storage.EvaluationRollout)
		// rolloutId -> segmentKey -> segment
		uniqueEvaluationRolloutsSegments = make(map[string]map[string]*storage.EvaluationSegment)
		evaluationRollouts               = []*storage.EvaluationRollout{}
	)

	for rows.Next() {
		var (
			rolloutId          string
			ns                 string
			evaluationRollout  storage.EvaluationRollout
			rtPercentageNumber sql.NullFloat64
			rtPercentageValue  sql.NullBool
			rsSegmentKey       sql.NullString
			rsSegmentValue     sql.NullBool
			rsSegmentOperator  sql.NullInt32
			rsMatchType        sql.NullInt32
			optionalConstraint optionalConstraint
		)

		if err := rows.Scan(
			&rolloutId,
			&ns,
			&evaluationRollout.Type,
			&evaluationRollout.Rank,
			&rtPercentageNumber,
			&rtPercentageValue,
			&rsSegmentKey,
			&rsSegmentValue,
			&rsSegmentOperator,
			&rsMatchType,
			&optionalConstraint.Type,
			&optionalConstraint.Property,
			&optionalConstraint.Operator,
			&optionalConstraint.Value,
		); err != nil {
			return evaluationRollouts, err
		}

		if rtPercentageNumber.Valid && rtPercentageValue.Valid {
			evaluationRollout.Rule = &evaluation.EvaluationRollout_Threshold{
				Threshold: &storage.RolloutThreshold{
					Percentage: float32(rtPercentageNumber.Float64),
					Value:      rtPercentageValue.Bool,
				},
			}

		} else if rsSegmentKey.Valid &&
			rsSegmentValue.Valid &&
			rsSegmentOperator.Valid &&
			rsMatchType.Valid {

			var constraint *storage.EvaluationConstraint
			if optionalConstraint.Type.Valid {
				constraint = &storage.EvaluationConstraint{
					Type:     flipt.ComparisonType(optionalConstraint.Type.Int32),
					Property: optionalConstraint.Property.String,
					Operator: optionalConstraint.Operator.String,
					Value:    optionalConstraint.Value.String,
				}
			}

			// check if we've seen this rollout before
			if rollout, ok := uniqueSegmentedEvaluationRollouts[rolloutId]; ok {
				// check if segment exists and either append constraints to an already existing segment,
				// or add another segment to the map.
				if segment, ok := uniqueEvaluationRolloutsSegments[rolloutId][rsSegmentKey.String]; ok {
					if constraint != nil {
						segment.Constraints = append(segment.Constraints, constraint)
					}
				} else {
					segment := &storage.EvaluationSegment{
						Key:       rsSegmentKey.String,
						MatchType: flipt.MatchType(rsMatchType.Int32),
					}

					if constraint != nil {
						segment.Constraints = []*storage.EvaluationConstraint{constraint}
					}

					if _, ok := uniqueEvaluationRolloutsSegments[rolloutId]; !ok {
						uniqueEvaluationRolloutsSegments[rolloutId] = make(map[string]*storage.EvaluationSegment)
					}

					if rollout.GetSegment() != nil {
						rollout.GetSegment().Segments = append(rollout.GetSegment().Segments, segment)
					} else {
						rollout.Rule = &evaluation.EvaluationRollout_Segment{
							Segment: &storage.RolloutSegment{
								Value:           rsSegmentValue.Bool,
								SegmentOperator: flipt.SegmentOperator(rsSegmentOperator.Int32),
								Segments:        []*storage.EvaluationSegment{segment},
							},
						}
					}

					uniqueEvaluationRolloutsSegments[rolloutId][rsSegmentKey.String] = segment
				}

				// this rollout has already been seen, so we can continue and dont need to add it to the slice
				continue
			}

			// haven't seen this rollout before, so we need to create it and the segment
			segment := &storage.EvaluationSegment{
				Key:       rsSegmentKey.String,
				MatchType: flipt.MatchType(rsMatchType.Int32),
			}

			if constraint != nil {
				segment.Constraints = []*storage.EvaluationConstraint{constraint}
			}

			evaluationRollout.Rule = &evaluation.EvaluationRollout_Segment{
				Segment: &storage.RolloutSegment{
					Value:           rsSegmentValue.Bool,
					SegmentOperator: flipt.SegmentOperator(rsSegmentOperator.Int32),
					Segments:        []*storage.EvaluationSegment{segment},
				},
			}

			uniqueSegmentedEvaluationRollouts[rolloutId] = &evaluationRollout

			if _, ok := uniqueEvaluationRolloutsSegments[rolloutId]; !ok {
				uniqueEvaluationRolloutsSegments[rolloutId] = make(map[string]*storage.EvaluationSegment)
			}

			uniqueEvaluationRolloutsSegments[rolloutId][rsSegmentKey.String] = segment
		}

		evaluationRollouts = append(evaluationRollouts, &evaluationRollout)
	}

	if err := rows.Err(); err != nil {
		return evaluationRollouts, err
	}

	if err := rows.Close(); err != nil {
		return evaluationRollouts, err
	}

	return evaluationRollouts, nil
}
