package flipt

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"go.flipt.io/flipt/core/validation"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/ext"
	serverenvironments "go.flipt.io/flipt/internal/server/environments"
	environmentsfs "go.flipt.io/flipt/internal/storage/environments/fs"
	"go.flipt.io/flipt/rpc/flipt/core"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"
	"gopkg.in/yaml.v3"
	"slices"
)

var _ environmentsfs.ResourceStorage = (*FlagStorage)(nil)

// FlagStorage implements the configuration storage ResourceTypeStorage
// and handles retrieving and storing Flag types from Flipt features.yaml
// declarative format through an opinionated convention for flag state layout
type FlagStorage struct {
	logger *zap.Logger
}

// NewFlagStorage constructs and configures a new flag storage implementation
func NewFlagStorage(logger *zap.Logger) *FlagStorage {
	return &FlagStorage{logger: logger}
}

// ResourceType returns the Flag specific resource type
func (FlagStorage) ResourceType() serverenvironments.ResourceType {
	return serverenvironments.NewResourceType("flipt.core", "Flag")
}

// GetResource fetches the requested flag from the namespaces features.yaml file
func (f *FlagStorage) GetResource(ctx context.Context, fs environmentsfs.Filesystem, namespace string, key string) (_ *rpcenvironments.Resource, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("getting flag %s/%s: %w", namespace, key, err)
		}
	}()

	docs, err := parseNamespace(ctx, fs, namespace)
	if err != nil {
		return nil, err
	}

	for _, doc := range docs {
		if doc.Namespace.GetKey() != namespace {
			continue
		}

		for _, ff := range doc.Flags {
			if ff.Key == key {
				payload, err := payloadFromFlag(ff)
				if err != nil {
					return nil, err
				}

				return &rpcenvironments.Resource{
					NamespaceKey: namespace,
					Key:          ff.Key,
					Payload:      payload,
				}, nil
			}
		}
	}

	return nil, errors.ErrNotFoundf("flag %s/%s", namespace, key)
}

// ListResources fetches all flags within the namespaces features.yaml file
func (f *FlagStorage) ListResources(ctx context.Context, fs environmentsfs.Filesystem, namespace string) (rs []*rpcenvironments.Resource, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("listing flags %s: %w", namespace, err)
		}
	}()

	docs, err := parseNamespace(ctx, fs, namespace)
	if err != nil {
		return nil, err
	}

	for _, doc := range docs {
		if doc.Namespace.GetKey() != namespace {
			continue
		}

		for _, f := range doc.Flags {
			payload, err := payloadFromFlag(f)
			if err != nil {
				return nil, err
			}

			rs = append(rs, &rpcenvironments.Resource{
				NamespaceKey: namespace,
				Key:          f.Key,
				Payload:      payload,
			})
		}
	}

	return
}

// PutResource updates or inserts the requested flag from the target namespaces features.yaml file
func (f *FlagStorage) PutResource(ctx context.Context, fs environmentsfs.Filesystem, rs *rpcenvironments.Resource) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("putting flag %s/%s: %w", rs.NamespaceKey, rs.Key, err)
		}
	}()

	flag, err := resourceToFlag(rs)
	if err != nil {
		return err
	}

	docs, idx, err := getDocsAndNamespace(ctx, fs, rs.NamespaceKey)
	if err != nil {
		return err
	}

	var found bool
	for i, f := range docs[idx].Flags {
		// replace the flag if located
		if found = f.Key == flag.Key; found {
			docs[idx].Flags[i] = flag
			break
		}
	}

	// append the flag if not found
	if !found {
		docs[idx].Flags = append(docs[idx].Flags, flag)
	}

	fi, err := fs.OpenFile(path.Join(rs.NamespaceKey, "features.yaml"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fi.Close()

	enc := newDocumentEncoder(fi)
	for _, doc := range docs {
		if err := enc.Encode(doc); err != nil {
			return err
		}
	}

	return enc.Close()
}

// DeleteResource removes the requested flag from the target namespaces features.yaml file
func (f *FlagStorage) DeleteResource(ctx context.Context, fs environmentsfs.Filesystem, namespace string, key string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("deleting flag %s/%s: %w", namespace, key, err)
		}
	}()

	docs, err := parseNamespace(ctx, fs, namespace)
	if err != nil {
		return err
	}

	if len(docs) == 0 {
		return nil
	}

	var found bool
	for _, doc := range docs {
		if doc.Namespace.GetKey() != namespace {
			continue
		}

		for i, f := range doc.Flags {
			if f.Key == key {
				found = true
				// remove entry from list
				doc.Flags = slices.Delete(doc.Flags, i, i+1)
			}
		}
	}

	// file contents remains unchanged
	if !found {
		return nil
	}

	fi, err := fs.OpenFile(path.Join(namespace, "features.yaml"), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer fi.Close()

	enc := newDocumentEncoder(fi)
	for _, doc := range docs {
		if err := enc.Encode(doc); err != nil {
			return err
		}
	}

	return enc.Close()
}

type documentEncoder struct {
	*yaml.Encoder
	buf *bytes.Buffer
}

func newDocumentEncoder(wr io.Writer) documentEncoder {
	docEnc := documentEncoder{buf: &bytes.Buffer{}}
	docEnc.Encoder = yaml.NewEncoder(io.MultiWriter(wr, docEnc.buf))
	return docEnc
}

