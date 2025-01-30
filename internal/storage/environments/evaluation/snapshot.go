package evaluation

import (
	"context"
	"fmt"

	"github.com/blang/semver/v4"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
)

type minVersion struct {
	semver.Version
	msg string
}

func (m minVersion) String() string {
	return m.msg
}

func createStoreSnapshot(ctx context.Context, store storage.ReadOnlyStore, ref string) (*evaluation.EvaluationSnapshot, error) {
	namespaces, err := storage.ListAll(ctx, store.ListNamespaces, storage.ListAllParams{})
	if err != nil {
		return nil, err
	}

	snap := &evaluation.EvaluationSnapshot{
		Namespaces: map[string]*evaluation.EvaluationNamespaceSnapshot{},
	}

	for _, ns := range namespaces {
		var (
			namespaceKey = ns.Key
			resp         = &evaluation.EvaluationNamespaceSnapshot{
				Namespace: &evaluation.EvaluationNamespace{
					Key: namespaceKey,
				},
				Flags: make([]*evaluation.EvaluationFlag, 0),
			}
			remaining = true
			nextPage  string
			segments  = make(map[string]*evaluation.EvaluationSegment)
		)

		resp.Digest, err = store.GetDigest(ctx, storage.NewNamespace(namespaceKey))
		if err != nil {
			return nil, err
		}

		// preemptively set reference in namespaces map
		snap.Namespaces[namespaceKey] = resp

		//  flags/variants in batches
		for remaining {
			res, err := store.ListFlags(
				ctx,
				storage.ListWithOptions(
					storage.NewNamespace(namespaceKey, storage.WithReference(ref)),
					storage.ListWithQueryParamOptions[storage.NamespaceRequest](storage.WithPageToken(nextPage)),
				),
			)
			if err != nil {
				return nil, fmt.Errorf("getting flags: %w", err)
			}

			flags := res.Results
			nextPage = res.NextPageToken
			remaining = nextPage != ""

			for _, f := range flags {
				flag := &evaluation.EvaluationFlag{
					Key:         f.Key,
					Name:        f.Name,
					Description: f.Description,
					Enabled:     f.Enabled,
					Type:        toEvaluationFlagType(f.Type),
					CreatedAt:   f.CreatedAt,
					UpdatedAt:   f.UpdatedAt,
				}

				flagKey := storage.NewResource(namespaceKey, f.Key, storage.WithReference(ref))
				switch f.Type {
				// variant flag
				case flipt.FlagType_VARIANT_FLAG_TYPE:
					if f.DefaultVariant != nil {
						flag.DefaultVariant = &evaluation.EvaluationVariant{
							Id:         f.DefaultVariant.Id,
							Key:        f.DefaultVariant.Key,
							Attachment: f.DefaultVariant.Attachment,
						}
					}

					rules, err := store.GetEvaluationRules(ctx, flagKey)
					if err != nil {
						return nil, fmt.Errorf("getting rules for flag %q: %w", f.Key, err)
					}

					for _, r := range rules {
						rule := &evaluation.EvaluationRule{
							Id:              r.ID,
							Rank:            r.Rank,
							SegmentOperator: toEvaluationSegmentOperator(r.SegmentOperator),
						}

						for _, s := range r.Segments {
							// optimization: reuse segment if already seen
							if ss, ok := segments[s.SegmentKey]; ok {
								rule.Segments = append(rule.Segments, ss)
							} else {

								ss := &evaluation.EvaluationSegment{
									Key:       s.SegmentKey,
									MatchType: toEvaluationSegmentMatchType(s.MatchType),
								}

								for _, c := range s.Constraints {
									ss.Constraints = append(ss.Constraints, &evaluation.EvaluationConstraint{
										Id:       c.ID,
										Type:     toEvaluationConstraintComparisonType(c.Type),
										Property: c.Property,
										Operator: c.Operator,
										Value:    c.Value,
									})
								}

								segments[s.SegmentKey] = ss
								rule.Segments = append(rule.Segments, ss)
							}

							distributions, err := store.GetEvaluationDistributions(ctx, storage.NewID(r.ID, storage.WithReference(ref)))
							if err != nil {
								return nil, fmt.Errorf("getting distributions for rule %q: %w", r.ID, err)
							}

							// distributions for rule
							for _, d := range distributions {
								dist := &evaluation.EvaluationDistribution{
									VariantId:         d.VariantID,
									VariantKey:        d.VariantKey,
									VariantAttachment: d.VariantAttachment,
									Rollout:           d.Rollout,
								}
								rule.Distributions = append(rule.Distributions, dist)
							}

							flag.Rules = append(flag.Rules, rule)
						}

					}

				// boolean flag
				case flipt.FlagType_BOOLEAN_FLAG_TYPE:
					rollouts, err := store.GetEvaluationRollouts(ctx, flagKey)
					if err != nil {
						return nil, fmt.Errorf("getting rollout rules for flag %q: %w", f.Key, err)
					}

					for _, r := range rollouts {
						rollout := &evaluation.EvaluationRollout{
							Type: toEvaluationRolloutType(r.RolloutType),
							Rank: r.Rank,
						}

						switch r.RolloutType {
						case flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE:
							rollout.Rule = &evaluation.EvaluationRollout_Threshold{
								Threshold: &evaluation.EvaluationRolloutThreshold{
									Percentage: r.Threshold.Percentage,
									Value:      r.Threshold.Value,
								},
							}

						case flipt.RolloutType_SEGMENT_ROLLOUT_TYPE:
							segment := &evaluation.EvaluationRolloutSegment{
								Value:           r.Segment.Value,
								SegmentOperator: toEvaluationSegmentOperator(r.Segment.SegmentOperator),
							}

							for _, s := range r.Segment.Segments {
								// optimization: reuse segment if already seen
								ss, ok := segments[s.SegmentKey]
								if ok {
									segment.Segments = append(segment.Segments, ss)
									continue
								}

								ss = &evaluation.EvaluationSegment{
									Key:       s.SegmentKey,
									MatchType: toEvaluationSegmentMatchType(s.MatchType),
								}

								for _, c := range s.Constraints {
									ss.Constraints = append(ss.Constraints, &evaluation.EvaluationConstraint{
										Id:       c.ID,
										Type:     toEvaluationConstraintComparisonType(c.Type),
										Property: c.Property,
										Operator: c.Operator,
										Value:    c.Value,
									})
								}

								segments[s.SegmentKey] = ss
								segment.Segments = append(segment.Segments, ss)
							}

							rollout.Rule = &evaluation.EvaluationRollout_Segment{
								Segment: segment,
							}
						}

						flag.Rollouts = append(flag.Rollouts, rollout)
					}
				}

				resp.Flags = append(resp.Flags, flag)
			}
		}
	}

	return snap, nil
}

