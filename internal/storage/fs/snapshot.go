package fs

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/google/uuid"
	"go.flipt.io/flipt/core/validation"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/core"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

const (
	defaultNs = "default"
)

var _ storage.ReadOnlyStore = (*Snapshot)(nil)

// Snapshot contains the structures necessary for serving
// flag state to a client.
type Snapshot struct {
	ns        map[string]*namespace
	evalDists map[string][]*storage.EvaluationDistribution
	now       *timestamppb.Timestamp
}

type namespace struct {
	flags        map[string]*core.Flag
	segments     map[string]*core.Segment
	evalRules    map[string][]*storage.EvaluationRule
	evalRollouts map[string][]*storage.EvaluationRollout
	etag         string
}

func newNamespace(ns *ext.NamespaceEmbed, created *timestamppb.Timestamp) *namespace {
	return &namespace{
		flags:        map[string]*core.Flag{},
		segments:     map[string]*core.Segment{},
		evalRules:    map[string][]*storage.EvaluationRule{},
		evalRollouts: map[string][]*storage.EvaluationRollout{},
	}
}

type SnapshotOption struct {
	validatorOption []validation.FeaturesValidatorOption
	etagFn          EtagFn
}

// EtagFn is a function type that takes an fs.FileInfo object as input and
// returns a string representing the ETag.
type EtagFn func(stat fs.FileInfo) string

// EtagInfo is an interface that defines a single method, Etag(), which returns
// a string representing the ETag of an object.
type EtagInfo interface {
	// Etag returns the ETag of the implementing object.
	Etag() string
}

func WithValidatorOption(opts ...validation.FeaturesValidatorOption) containers.Option[SnapshotOption] {
	return func(so *SnapshotOption) {
		so.validatorOption = opts
	}
}

// WithEtag returns a containers.Option[SnapshotOption] that sets the ETag function
// to always return the provided ETag string.
func WithEtag(etag string) containers.Option[SnapshotOption] {
	return func(so *SnapshotOption) {
		so.etagFn = func(stat fs.FileInfo) string { return etag }
	}
}

// WithFileInfoEtag returns a containers.Option[SnapshotOption] that sets the ETag function
// to generate an ETag based on the file information. If the file information implements
// the EtagInfo interface, the Etag method is used. Otherwise, it generates an ETag
// based on the file's modification time and size.
func WithFileInfoEtag() containers.Option[SnapshotOption] {
	return func(so *SnapshotOption) {
		so.etagFn = func(stat fs.FileInfo) string {
			if s, ok := stat.(EtagInfo); ok {
				return s.Etag()
			}
			return fmt.Sprintf("%x-%x", stat.ModTime().Unix(), stat.Size())
		}
	}
}

// SnapshotFromFS is a convenience function for building a snapshot
// directly from an implementation of fs.FS using the list state files
// function to source the relevant Flipt configuration files.
func SnapshotFromFS(logger *zap.Logger, src fs.FS, opts ...containers.Option[SnapshotOption]) (*Snapshot, error) {
	paths, err := listStateFiles(logger, src)
	if err != nil {
		return nil, err
	}

	return SnapshotFromPaths(logger, src, paths, opts...)
}

// SnapshotFromPaths constructs a StoreSnapshot from the provided
// slice of paths resolved against the provided fs.FS.
func SnapshotFromPaths(logger *zap.Logger, ffs fs.FS, paths []string, opts ...containers.Option[SnapshotOption]) (*Snapshot, error) {
	logger.Debug("opening state files", zap.Strings("paths", paths))

	var files []fs.File
	for _, file := range paths {
		fi, err := ffs.Open(file)
		if err != nil {
			return nil, err
		}

		files = append(files, fi)
	}

	return SnapshotFromFiles(logger, files, opts...)
}

