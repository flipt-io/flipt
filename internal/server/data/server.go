package data

import (
	"context"
	"encoding/json"
	"fmt"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/data"
	"go.uber.org/zap"
)

type Server struct {
	logger *zap.Logger
	store  storage.Store

	data.UnimplementedDataServer
}

func NewServer(logger *zap.Logger, store storage.Store) *Server {
	return &Server{
		logger: logger,
		store:  store,
	}
}

func (s *Server) Snapshot(ctx context.Context, r *data.NamespaceRequest) (*data.NamespaceResponse, error) {
	var (
		namespaceKey = r.Key
		resp         = &data.NamespaceResponse{}
		remaining    = true
		nextPage     string
	)

	//  flags/variants in batches
	for batch := int32(0); remaining; batch++ {
		res, err := s.store.ListFlags(
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
			flag := &data.Flag{
				Key:         f.Key,
				Name:        f.Name,
				Type:        f.Type.String(),
				Description: f.Description,
				Enabled:     f.Enabled,
			}

			// map variant id => variant key
			variantKeys := make(map[string]string)

			for _, v := range f.Variants {
				var attachment interface{}

				if v.Attachment != "" {
					if err := json.Unmarshal([]byte(v.Attachment), &attachment); err != nil {
						return nil, fmt.Errorf("unmarshaling variant attachment: %w", err)
					}
				}

				flag.Variants = append(flag.Variants, &data.Variant{
					Key:         v.Key,
					Name:        v.Name,
					Description: v.Description,
					Attachment:  attachment,
				})

				variantKeys[v.Id] = v.Key
			}

			//  rules for flag
			resp, err := e.store.ListRules(
				ctx,
				&flipt.ListRuleRequest{
					NamespaceKey: namespaces[i],
					FlagKey:      flag.Key,
				},
			)
			if err != nil {
				return fmt.Errorf("getting rules for flag %q: %w", flag.Key, err)
			}

			rules := resp.Rules
			for _, r := range rules {
				rule := &Rule{}

				switch {
				case r.SegmentKey != "":
					rule.Segment = &SegmentEmbed{
						IsSegment: SegmentKey(r.SegmentKey),
					}
				case len(r.SegmentKeys) > 0:
					rule.Segment = &SegmentEmbed{
						IsSegment: &Segments{
							Keys:            r.SegmentKeys,
							SegmentOperator: r.SegmentOperator.String(),
						},
					}
				default:
					return fmt.Errorf("wrong format for rule segments")
				}

				for _, d := range r.Distributions {
					rule.Distributions = append(rule.Distributions, &Distribution{
						VariantKey: variantKeys[d.VariantId],
						Rollout:    d.Rollout,
					})
				}

				flag.Rules = append(flag.Rules, rule)
			}

			rollouts, err := e.store.ListRollouts(ctx, &flipt.ListRolloutRequest{
				NamespaceKey: namespaces[i],
				FlagKey:      flag.Key,
			})
			if err != nil {
				return fmt.Errorf("getting rollout rules for flag %q: %w", flag.Key, err)
			}

			for _, r := range rollouts.Rules {
				rollout := Rollout{
					Description: r.Description,
				}

				switch rule := r.Rule.(type) {
				case *flipt.Rollout_Segment:
					rollout.Segment = &SegmentRule{
						Value: rule.Segment.Value,
					}

					if rule.Segment.SegmentKey != "" {
						rollout.Segment.Key = rule.Segment.SegmentKey
					} else if len(rule.Segment.SegmentKeys) > 0 {
						rollout.Segment.Keys = rule.Segment.SegmentKeys
					}

					if rule.Segment.SegmentOperator == flipt.SegmentOperator_AND_SEGMENT_OPERATOR {
						rollout.Segment.Operator = rule.Segment.SegmentOperator.String()
					}
				case *flipt.Rollout_Threshold:
					rollout.Threshold = &ThresholdRule{
						Percentage: rule.Threshold.Percentage,
						Value:      rule.Threshold.Value,
					}
				}

				flag.Rollouts = append(flag.Rollouts, &rollout)
			}

			doc.Flags = append(doc.Flags, flag)
		}
	}

	remaining = true
	nextPage = ""

	//  segments/constraints in batches
	for batch := int32(0); remaining; batch++ {
		resp, err := e.store.ListSegments(
			ctx,
			&flipt.ListSegmentRequest{
				NamespaceKey: namespaces[i],
				PageToken:    nextPage,
				Limit:        batchSize,
			},
		)
		if err != nil {
			return fmt.Errorf("getting segments: %w", err)
		}

		segments := resp.Segments
		nextPage = resp.NextPageToken
		remaining = nextPage != ""

		for _, s := range segments {
			segment := &Segment{
				Key:         s.Key,
				Name:        s.Name,
				Description: s.Description,
				MatchType:   s.MatchType.String(),
			}

			for _, c := range s.Constraints {
				segment.Constraints = append(segment.Constraints, &Constraint{
					Type:        c.Type.String(),
					Property:    c.Property,
					Operator:    c.Operator,
					Value:       c.Value,
					Description: c.Description,
				})
			}

			doc.Segments = append(doc.Segments, segment)
		}
	}

	return nil, nil
}
