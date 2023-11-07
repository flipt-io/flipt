package ext

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/blang/semver/v4"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/rpc/flipt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Creator interface {
	GetNamespace(context.Context, *flipt.GetNamespaceRequest) (*flipt.Namespace, error)
	CreateNamespace(context.Context, *flipt.CreateNamespaceRequest) (*flipt.Namespace, error)
	CreateFlag(context.Context, *flipt.CreateFlagRequest) (*flipt.Flag, error)
	CreateVariant(context.Context, *flipt.CreateVariantRequest) (*flipt.Variant, error)
	CreateSegment(context.Context, *flipt.CreateSegmentRequest) (*flipt.Segment, error)
	CreateConstraint(context.Context, *flipt.CreateConstraintRequest) (*flipt.Constraint, error)
	CreateRule(context.Context, *flipt.CreateRuleRequest) (*flipt.Rule, error)
	CreateDistribution(context.Context, *flipt.CreateDistributionRequest) (*flipt.Distribution, error)
	CreateRollout(context.Context, *flipt.CreateRolloutRequest) (*flipt.Rollout, error)
}

type Importer struct {
	creator Creator
}

type ImportOpt func(*Importer)

func NewImporter(store Creator, opts ...ImportOpt) *Importer {
	i := &Importer{
		creator: store,
	}

	for _, opt := range opts {
		opt(i)
	}

	return i
}

