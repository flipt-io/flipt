package ext

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/markphelps/flipt/rpc/flipt"
	"github.com/markphelps/flipt/storage"
	"gopkg.in/yaml.v2"
)

const defaultBatchSize = 25

type lister interface {
	ListFlags(ctx context.Context, opts ...storage.QueryOption) ([]*flipt.Flag, error)
	ListSegments(ctx context.Context, opts ...storage.QueryOption) ([]*flipt.Segment, error)
	ListRules(ctx context.Context, flagKey string, opts ...storage.QueryOption) ([]*flipt.Rule, error)
}

type Exporter struct {
	store     lister
	batchSize uint64
}

func NewExporter(store lister) *Exporter {
	return &Exporter{
		store:     store,
		batchSize: defaultBatchSize,
	}
}

func (e *Exporter) Export(ctx context.Context, w io.Writer) error {
	var (
		enc       = yaml.NewEncoder(w)
		doc       = new(Document)
		batchSize = e.batchSize
	)

	defer enc.Close()

	var remaining = true

	// export flags/variants in batches
	for batch := uint64(0); remaining; batch++ {
		flags, err := e.store.ListFlags(ctx, storage.WithOffset(batch*batchSize), storage.WithLimit(batchSize))
		if err != nil {
			return fmt.Errorf("getting flags: %w", err)
		}

		remaining = uint64(len(flags)) == batchSize

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
			rules, err := e.store.ListRules(ctx, flag.Key)
			if err != nil {
				return fmt.Errorf("getting rules for flag %q: %w", flag.Key, err)
			}

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

	// export segments/constraints in batches
	for batch := uint64(0); remaining; batch++ {
		segments, err := e.store.ListSegments(ctx, storage.WithOffset(batch*batchSize), storage.WithLimit(batchSize))
		if err != nil {
			return fmt.Errorf("getting segments: %w", err)
		}

		remaining = uint64(len(segments)) == batchSize

		for _, s := range segments {
			segment := &Segment{
				Key:         s.Key,
				Name:        s.Name,
				Description: s.Description,
				MatchType:   s.MatchType.String(),
			}

			for _, c := range s.Constraints {
				segment.Constraints = append(segment.Constraints, &Constraint{
					Type:     c.Type.String(),
					Property: c.Property,
					Operator: c.Operator,
					Value:    c.Value,
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