// SnapshotFromFiles constructs a StoreSnapshot from the provided slice
// of fs.File implementations.
func SnapshotFromFiles(logger *zap.Logger, files []fs.File, opts ...containers.Option[SnapshotOption]) (*Snapshot, error) {
	s := newSnapshot()

	var so SnapshotOption
	containers.ApplyAll(&so, WithFileInfoEtag())
	containers.ApplyAll(&so, opts...)

	for _, fi := range files {
		defer fi.Close()
		info, err := fi.Stat()
		if err != nil {
			return nil, err
		}

		logger.Debug("opening state file", zap.String("path", info.Name()))

		docs, err := documentsFromFile(fi, so)
		if err != nil {
			return nil, err
		}

		for _, doc := range docs {
			if err := s.addDoc(doc); err != nil {
				return nil, err
			}
		}
	}

	return s, nil
}

// WalkDocuments walks all the Flipt feature documents found in the target fs.FS
// based on either the default index file or an index file located in the root
func WalkDocuments(logger *zap.Logger, src fs.FS, fn func(*ext.Document) error) error {
	paths, err := listStateFiles(logger, src)
	if err != nil {
		return err
	}

	for _, file := range paths {
		logger.Debug("opening state file", zap.String("path", file))

		fi, err := src.Open(file)
		if err != nil {
			return err
		}
		defer fi.Close()

		var so SnapshotOption
		containers.ApplyAll(&so, WithFileInfoEtag())
		docs, err := documentsFromFile(fi, so)
		if err != nil {
			return err
		}

		for _, doc := range docs {
			if err := fn(doc); err != nil {
				return err
			}
		}
	}

	return nil
}

func newSnapshot() *Snapshot {
	now := flipt.Now()
	return &Snapshot{
		ns: map[string]*namespace{
			defaultNs: newNamespace(ext.DefaultNamespace, now),
		},
		evalDists: map[string][]*storage.EvaluationDistribution{},
		now:       now,
	}
}

// documentsFromFile parses and validates a document from a single fs.File instance
func documentsFromFile(fi fs.File, opts SnapshotOption) ([]*ext.Document, error) {
	validator, err := validation.NewFeaturesValidator(opts.validatorOption...)
	if err != nil {
		return nil, err
	}

	stat, err := fi.Stat()
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	reader := io.TeeReader(fi, buf)

	var docs []*ext.Document
	extn := filepath.Ext(stat.Name())

	var decode func(any) error
	switch extn {
	case ".yaml", ".yml":
		// Support YAML stream by looping until we reach an EOF.
		decode = yaml.NewDecoder(buf).Decode
	case "", ".json":
		decode = json.NewDecoder(buf).Decode
	default:
		return nil, fmt.Errorf("unexpected extension: %q", extn)
	}

	// validate after we have checked supported
	// extensions but before we attempt to decode the
	// buffers contents to ensure we fill the buffer
	// via the TeeReader
	if err := validator.Validate(stat.Name(), reader); err != nil {
		return nil, err
	}

	for {
		doc := &ext.Document{}
		if err := decode(doc); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, err
		}

		// set namespace to default if empty in document
		if doc.Namespace == nil {
			doc.Namespace = ext.DefaultNamespace
		}

		doc.Etag = opts.etagFn(stat)
		docs = append(docs, doc)
	}

	return docs, nil
}

