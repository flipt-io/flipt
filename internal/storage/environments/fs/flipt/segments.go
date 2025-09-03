package flipt

import (
	"context"
	"fmt"
	"os"
	"path"

	"slices"

	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/ext"
	serverenvironments "go.flipt.io/flipt/internal/server/environments"
	environmentsfs "go.flipt.io/flipt/internal/storage/environments/fs"
	"go.flipt.io/flipt/rpc/flipt/core"
	rpcenvironments "go.flipt.io/flipt/rpc/v2/environments"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/anypb"
)

var _ environmentsfs.ResourceStorage = (*SegmentStorage)(nil)

// SegmentStorage implements the configuration storage ResourceTypeStorage
// and handles retrieving and storing Segment types from Flipt features.yaml
// declarative format through an opinionated convention for flag state layout
type SegmentStorage struct {
	logger *zap.Logger
}

// NewSegmentStorage constructs and configures a new segment storage implementation
func NewSegmentStorage(logger *zap.Logger) *SegmentStorage {
	return &SegmentStorage{logger: logger}
}

// ResourceType returns the Flag specific resource type
func (SegmentStorage) ResourceType() serverenvironments.ResourceType {
	return serverenvironments.NewResourceType("flipt.core", "Segment")
}

// GetResource fetches the requested segment from the namespaces features.yaml file
func (f *SegmentStorage) GetResource(ctx context.Context, fs environmentsfs.Filesystem, namespace string, key string) (_ *rpcenvironments.Resource, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("getting segment %s/%s: %w", namespace, key, err)
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

		for _, s := range doc.Segments {
			if s.Key == key {
				payload, err := payloadFromSegment(s)
				if err != nil {
					return nil, err
				}

				return &rpcenvironments.Resource{
					NamespaceKey: namespace,
					Key:          s.Key,
					Payload:      payload,
				}, nil
			}
		}
	}

	return nil, errors.ErrNotFoundf("segment %s/%s", namespace, key)
}

// ListResources fetches all segments within the namespaces features.yaml file
func (f *SegmentStorage) ListResources(ctx context.Context, fs environmentsfs.Filesystem, namespace string) (rs []*rpcenvironments.Resource, err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("listing segments %s: %w", namespace, err)
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

		for _, s := range doc.Segments {
			payload, err := payloadFromSegment(s)
			if err != nil {
				return nil, err
			}

			rs = append(rs, &rpcenvironments.Resource{
				NamespaceKey: namespace,
				Key:          s.Key,
				Payload:      payload,
			})
		}
	}

	return
}

// PutResource updates or inserts the requested segment from the target namespaces features.yaml file
func (f *SegmentStorage) PutResource(ctx context.Context, fs environmentsfs.Filesystem, rs *rpcenvironments.Resource) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("putting segment %s/%s: %w", rs.NamespaceKey, rs.Key, err)
		}
	}()

	segment, err := resourceToSegment(rs)
	if err != nil {
		return err
	}

	docs, idx, err := getDocsAndNamespace(ctx, fs, rs.NamespaceKey)
	if err != nil {
		return err
	}

	var found bool
	for i, s := range docs[idx].Segments {
		if found = s.Key == segment.Key; found {
			docs[idx].Segments[i] = segment
			break
		}
	}

	if !found {
		docs[idx].Segments = append(docs[idx].Segments, segment)
	}

	filename, err := environmentsfs.FindFeaturesFilename(fs, rs.NamespaceKey)
	if err != nil {
		return err
	}
	fi, err := fs.OpenFile(path.Join(rs.NamespaceKey, filename), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
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

// DeleteResource removes the requested segment from the target namespaces features.yaml file
func (f *SegmentStorage) DeleteResource(ctx context.Context, fs environmentsfs.Filesystem, namespace string, key string) (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("deleting segment %s/%s: %w", namespace, key, err)
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

		for i, s := range doc.Segments {
			if s.Key == key {
				found = true
				// remove entry from list
				doc.Segments = slices.Delete(doc.Segments, i, i+1)
				break
			}
		}
	}

	// file contents remains unchanged
	if !found {
		return nil
	}

	filename, err := environmentsfs.FindFeaturesFilename(fs, namespace)
	if err != nil {
		return err
	}
	fi, err := fs.OpenFile(path.Join(namespace, filename), os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)
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

func payloadFromSegment(segment *ext.Segment) (*anypb.Any, error) {
	dst := &core.Segment{
		Key:         segment.Key,
		MatchType:   core.MatchType(core.MatchType_value[segment.MatchType]),
		Name:        segment.Name,
		Description: segment.Description,
	}

	for _, constraint := range segment.Constraints {
		dst.Constraints = append(dst.Constraints, &core.Constraint{
			Type:        core.ComparisonType(core.ComparisonType_value[constraint.Type]),
			Description: constraint.Description,
			Property:    constraint.Property,
			Operator:    constraint.Operator,
			Value:       constraint.Value,
		})
	}

	return newAny(dst)
}

func resourceToSegment(r *rpcenvironments.Resource) (*ext.Segment, error) {
	var s core.Segment
	if err := r.Payload.UnmarshalTo(&s); err != nil {
		return nil, err
	}

	segment := &ext.Segment{
		MatchType:   s.MatchType.String(),
		Key:         r.Key,
		Name:        s.Name,
		Description: s.Description,
	}

	for _, constraint := range s.Constraints {
		c := &ext.Constraint{
			Type:        constraint.Type.String(),
			Description: constraint.Description,
			Property:    constraint.Property,
			Operator:    constraint.Operator,
			Value:       constraint.Value,
		}

		segment.Constraints = append(segment.Constraints, c)
	}

	return segment, nil
}
