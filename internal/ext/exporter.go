package ext

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/blang/semver/v4"
	"go.flipt.io/flipt/rpc/flipt"
)

const (
	defaultBatchSize = 25
)

var (
	latestVersion     = semver.Version{Major: 1, Minor: 2}
	supportedVersions = semver.Versions{
		{Major: 1},
		{Major: 1, Minor: 1},
		latestVersion,
	}
)

type Lister interface {
	ListNamespaces(context.Context, *flipt.ListNamespaceRequest) (*flipt.NamespaceList, error)
	ListFlags(context.Context, *flipt.ListFlagRequest) (*flipt.FlagList, error)
	ListSegments(context.Context, *flipt.ListSegmentRequest) (*flipt.SegmentList, error)
	ListRules(context.Context, *flipt.ListRuleRequest) (*flipt.RuleList, error)
	ListRollouts(context.Context, *flipt.ListRolloutRequest) (*flipt.RolloutList, error)
}

type Exporter struct {
	store         Lister
	batchSize     int32
	namespaces    []string
	allNamespaces bool
}

func NewExporter(store Lister, namespaces string, allNamespaces bool) *Exporter {
	ns := strings.Split(namespaces, ",")

	return &Exporter{
		store:         store,
		batchSize:     defaultBatchSize,
		namespaces:    ns,
		allNamespaces: allNamespaces,
	}
}

// We currently only do minor bumps and print out just major.minor
func versionString(v semver.Version) string {
	return fmt.Sprintf("%d.%d", v.Major, v.Minor)
}

func (e *Exporter) Export(ctx context.Context, encoding Encoding, w io.Writer) error {
	var (
		enc       = encoding.NewEncoder(w)
		batchSize = e.batchSize
	)

	defer enc.Close()

	var namespaces = e.namespaces

	// If allNamespaces is "true", then retrieve all the namespaces, and store them in a string slice.
	if e.allNamespaces {
		var (
			remaining = true
			nextPage  string
		)

		intermediateNamespaces := make([]string, 0)

		for batch := int32(0); remaining; batch++ {
			resp, err := e.store.ListNamespaces(ctx, &flipt.ListNamespaceRequest{
				PageToken: nextPage,
				Limit:     batchSize,
			})
			if err != nil {
				return fmt.Errorf("getting namespaces: %w", err)
			}

			nextPage := resp.NextPageToken
			remaining = nextPage != ""

			for _, ns := range resp.Namespaces {
				intermediateNamespaces = append(intermediateNamespaces, ns.Key)
			}
		}

		namespaces = intermediateNamespaces
	}

	for i := 0; i < len(namespaces); i++ {
		doc := new(Document)
		// Only provide the version to the first document in the stream.
		if i == 0 {
			doc.Version = versionString(latestVersion)
		}
		doc.Namespace = namespaces[i]

		var (
			remaining = true
			nextPage  string
		)

		// export flags/variants in batches
		for batch := int32(0); remaining; batch++ {
			resp, err := e.store.ListFlags(
				ctx,
				&flipt.ListFlagRequest{
					NamespaceKey: namespaces[i],
					PageToken:    nextPage,
					Limit:        batchSize,
				},
			)
			if err != nil {
				return fmt.Errorf("getting flags: %w", err)
			}

			flags := resp.Flags
			nextPage = resp.NextPageToken
			remaining = nextPage != ""

			for _, f := range flags {
				flag := &Flag{
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
							return fmt.Errorf("unmarshaling variant attachment: %w", err)
						}
					}

					flag.Variants = append(flag.Variants, &Variant{
						Key:         v.Key,
						Name:        v.Name,
						Description: v.Description,
						Attachment:  attachment,
					})

					variantKeys[v.Id] = v.Key
				}

				// export rules for flag
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

		// export segments/constraints in batches
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

		if err := enc.Encode(doc); err != nil {
			return fmt.Errorf("marshaling document: %w", err)
		}
	}

	return nil
}