func (i *Importer) Import(ctx context.Context, enc Encoding, r io.Reader) (err error) {
	var (
		dec     = enc.NewDecoder(r)
		version semver.Version
	)

	idx := 0

	for {
		var doc = new(Document)
		if err := dec.Decode(doc); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("unmarshalling document: %w", err)
		}

		// Only support parsing vesrion at the top of each import file.
		if idx == 0 {
			version = latestVersion
			if doc.Version != "" {
				version, err = semver.ParseTolerant(doc.Version)
				if err != nil {
					return fmt.Errorf("parsing document version: %w", err)
				}

				var found bool
				for _, sv := range supportedVersions {
					if found = sv.EQ(version); found {
						break
					}
				}

				if !found {
					return fmt.Errorf("unsupported version: %s", doc.Version)
				}
			}
		}

		var namespace = doc.Namespace

		if namespace != "" && namespace != flipt.DefaultNamespace {
			_, err := i.creator.GetNamespace(ctx, &flipt.GetNamespaceRequest{
				Key: namespace,
			})

			if err != nil {
				if status.Code(err) != codes.NotFound && !errs.AsMatch[errs.ErrNotFound](err) {
					return err
				}

				_, err = i.creator.CreateNamespace(ctx, &flipt.CreateNamespaceRequest{
					Key:  namespace,
					Name: namespace,
				})
				if err != nil {
					return err
				}
			}
		}

		var (
			// map flagKey => *flag
			createdFlags = make(map[string]*flipt.Flag)
			// map segmentKey => *segment
			createdSegments = make(map[string]*flipt.Segment)
			// map flagKey:variantKey => *variant
			createdVariants = make(map[string]*flipt.Variant)
		)

		// create flags/variants
		for _, f := range doc.Flags {
			if f == nil {
				continue
			}

			req := &flipt.CreateFlagRequest{
				Key:          f.Key,
				Name:         f.Name,
				Description:  f.Description,
				Enabled:      f.Enabled,
				NamespaceKey: namespace,
			}

			// support explicitly setting flag type from 1.1
			if f.Type != "" {
				if err := ensureFieldSupported("flag.type", semver.Version{
					Major: 1,
					Minor: 1,
				}, version); err != nil {
					return err
				}

				req.Type = flipt.FlagType(flipt.FlagType_value[f.Type])
			}

			flag, err := i.creator.CreateFlag(ctx, req)
			if err != nil {
				return fmt.Errorf("creating flag: %w", err)
			}

			for _, v := range f.Variants {
				if v == nil {
					continue
				}

				var out []byte

				if v.Attachment != nil {
					converted := convert(v.Attachment)
					out, err = json.Marshal(converted)
					if err != nil {
						return fmt.Errorf("marshalling attachment: %w", err)
					}
				}

				variant, err := i.creator.CreateVariant(ctx, &flipt.CreateVariantRequest{
					FlagKey:      f.Key,
					Key:          v.Key,
					Name:         v.Name,
					Description:  v.Description,
					Attachment:   string(out),
					NamespaceKey: namespace,
				})

				if err != nil {
					return fmt.Errorf("creating variant: %w", err)
				}

				createdVariants[fmt.Sprintf("%s:%s", flag.Key, variant.Key)] = variant
			}

			createdFlags[flag.Key] = flag
		}

		// create segments/constraints
		for _, s := range doc.Segments {
			if s == nil {
				continue
			}

			segment, err := i.creator.CreateSegment(ctx, &flipt.CreateSegmentRequest{
				Key:          s.Key,
				Name:         s.Name,
				Description:  s.Description,
				MatchType:    flipt.MatchType(flipt.MatchType_value[s.MatchType]),
				NamespaceKey: namespace,
			})

			if err != nil {
				return fmt.Errorf("creating segment: %w", err)
			}

			for _, c := range s.Constraints {
				if c == nil {
					continue
				}

				_, err := i.creator.CreateConstraint(ctx, &flipt.CreateConstraintRequest{
					SegmentKey:   s.Key,
					Type:         flipt.ComparisonType(flipt.ComparisonType_value[c.Type]),
					Property:     c.Property,
					Operator:     c.Operator,
					Value:        c.Value,
					NamespaceKey: namespace,
				})

				if err != nil {
					return fmt.Errorf("creating constraint: %w", err)
				}
			}

			createdSegments[segment.Key] = segment
		}

		// create rules/distributions
		for _, f := range doc.Flags {
			if f == nil {
				continue
			}

			// loop through rules
			for idx, r := range f.Rules {
				if r == nil {
					continue
				}

				// support implicit rank from version >=1.1
				rank := int32(r.Rank)
				if rank == 0 && version.GE(semver.Version{Major: 1, Minor: 1}) {
					rank = int32(idx) + 1
				}

				fcr := &flipt.CreateRuleRequest{
					FlagKey:      f.Key,
					Rank:         rank,
					NamespaceKey: namespace,
				}

				switch s := r.Segment.IsSegment.(type) {
				case SegmentKey:
					fcr.SegmentKey = string(s)
				case *Segments:
					fcr.SegmentKeys = s.Keys
					fcr.SegmentOperator = flipt.SegmentOperator(flipt.SegmentOperator_value[s.SegmentOperator])
				}

				rule, err := i.creator.CreateRule(ctx, fcr)

				if err != nil {
					return fmt.Errorf("creating rule: %w", err)
				}

				for _, d := range r.Distributions {
					if d == nil {
						continue
					}

					variant, found := createdVariants[fmt.Sprintf("%s:%s", f.Key, d.VariantKey)]
					if !found {
						return fmt.Errorf("finding variant: %s; flag: %s", d.VariantKey, f.Key)
					}

					_, err := i.creator.CreateDistribution(ctx, &flipt.CreateDistributionRequest{
						FlagKey:      f.Key,
						RuleId:       rule.Id,
						VariantId:    variant.Id,
						Rollout:      d.Rollout,
						NamespaceKey: namespace,
					})

					if err != nil {
						return fmt.Errorf("creating distribution: %w", err)
					}
				}
			}

			// support explicitly setting flag type from 1.1
			if len(f.Rollouts) > 0 {
				if err := ensureFieldSupported("flag.rollouts", semver.Version{
					Major: 1,
					Minor: 1,
				}, version); err != nil {
					return err
				}

				for idx, r := range f.Rollouts {
					if r.Segment != nil && r.Threshold != nil {
						return fmt.Errorf(`rollout "%s/%s/%d" cannot have both segment and percentage rule`,
							namespace,
							f.Key,
							idx,
						)
					}

					req := &flipt.CreateRolloutRequest{
						NamespaceKey: namespace,
						FlagKey:      f.Key,
						Description:  r.Description,
						Rank:         int32(idx + 1),
					}

					if r.Segment != nil {
						frs := &flipt.RolloutSegment{
							Value:      r.Segment.Value,
							SegmentKey: r.Segment.Key,
						}

						if len(r.Segment.Keys) > 0 && r.Segment.Key != "" {
							return fmt.Errorf("rollout %s/%s/%d cannot have both segment.keys and segment.key",
								namespace,
								f.Key,
								idx,
							)
						}

						// support explicitly setting only "keys" on rules from 1.2
						if len(r.Segment.Keys) > 0 {
							if err := ensureFieldSupported("flag.rollouts[*].segment.keys", semver.Version{
								Major: 1,
								Minor: 2,
							}, version); err != nil {
								return err
							}

							frs.SegmentKeys = r.Segment.Keys
						}

						frs.SegmentOperator = flipt.SegmentOperator(flipt.SegmentOperator_value[r.Segment.Operator])

						req.Rule = &flipt.CreateRolloutRequest_Segment{
							Segment: frs,
						}
					} else if r.Threshold != nil {
						req.Rule = &flipt.CreateRolloutRequest_Threshold{
							Threshold: &flipt.RolloutThreshold{
								Percentage: r.Threshold.Percentage,
								Value:      r.Threshold.Value,
							},
						}
					}

					if _, err := i.creator.CreateRollout(ctx, req); err != nil {
						return fmt.Errorf("creating rollout: %w", err)
					}
				}
			}
		}

		idx += 1
	}

	return nil
}

// convert converts each encountered map[interface{}]interface{} to a map[string]interface{} value.
// This is necessary because the json library does not support map[interface{}]interface{} values which nested
// maps get unmarshalled into from the yaml library.
func convert(i interface{}) interface{} {
	switch x := i.(type) {
	case map[interface{}]interface{}:
		m := map[string]interface{}{}
		for k, v := range x {
			if sk, ok := k.(string); ok {
				m[sk] = convert(v)
			}
		}
		return m
	case []interface{}:
		for i, v := range x {
			x[i] = convert(v)
		}
	}
	return i
}

func ensureFieldSupported(field string, expected, have semver.Version) error {
	if have.LT(expected) {
		return fmt.Errorf("%s is supported in version >=%s, found %s",
			field,
			versionString(expected),
			versionString(have))
	}

	return nil
}
