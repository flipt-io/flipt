package data

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type EvaluationStore interface {
	ListFlags(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Flag], error)
	storage.EvaluationStore
}

type Server struct {
	logger *zap.Logger
	store  EvaluationStore

	evaluation.UnimplementedDataServiceServer
}

func New(logger *zap.Logger, store EvaluationStore) *Server {
	return &Server{
		logger: logger,
		store:  store,
	}
}

// RegisterGRPC registers the *Server onto the provided grpc Server.
func (srv *Server) RegisterGRPC(server *grpc.Server) {
	evaluation.RegisterDataServiceServer(server, srv)
}

func (srv *Server) EvaluationSnapshotNamespace(ctx context.Context, r *evaluation.EvaluationNamespaceSnapshotRequest) (*evaluation.EvaluationNamespaceSnapshot, error) {
	var (
		namespaceKey = r.Key
		resp         = &evaluation.EvaluationNamespaceSnapshot{}
		remaining    = true
		nextPage     string
		segments     = make(map[string]*evaluation.EvaluationSegment)
	)

	//  flags/variants in batches
	for remaining {
		res, err := srv.store.ListFlags(
			ctx,
			namespaceKey,
			storage.WithPageToken(nextPage),
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
				Type:        f.Type,
				CreatedAt:   f.CreatedAt,
				UpdatedAt:   f.UpdatedAt,
			}

			if f.Type == flipt.FlagType_VARIANT_FLAG_TYPE {
				rules, err := srv.store.GetEvaluationRules(ctx, namespaceKey, f.Key)
				if err != nil {
					return nil, fmt.Errorf("getting rules for flag %q: %w", f.Key, err)
				}

				for _, r := range rules {
					rule := &evaluation.EvaluationRule{
						Id:              r.Id,
						Rank:            r.Rank,
						SegmentOperator: r.SegmentOperator,
					}

					for _, s := range r.Segments {
						// optimization: reuse segment if already seen
						if ss, ok := segments[s.Key]; ok {
							rule.Segments = append(rule.Segments, ss)
						} else {

							ss := &evaluation.EvaluationSegment{
								Key:       s.Key,
								MatchType: s.MatchType,
							}

							for _, c := range s.Constraints {
								ss.Constraints = append(ss.Constraints, &evaluation.EvaluationConstraint{
									Id:       c.Id,
									Type:     c.Type,
									Property: c.Property,
									Operator: c.Operator,
									Value:    c.Value,
								})
							}

							segments[s.Key] = ss
							rule.Segments = append(rule.Segments, ss)
						}

						distributions, err := srv.store.GetEvaluationDistributions(ctx, r.Id)
						if err != nil {
							return nil, fmt.Errorf("getting distributions for rule %q: %w", r.Id, err)
						}

						// distributions for rule
						for _, d := range distributions {
							dist := &evaluation.EvaluationDistribution{
								VariantId:         d.VariantId,
								VariantKey:        d.VariantKey,
								VariantAttachment: d.VariantAttachment,
								Rollout:           d.Rollout,
							}
							rule.Distributions = append(rule.Distributions, dist)
						}

						flag.Rules = append(flag.Rules, rule)
					}

				}

				if f.Type == flipt.FlagType_BOOLEAN_FLAG_TYPE {
					rollouts, err := srv.store.GetEvaluationRollouts(ctx, namespaceKey, f.Key)
					if err != nil {
						return nil, fmt.Errorf("getting rollout rules for flag %q: %w", f.Key, err)
					}

					for _, r := range rollouts {
						rollout := &evaluation.EvaluationRollout{
							Type: r.Type,
							Rank: r.Rank,
						}

						switch r.Type {
						case flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE:
							rollout.Rule = &evaluation.EvaluationRollout_Threshold{
								Threshold: &evaluation.EvaluationRolloutThreshold{
									Percentage: r.GetThreshold().Percentage,
									Value:      r.GetThreshold().Value,
								},
							}

						case flipt.RolloutType_SEGMENT_ROLLOUT_TYPE:
							segment := &evaluation.EvaluationRolloutSegment{
								Value:           r.GetSegment().Value,
								SegmentOperator: r.GetSegment().SegmentOperator,
							}

							for _, s := range r.GetSegment().Segments {
								// optimization: reuse segment if already seen
								ss, ok := segments[s.Key]
								if !ok {
									ss := &evaluation.EvaluationSegment{
										Key:       s.Key,
										MatchType: s.MatchType,
									}

									for _, c := range s.Constraints {
										ss.Constraints = append(ss.Constraints, &evaluation.EvaluationConstraint{
											Id:       c.Id,
											Type:     c.Type,
											Property: c.Property,
											Operator: c.Operator,
											Value:    c.Value,
										})
									}

									segments[s.Key] = ss
								}

								segment.Segments = append(segment.Segments, ss)
							}

							rollout.Rule = &evaluation.EvaluationRollout_Segment{
								Segment: segment,
							}
						}

						flag.Rollouts = append(flag.Rollouts, rollout)
					}
				}
			}
		}
	}

	return resp, nil
}