// listStateFiles lists all the file paths in a provided fs.FS containing Flipt feature state
func listStateFiles(logger *zap.Logger, src fs.FS) ([]string, error) {
	idx, err := OpenFliptIndex(logger, src)
	if err != nil {
		return nil, err
	}

	var paths []string
	if err := fs.WalkDir(src, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		if idx.Match(path) {
			paths = append(paths, path)
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return paths, nil
}

func (ss *Snapshot) addDoc(doc *ext.Document) error {
	var (
		namespaceKey = doc.Namespace.GetKey()
		ns           = ss.ns[namespaceKey]
	)

	if ns == nil {
		ns = newNamespace(doc.Namespace, ss.now)
	}

	evalDists := map[string][]*storage.EvaluationDistribution{}
	if len(ss.evalDists) > 0 {
		evalDists = ss.evalDists
	}

	for _, s := range doc.Segments {
		matchType := core.MatchType_value[s.MatchType]
		segment := &core.Segment{
			Name:        s.Name,
			Key:         s.Key,
			Description: s.Description,
			MatchType:   core.MatchType(matchType),
		}

		for _, constraint := range s.Constraints {
			constraintType := core.ComparisonType_value[constraint.Type]
			segment.Constraints = append(segment.Constraints, &core.Constraint{
				Operator:    constraint.Operator,
				Property:    constraint.Property,
				Type:        core.ComparisonType(constraintType),
				Value:       constraint.Value,
				Description: constraint.Description,
			})
		}

		ns.segments[segment.Key] = segment
	}

	for _, f := range doc.Flags {
		flagType := core.FlagType_value[f.Type]
		flag := &core.Flag{
			Key:         f.Key,
			Name:        f.Name,
			Description: f.Description,
			Enabled:     f.Enabled,
			Type:        core.FlagType(flagType),
		}

		for _, v := range f.Variants {
			attachment, err := structpb.NewValue(v.Attachment)
			if err != nil {
				return err
			}

			variant := &core.Variant{
				Key:         v.Key,
				Name:        v.Name,
				Description: v.Description,
				Attachment:  attachment,
			}

			flag.Variants = append(flag.Variants, variant)

			if v.Default {
				flag.DefaultVariant = &v.Key
			}
		}

		ns.flags[f.Key] = flag

		evalRules := []*storage.EvaluationRule{}
		for i, r := range f.Rules {
			rank := int32(i + 1)
			rule := &core.Rule{}

			evalRule := &storage.EvaluationRule{
				ID:           uuid.NewString(),
				NamespaceKey: namespaceKey,
				FlagKey:      f.Key,
				Rank:         rank,
			}

			switch s := r.Segment.IsSegment.(type) {
			case ext.SegmentKey:
				rule.Segments = []string{string(s)}
			case *ext.Segments:
				rule.Segments = s.Keys
				segmentOperator := core.SegmentOperator_value[s.SegmentOperator]

				rule.SegmentOperator = core.SegmentOperator(segmentOperator)
			}

			var (
				segmentKeys = []string{}
				segments    = make(map[string]*storage.EvaluationSegment)
			)

			if len(rule.Segments) > 0 {
				segmentKeys = append(segmentKeys, rule.Segments...)
			}

			for _, segmentKey := range segmentKeys {
				segment := ns.segments[segmentKey]
				if segment == nil {
					return errs.ErrInvalidf("flag %s/%s rule %d references unknown segment %q", doc.Namespace, flag.Key, rank, segmentKey)
				}

				evc := make([]storage.EvaluationConstraint, 0, len(segment.Constraints))
				for _, constraint := range segment.Constraints {
					evc = append(evc, storage.EvaluationConstraint{
						Operator: constraint.Operator,
						Property: constraint.Property,
						Type:     constraint.Type,
						Value:    constraint.Value,
					})
				}

				segments[segmentKey] = &storage.EvaluationSegment{
					SegmentKey:  segmentKey,
					MatchType:   segment.MatchType,
					Constraints: evc,
				}
			}

			if rule.SegmentOperator == core.SegmentOperator_AND_SEGMENT_OPERATOR {
				evalRule.SegmentOperator = core.SegmentOperator_AND_SEGMENT_OPERATOR
			}

			evalRule.Segments = segments

			evalRules = append(evalRules, evalRule)

			for _, d := range r.Distributions {
				variant, found := findByKey(d.VariantKey, flag.Variants...)
				if !found {
					return errs.ErrInvalidf("flag %s/%s rule %d references unknown variant %q", doc.Namespace, flag.Key, rank, d.VariantKey)
				}

				id := uuid.NewString()
				rule.Distributions = append(rule.Distributions, &core.Distribution{
					Rollout: d.Rollout,
					Variant: variant.Key,
				})

				attachment, err := variant.Attachment.MarshalJSON()
				if err != nil {
					return err
				}

				evalDists[evalRule.ID] = append(evalDists[evalRule.ID], &storage.EvaluationDistribution{
					ID:                id,
					Rollout:           d.Rollout,
					VariantKey:        variant.Key,
					VariantAttachment: string(attachment),
				})
			}

			flag.Rules = append(flag.Rules, rule)
		}

		ns.evalRules[f.Key] = evalRules

		evalRollouts := make([]*storage.EvaluationRollout, 0, len(f.Rollouts))
		for i, rollout := range f.Rollouts {
			rank := int32(i + 1)
			s := &storage.EvaluationRollout{
				NamespaceKey: namespaceKey,
				Rank:         rank,
			}

			flagRollout := &core.Rollout{}

			if rollout.Threshold != nil {
				s.Threshold = &storage.RolloutThreshold{
					Percentage: rollout.Threshold.Percentage,
					Value:      rollout.Threshold.Value,
				}
				s.RolloutType = core.RolloutType_THRESHOLD_ROLLOUT_TYPE

				flagRollout.Type = s.RolloutType
				flagRollout.Rule = &core.Rollout_Threshold{
					Threshold: &core.RolloutThreshold{
						Percentage: rollout.Threshold.Percentage,
						Value:      rollout.Threshold.Value,
					},
				}
			} else if rollout.Segment != nil {
				var (
					segmentKeys = []string{}
					segments    = make(map[string]*storage.EvaluationSegment)
				)

				if rollout.Segment.Key != "" {
					segmentKeys = append(segmentKeys, rollout.Segment.Key)
				} else if len(rollout.Segment.Keys) > 0 {
					segmentKeys = append(segmentKeys, rollout.Segment.Keys...)
				}

				for _, segmentKey := range segmentKeys {
					segment, ok := ns.segments[segmentKey]
					if !ok {
						return errs.ErrInvalidf("flag %s/%s rule %d references unknown segment %q", doc.Namespace, flag.Key, rank, segmentKey)
					}

					constraints := make([]storage.EvaluationConstraint, 0, len(segment.Constraints))
					for _, c := range segment.Constraints {
						constraints = append(constraints, storage.EvaluationConstraint{
							Operator: c.Operator,
							Property: c.Property,
							Type:     c.Type,
							Value:    c.Value,
						})
					}

					segments[segmentKey] = &storage.EvaluationSegment{
						SegmentKey:  segmentKey,
						MatchType:   segment.MatchType,
						Constraints: constraints,
					}
				}

				segmentOperator := core.SegmentOperator_value[rollout.Segment.Operator]

				s.Segment = &storage.RolloutSegment{
					Segments:        segments,
					SegmentOperator: core.SegmentOperator(segmentOperator),
					Value:           rollout.Segment.Value,
				}

				s.RolloutType = core.RolloutType_SEGMENT_ROLLOUT_TYPE

				frs := &core.RolloutSegment{
					Value:           rollout.Segment.Value,
					SegmentOperator: core.SegmentOperator(segmentOperator),
					Segments:        segmentKeys,
				}

				flagRollout.Type = s.RolloutType
				flagRollout.Rule = &core.Rollout_Segment{
					Segment: frs,
				}
			}

			flag.Rollouts = append(flag.Rollouts, flagRollout)

			evalRollouts = append(evalRollouts, s)
		}

		ns.evalRollouts[f.Key] = evalRollouts
	}

	ns.etag = doc.Etag
	ss.ns[namespaceKey] = ns

	ss.evalDists = evalDists

	return nil
}

func (ss Snapshot) String() string {
	return "snapshot"
}

func (ss *Snapshot) GetFlag(ctx context.Context, req storage.ResourceRequest) (*core.Flag, error) {
	ns, err := ss.getNamespace(req.Namespace())
	if err != nil {
		return nil, err
	}

	flag, ok := ns.flags[req.Key]
	if !ok {
		return nil, errs.ErrNotFoundf("flag %q", req)
	}

	return flag, nil
}

func (ss *Snapshot) ListFlags(ctx context.Context, req *storage.ListRequest[storage.NamespaceRequest]) (set storage.ResultSet[*core.Flag], err error) {
	ns, err := ss.getNamespace(req.Predicate.Namespace())
	if err != nil {
		return set, err
	}

	flags := make([]*core.Flag, 0, len(ns.flags))
	for _, flag := range ns.flags {
		flags = append(flags, flag)
	}

	return paginate(req.QueryParams, func(i, j int) bool {
		return flags[i].Key < flags[j].Key
	}, flags...)
}

func (ss *Snapshot) CountFlags(ctx context.Context, p storage.NamespaceRequest) (uint64, error) {
	ns, err := ss.getNamespace(p.Namespace())
	if err != nil {
		return 0, err
	}

	return uint64(len(ns.flags)), nil
}

func (ss *Snapshot) GetEvaluationRules(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRule, error) {
	ns, ok := ss.ns[flag.Namespace()]
	if !ok {
		return nil, errs.ErrNotFoundf("namespace %q", flag.NamespaceRequest)
	}

	rules, ok := ns.evalRules[flag.Key]
	if !ok {
		return nil, errs.ErrNotFoundf("flag %q", flag)
	}

	return rules, nil
}

func (ss *Snapshot) GetEvaluationDistributions(ctx context.Context, flag storage.ResourceRequest, rule storage.IDRequest) ([]*storage.EvaluationDistribution, error) {
	dists, ok := ss.evalDists[rule.ID]
	if !ok {
		return []*storage.EvaluationDistribution{}, nil
	}

	return dists, nil
}

func (ss *Snapshot) GetEvaluationRollouts(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRollout, error) {
	ns, ok := ss.ns[flag.Namespace()]
	if !ok {
		return nil, errs.ErrNotFoundf("namespace %q", flag.NamespaceRequest)
	}

	rollouts, ok := ns.evalRollouts[flag.Key]
	if !ok {
		return nil, errs.ErrNotFoundf("flag %q", flag)
	}

	return rollouts, nil
}

func findByKey[T interface{ GetKey() string }](key string, ts ...T) (t T, _ bool) {
	return find(func(t T) bool { return t.GetKey() == key }, ts...)
}

func find[T any](match func(T) bool, ts ...T) (t T, _ bool) {
	for _, t := range ts {
		if match(t) {
			return t, true
		}
	}

	return t, false
}

func paginate[T any](params storage.QueryParams, less func(i, j int) bool, items ...T) (storage.ResultSet[T], error) {
	if len(items) == 0 {
		return storage.ResultSet[T]{}, nil
	}

	set := storage.ResultSet[T]{
		Results: items,
	}

	// sort by created_at and specified order
	sort.Slice(set.Results, func(i, j int) bool {
		if params.Order != storage.OrderAsc {
			i, j = j, i
		}

		return less(i, j)
	})

	// parse page token as an offset integer
	var offset int
	v, err := strconv.ParseInt(params.PageToken, 10, 64)
	if params.PageToken != "" && err != nil {
		return storage.ResultSet[T]{}, errs.ErrInvalidf("pageToken is not valid: %q", params.PageToken)
	}

	offset = int(v)

	if offset >= len(set.Results) {
		return storage.ResultSet[T]{}, errs.ErrInvalidf("invalid offset: %d", offset)
	}

	// 0 means no limit on page size (all items from offset)
	if params.Limit == 0 {
		set.Results = set.Results[offset:]
		return set, nil
	}

	// ensure end of page does not exceed entire set
	end := offset + int(params.Limit)
	if end > len(set.Results) {
		end = len(set.Results)
	} else if end < len(set.Results) {
		// set next page token given there are more entries
		set.NextPageToken = fmt.Sprintf("%d", end)
	}

	// reduce results set to requested page
	set.Results = set.Results[offset:end]

	return set, nil
}

func (ss *Snapshot) getNamespace(key string) (namespace, error) {
	ns, ok := ss.ns[key]
	if !ok {
		return namespace{}, errs.ErrNotFoundf("namespace %q", key)
	}

	return *ns, nil
}

func (ss *Snapshot) GetVersion(ctx context.Context, req storage.NamespaceRequest) (string, error) {
	ns, err := ss.getNamespace(req.Namespace())
	if err != nil {
		return "", err
	}
	return ns.etag, nil
}
