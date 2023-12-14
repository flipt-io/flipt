package fs

import (
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
)

// SnapshotFromEvaluationSnapshot returns a snapshot based on the provided evaluation snapshot.
// Due to this nature of this source, the resulting snapshot should only be used for evaluations,
// with the exception of it also supporting get and list flag operations.
func SnapshotFromEvaluationSnapshot(esnap *evaluation.EvaluationSnapshot) (*StoreSnapshot, error) {
	snap := StoreSnapshot{
		ns:        map[string]*namespace{},
		evalDists: map[string][]*storage.EvaluationDistribution{},
	}

	for ns, contents := range esnap.Namespaces {
		var (
			n = namespace{
				flags: map[string]*flipt.Flag{},
				resource: &flipt.Namespace{
					Key:  ns,
					Name: ns,
				},
				evalRules:    map[string][]*storage.EvaluationRule{},
				evalRollouts: map[string][]*storage.EvaluationRollout{},
			}
			dists []*storage.EvaluationDistribution
		)

		for _, f := range contents.Flags {
			n.flags[f.Key] = &flipt.Flag{
				NamespaceKey: ns,
				Key:          f.Key,
				Name:         f.Name,
				Description:  f.Description,
				Type:         flipt.FlagType(f.Type),
				Enabled:      f.Enabled,
				UpdatedAt:    f.UpdatedAt,
				CreatedAt:    f.CreatedAt,
			}

			if _, ok := n.evalRules[f.Key]; !ok {
				n.evalRules[f.Key] = []*storage.EvaluationRule{}
			}

			for _, rule := range f.Rules {
				r := &storage.EvaluationRule{
					NamespaceKey:    ns,
					FlagKey:         f.Key,
					ID:              rule.Id,
					Rank:            rule.Rank,
					SegmentOperator: flipt.SegmentOperator(rule.SegmentOperator),
				}

				for _, s := range rule.GetSegments() {
					var constraints []storage.EvaluationConstraint
					for _, c := range s.Constraints {
						constraints = append(constraints, storage.EvaluationConstraint{
							ID:       c.Id,
							Type:     flipt.ComparisonType(c.Type),
							Operator: c.Operator,
							Property: c.Property,
							Value:    c.Value,
						})
					}

					r.Segments[s.Key] = &storage.EvaluationSegment{
						SegmentKey:  s.Key,
						MatchType:   flipt.MatchType(s.MatchType),
						Constraints: constraints,
					}
				}

				n.evalRules[f.Key] = append(n.evalRules[f.Key], r)
			}

			if _, ok := n.evalRollouts[f.Key]; !ok {
				n.evalRollouts[f.Key] = []*storage.EvaluationRollout{}
			}

			for _, rollout := range f.Rollouts {
				r := &storage.EvaluationRollout{
					NamespaceKey: ns,
					Rank:         rollout.Rank,
					RolloutType:  flipt.RolloutType(rollout.Type),
				}

				switch rollout.Type {
				case evaluation.EvaluationRolloutType_SEGMENT_ROLLOUT_TYPE:
					r.Segment = &storage.RolloutSegment{
						SegmentOperator: flipt.SegmentOperator(rollout.GetSegment().SegmentOperator),
						Segments:        map[string]*storage.EvaluationSegment{},
					}
					for _, s := range rollout.GetSegment().Segments {
						var constraints []storage.EvaluationConstraint
						for _, c := range s.Constraints {
							constraints = append(constraints, storage.EvaluationConstraint{
								ID:       c.Id,
								Type:     flipt.ComparisonType(c.Type),
								Operator: c.Operator,
								Property: c.Property,
								Value:    c.Value,
							})
						}

						r.Segment.Segments[s.Key] = &storage.EvaluationSegment{
							SegmentKey:  s.Key,
							MatchType:   flipt.MatchType(s.MatchType),
							Constraints: constraints,
						}
					}
				case evaluation.EvaluationRolloutType_THRESHOLD_ROLLOUT_TYPE:
					r.Threshold = &storage.RolloutThreshold{
						Percentage: rollout.GetThreshold().Percentage,
						Value:      rollout.GetThreshold().Value,
					}
				}

				n.evalRollouts[f.Key] = append(n.evalRollouts[f.Key], r)
			}
		}

		snap.ns[ns] = &n
		snap.evalDists[ns] = dists
	}

	return &snap, nil
}