func (e documentEncoder) Close() error {
	validator, err := validation.NewFeaturesValidator()
	if err != nil {
		return err
	}

	return validator.Validate("features.yaml", e.buf)
}

func payloadFromFlag(flag *ext.Flag) (_ *anypb.Any, err error) {
	dst := &core.Flag{
		Key:         flag.Key,
		Name:        flag.Name,
		Type:        core.FlagType(core.FlagType_value[flag.Type]),
		Description: flag.Description,
		Enabled:     flag.Enabled,
	}

	if flag.Metadata != nil {
		dst.Metadata, err = structpb.NewStruct(flag.Metadata)
		if err != nil {
			return nil, err
		}
	}

	for _, variant := range flag.Variants {
		var attach *structpb.Value
		if variant.Attachment != nil {
			attach, err = structpb.NewValue(variant.Attachment)
			if err != nil {
				return nil, err
			}
		}

		dst.Variants = append(dst.Variants, &core.Variant{
			Key:         variant.Key,
			Name:        variant.Name,
			Description: variant.Description,
			Attachment:  attach,
		})

		key := variant.Key
		if variant.Default {
			dst.DefaultVariant = &key
		}
	}

	for _, rule := range flag.Rules {
		r := &core.Rule{
			Segments:        rule.Segment.Keys,
			SegmentOperator: core.SegmentOperator(core.SegmentOperator_value[rule.Segment.Operator]),
		}

		for _, dist := range rule.Distributions {
			r.Distributions = append(r.Distributions, &core.Distribution{
				Rollout: dist.Rollout,
				Variant: dist.VariantKey,
			})
		}

		dst.Rules = append(dst.Rules, r)
	}

	for _, rollout := range flag.Rollouts {
		r := &core.Rollout{
			Description: rollout.Description,
		}

		if rollout.Segment != nil {
			var (
				segmentKeys = []string{}
			)

			if len(rollout.Segment.Keys) > 0 {
				segmentKeys = append(segmentKeys, rollout.Segment.Keys...)
			}

			r.Type = core.RolloutType_SEGMENT_ROLLOUT_TYPE
			r.Rule = &core.Rollout_Segment{
				Segment: &core.RolloutSegment{
					SegmentOperator: core.SegmentOperator(core.SegmentOperator_value[rollout.Segment.Operator]),
					Segments:        segmentKeys,
					Value:           rollout.Segment.Value,
				},
			}
			dst.Rollouts = append(dst.Rollouts, r)
			continue
		}

		if rollout.Threshold != nil {
			r.Type = core.RolloutType_THRESHOLD_ROLLOUT_TYPE
			r.Rule = &core.Rollout_Threshold{
				Threshold: &core.RolloutThreshold{
					Percentage: rollout.Threshold.Percentage,
					Value:      rollout.Threshold.Value,
				},
			}
			dst.Rollouts = append(dst.Rollouts, r)
			continue
		}

		return nil, fmt.Errorf("unknown rollout type")
	}
	return newAny(dst)
}

func resourceToFlag(r *rpcenvironments.Resource) (*ext.Flag, error) {
	var f core.Flag
	if err := r.Payload.UnmarshalTo(&f); err != nil {
		return nil, err
	}

	flag := &ext.Flag{
		Type:        f.Type.String(),
		Key:         r.Key,
		Name:        f.Name,
		Description: f.Description,
		Enabled:     f.Enabled,
		Metadata:    f.Metadata.AsMap(),
	}

	for _, variant := range f.Variants {
		v := &ext.Variant{
			Default:     f.DefaultVariant != nil && *f.DefaultVariant == variant.Key,
			Key:         variant.Key,
			Name:        variant.Name,
			Description: variant.Description,
		}

		if attach := variant.Attachment.AsInterface(); attach != nil {
			v.Attachment = attach
		}

		flag.Variants = append(flag.Variants, v)
	}

	for _, rule := range f.Rules {
		r := &ext.Rule{
			Segment: &ext.SegmentRule{
				Keys:     rule.Segments,
				Operator: rule.SegmentOperator.String(),
			},
		}

		for _, dist := range rule.Distributions {
			r.Distributions = append(r.Distributions, &ext.Distribution{
				Rollout:    dist.Rollout,
				VariantKey: dist.Variant,
			})
		}

		flag.Rules = append(flag.Rules, r)
	}

	for _, rollout := range f.Rollouts {
		r := &ext.Rollout{
			Description: rollout.Description,
		}

		switch rollout.Type {
		case core.RolloutType_SEGMENT_ROLLOUT_TYPE:
			r.Segment = &ext.SegmentRule{
				Keys:     rollout.GetSegment().Segments,
				Operator: rollout.GetSegment().SegmentOperator.String(),
				Value:    rollout.GetSegment().Value,
			}
		case core.RolloutType_THRESHOLD_ROLLOUT_TYPE:
			r.Threshold = &ext.ThresholdRule{
				Percentage: rollout.GetThreshold().Percentage,
				Value:      rollout.GetThreshold().Value,
			}
		default:
			return nil, fmt.Errorf("unexpected rollout type: %q", rollout.Type)
		}

		flag.Rollouts = append(flag.Rollouts, r)
	}

	return flag, nil
}
