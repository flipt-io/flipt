package data

import (
	"context"
	"fmt"
	"time"

	"github.com/opencontainers/go-digest"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/evaluation"
	"go.uber.org/zap"
	"golang.org/x/exp/slices"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type EvaluationStore interface {
	ListFlags(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Flag], error)
	ListNamespaces(ctx context.Context, opts ...storage.QueryOption) (storage.ResultSet[*flipt.Namespace], error)
	storage.EvaluationStore
}

type Server struct {
	logger *zap.Logger
	store  EvaluationStore

	servers chan evaluation.DataService_EvaluationSnapshotStreamServer

	evaluation.UnimplementedDataServiceServer
}

func New(ctx context.Context, logger *zap.Logger, store EvaluationStore) *Server {
	s := &Server{
		logger:  logger,
		store:   store,
		servers: make(chan evaluation.DataService_EvaluationSnapshotStreamServer),
	}

	go s.pollSnapshots(ctx)

	return s
}

// RegisterGRPC registers the *Server onto the provided grpc Server.
func (srv *Server) RegisterGRPC(server *grpc.Server) {
	evaluation.RegisterDataServiceServer(server, srv)
}

func (srv *Server) EvaluationSnapshotNamespace(ctx context.Context, r *evaluation.EvaluationNamespaceSnapshotRequest) (*evaluation.EvaluationNamespaceSnapshot, error) {
	return srv.getEvaluationSnapshotNamespace(ctx, r.Key)
}

func (srv *Server) EvaluationSnapshotStream(r *evaluation.EvaluationSnapshotStreamRequest, s evaluation.DataService_EvaluationSnapshotStreamServer) error {
	select {
	// return early if client goes away before the background loop adds
	// it to the set of monitored servers
	case <-s.Context().Done():
		return s.Context().Err()
	case srv.servers <- s:
	}

	// block until client goes away
	<-s.Context().Done()

	return nil
}

func (srv *Server) pollSnapshots(ctx context.Context) {
	var (
		lastDigest digest.Digest
		lastSnap   *evaluation.EvaluationSnapshot
		buildSnap  = func() (*evaluation.EvaluationSnapshot, error) {
			ns, err := storage.ListAll[struct{}, *flipt.Namespace](ctx, func(ctx context.Context, lr *storage.ListRequest[struct{}]) (storage.ResultSet[*flipt.Namespace], error) {
				return srv.store.ListNamespaces(ctx,
					storage.WithPageToken(lr.QueryParams.PageToken),
					storage.WithOrder(lr.QueryParams.Order),
					storage.WithLimit(lr.QueryParams.Limit))
			}, storage.ListAllParams{PerPage: 100})
			if err != nil {
				return nil, err
			}

			snap := evaluation.EvaluationSnapshot{
				Namespaces: map[string]*evaluation.EvaluationNamespaceSnapshot{},
			}

			for _, n := range ns {
				namespaceSnap, err := srv.getEvaluationSnapshotNamespace(ctx, n.Key)
				if err != nil {
					return nil, fmt.Errorf("namespace %q: %w", n.Key, err)
				}

				snap.Namespaces[n.Key] = namespaceSnap
			}

			return &snap, nil
		}
		servers []evaluation.DataService_EvaluationSnapshotStreamServer
	)

	ticker := time.NewTicker(10 * time.Second)
	for {
		select {
		case server := <-srv.servers:
			srv.logger.Debug("accepting new server")

			servers = append(servers, server)

			// if a snapshot hasn't been built yet then we build it
			if lastSnap == nil {
				var err error
				lastSnap, err = buildSnap()
				if err != nil {
					srv.logger.Error("building snapshot", zap.Error(err))
					continue
				}
			}

			if err := server.Send(lastSnap); err != nil {
				// drop client as stream has been cancelled
				if status.Code(err) != codes.Canceled {
					// don't drop incase error is transient
					srv.logger.Error("error sending snapshot", zap.Error(err))
				}
			}
		case <-ticker.C:
			if len(servers) == 0 {
				continue
			}

			snap, err := buildSnap()
			if err != nil {
				srv.logger.Error("building snapshot", zap.Error(err))
				continue
			}

			state, err := proto.Marshal(snap)
			if err != nil {
				srv.logger.Error("marshalling snapshot", zap.Error(err))
				continue
			}

			currentDigest := digest.FromBytes(state)
			if lastDigest == currentDigest {
				srv.logger.Debug("nothing changed, skipping send", zap.String("digest", currentDigest.Hex()))
				continue
			}

			lastDigest = currentDigest
			lastSnap = snap

			srv.logger.Debug("new snapshot build", zap.String("digest", lastDigest.Hex()))
			// we call send in a deletefunc loop so that we drop any
			// clients that have their streams cancelled
			servers = slices.DeleteFunc(servers, func(ds evaluation.DataService_EvaluationSnapshotStreamServer) bool {
				if err := ds.Context().Err(); err != nil {
					// delete any clients with closed contexts
					return true
				}

				if err := ds.Send(snap); err != nil {
					// drop client as stream has been cancelled
					if status.Code(err) == codes.Canceled {
						return true
					}

					// don't drop incase error is transient
					srv.logger.Error("error sending snapshot", zap.Error(err))
				}

				return false
			})
		case <-ctx.Done():
			return
		}
	}
}

func (srv *Server) getEvaluationSnapshotNamespace(ctx context.Context, namespaceKey string) (*evaluation.EvaluationNamespaceSnapshot, error) {
	var (
		resp = &evaluation.EvaluationNamespaceSnapshot{
			Namespace: &evaluation.EvaluationNamespace{ // TODO: should we get from store?
				Key: namespaceKey,
			},
			Flags: make([]*evaluation.EvaluationFlag, 0),
		}
		remaining = true
		nextPage  string
		segments  = make(map[string]*evaluation.EvaluationSegment)
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
				Type:        toEvaluationFlagType(f.Type),
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

						distributions, err := srv.store.GetEvaluationDistributions(ctx, r.ID)
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
			}

			if f.Type == flipt.FlagType_BOOLEAN_FLAG_TYPE {
				rollouts, err := srv.store.GetEvaluationRollouts(ctx, namespaceKey, f.Key)
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
							if !ok {
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

			resp.Flags = append(resp.Flags, flag)
		}
	}

	return resp, nil
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
