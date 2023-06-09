package ext

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"go.flipt.io/flipt/rpc/flipt"
	"gopkg.in/yaml.v2"
)

const (
	defaultBatchSize = 25
	version          = "1.0"
)

type Lister interface {
	ListFlags(context.Context, *flipt.ListFlagRequest) (*flipt.FlagList, error)
	ListSegments(context.Context, *flipt.ListSegmentRequest) (*flipt.SegmentList, error)
	ListRules(context.Context, *flipt.ListRuleRequest) (*flipt.RuleList, error)
}

type Exporter struct {
	store     Lister
	batchSize int32
	namespace string
}

func NewExporter(store Lister, namespace string) *Exporter {
	return &Exporter{
		store:     store,
		batchSize: defaultBatchSize,
		namespace: namespace,
	}
}

func (e *Exporter) Export(ctx context.Context, w io.Writer) error {
	var (
		enc       = yaml.NewEncoder(w)
		doc       = new(Document)
		batchSize = e.batchSize
	)

	doc.Version = version
	doc.Namespace = e.namespace

	defer enc.Close()

	var (
		remaining = true
		nextPage  string
	)

	// export flags/variants in batches
	for batch := int32(0); remaining; batch++ {
		resp, err := e.store.ListFlags(
			ctx,
			&flipt.ListFlagRequest{
				NamespaceKey: e.namespace,
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
					NamespaceKey: e.namespace,
					FlagKey:      flag.Key,
				},
			)
			if err != nil {
				return fmt.Errorf("getting rules for flag %q: %w", flag.Key, err)
			}

			rules := resp.Rules
			for _, r := range rules {
				rule := &Rule{
					SegmentKey: r.SegmentKey,
					Rank:       uint(r.Rank),
				}

				for _, d := range r.Distributions {
					rule.Distributions = append(rule.Distributions, &Distribution{
						VariantKey: variantKeys[d.VariantId],
						Rollout:    d.Rollout,
					})
				}

				flag.Rules = append(flag.Rules, rule)
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
				NamespaceKey: e.namespace,
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

	return nil
}