func toEvaluationFlagType(f flipt.FlagType) evaluation.EvaluationFlagType {
	switch f {
	case flipt.FlagType_BOOLEAN_FLAG_TYPE:
		return evaluation.EvaluationFlagType_BOOLEAN_FLAG_TYPE
	case flipt.FlagType_VARIANT_FLAG_TYPE:
		return evaluation.EvaluationFlagType_VARIANT_FLAG_TYPE
	}
	return evaluation.EvaluationFlagType_VARIANT_FLAG_TYPE
}

func toEvaluationSegmentMatchType(s flipt.MatchType) evaluation.EvaluationSegmentMatchType {
	switch s {
	case flipt.MatchType_ANY_MATCH_TYPE:
		return evaluation.EvaluationSegmentMatchType_ANY_SEGMENT_MATCH_TYPE
	case flipt.MatchType_ALL_MATCH_TYPE:
		return evaluation.EvaluationSegmentMatchType_ALL_SEGMENT_MATCH_TYPE
	}
	return evaluation.EvaluationSegmentMatchType_ANY_SEGMENT_MATCH_TYPE
}

func toEvaluationSegmentOperator(s flipt.SegmentOperator) evaluation.EvaluationSegmentOperator {
	switch s {
	case flipt.SegmentOperator_OR_SEGMENT_OPERATOR:
		return evaluation.EvaluationSegmentOperator_OR_SEGMENT_OPERATOR
	case flipt.SegmentOperator_AND_SEGMENT_OPERATOR:
		return evaluation.EvaluationSegmentOperator_AND_SEGMENT_OPERATOR
	}
	return evaluation.EvaluationSegmentOperator_OR_SEGMENT_OPERATOR
}

func toEvaluationConstraintComparisonType(c flipt.ComparisonType) evaluation.EvaluationConstraintComparisonType {
	switch c {
	case flipt.ComparisonType_STRING_COMPARISON_TYPE:
		return evaluation.EvaluationConstraintComparisonType_STRING_CONSTRAINT_COMPARISON_TYPE
	case flipt.ComparisonType_NUMBER_COMPARISON_TYPE:
		return evaluation.EvaluationConstraintComparisonType_NUMBER_CONSTRAINT_COMPARISON_TYPE
	case flipt.ComparisonType_DATETIME_COMPARISON_TYPE:
		return evaluation.EvaluationConstraintComparisonType_DATETIME_CONSTRAINT_COMPARISON_TYPE
	case flipt.ComparisonType_BOOLEAN_COMPARISON_TYPE:
		return evaluation.EvaluationConstraintComparisonType_BOOLEAN_CONSTRAINT_COMPARISON_TYPE
	case flipt.ComparisonType_ENTITY_ID_COMPARISON_TYPE:
		return evaluation.EvaluationConstraintComparisonType_ENTITY_ID_CONSTRAINT_COMPARISON_TYPE
	}
	return evaluation.EvaluationConstraintComparisonType_UNKNOWN_CONSTRAINT_COMPARISON_TYPE
}

func toEvaluationRolloutType(r flipt.RolloutType) evaluation.EvaluationRolloutType {
	switch r {
	case flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE:
		return evaluation.EvaluationRolloutType_THRESHOLD_ROLLOUT_TYPE
	case flipt.RolloutType_SEGMENT_ROLLOUT_TYPE:
		return evaluation.EvaluationRolloutType_SEGMENT_ROLLOUT_TYPE
	}
	return evaluation.EvaluationRolloutType_UNKNOWN_ROLLOUT_TYPE
}
