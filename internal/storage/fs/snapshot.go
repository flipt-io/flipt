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

	"github.com/gofrs/uuid"
	"go.flipt.io/flipt/core/validation"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
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
	resource     *flipt.Namespace
	flags        map[string]*flipt.Flag
	segments     map[string]*flipt.Segment
	rules        map[string]*flipt.Rule
	rollouts     map[string]*flipt.Rollout
	evalRules    map[string][]*storage.EvaluationRule
	evalRollouts map[string][]*storage.EvaluationRollout
	etag         string
}

func newNamespace(ns *ext.NamespaceEmbed, created *timestamppb.Timestamp) *namespace {
	var (
		namespaceName        string
		namespaceDescription string
	)

	switch t := ns.IsNamespace.(type) {
	case ext.NamespaceKey:
		namespaceName = string(t)
	case *ext.Namespace:
		namespaceName = t.Name
		namespaceDescription = t.Description
	}
	return &namespace{
		resource: &flipt.Namespace{
			Key:         ns.GetKey(),
			Name:        namespaceName,
			Description: namespaceDescription,
			CreatedAt:   created,
			UpdatedAt:   created,
		},
		flags:        map[string]*flipt.Flag{},
		segments:     map[string]*flipt.Segment{},
		rules:        map[string]*flipt.Rule{},
		rollouts:     map[string]*flipt.Rollout{},
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
		matchType := flipt.MatchType_value[s.MatchType]
		segment := &flipt.Segment{
			NamespaceKey: namespaceKey,
			Name:         s.Name,
			Key:          s.Key,
			Description:  s.Description,
			MatchType:    flipt.MatchType(matchType),
			CreatedAt:    ss.now,
			UpdatedAt:    ss.now,
		}

		for _, constraint := range s.Constraints {
			constraintType := flipt.ComparisonType_value[constraint.Type]
			segment.Constraints = append(segment.Constraints, &flipt.Constraint{
				NamespaceKey: namespaceKey,
				SegmentKey:   segment.Key,
				Id:           uuid.Must(uuid.NewV4()).String(),
				Operator:     constraint.Operator,
				Property:     constraint.Property,
				Type:         flipt.ComparisonType(constraintType),
				Value:        constraint.Value,
				Description:  constraint.Description,
				CreatedAt:    ss.now,
				UpdatedAt:    ss.now,
			})
		}

		ns.segments[segment.Key] = segment
	}

	for _, f := range doc.Flags {
		flagType := flipt.FlagType_value[f.Type]
		flag := &flipt.Flag{
			NamespaceKey: namespaceKey,
			Key:          f.Key,
			Name:         f.Name,
			Description:  f.Description,
			Enabled:      f.Enabled,
			Type:         flipt.FlagType(flagType),
			CreatedAt:    ss.now,
			UpdatedAt:    ss.now,
		}

		for _, v := range f.Variants {
			attachment, err := json.Marshal(v.Attachment)
			if err != nil {
				return err
			}

			variant := &flipt.Variant{
				Id:           uuid.Must(uuid.NewV4()).String(),
				NamespaceKey: namespaceKey,
				Key:          v.Key,
				Name:         v.Name,
				Description:  v.Description,
				Attachment:   string(attachment),
				CreatedAt:    ss.now,
				UpdatedAt:    ss.now,
			}

			flag.Variants = append(flag.Variants, variant)

			if v.Default {
				flag.DefaultVariant = variant
			}
		}

		ns.flags[f.Key] = flag

		evalRules := []*storage.EvaluationRule{}
		for i, r := range f.Rules {
			rank := int32(i + 1)
			rule := &flipt.Rule{
				NamespaceKey: namespaceKey,
				Id:           uuid.Must(uuid.NewV4()).String(),
				FlagKey:      f.Key,
				Rank:         rank,
				CreatedAt:    ss.now,
				UpdatedAt:    ss.now,
			}

			evalRule := &storage.EvaluationRule{
				NamespaceKey: namespaceKey,
				FlagKey:      f.Key,
				ID:           rule.Id,
				Rank:         rank,
			}

			switch s := r.Segment.IsSegment.(type) {
			case ext.SegmentKey:
				rule.SegmentKey = string(s)
			case *ext.Segments:
				rule.SegmentKeys = s.Keys
				segmentOperator := flipt.SegmentOperator_value[s.SegmentOperator]

				rule.SegmentOperator = flipt.SegmentOperator(segmentOperator)
			}

			var (
				segmentKeys = []string{}
				segments    = make(map[string]*storage.EvaluationSegment)
			)

			if rule.SegmentKey != "" {
				segmentKeys = append(segmentKeys, rule.SegmentKey)
			} else if len(rule.SegmentKeys) > 0 {
				segmentKeys = append(segmentKeys, rule.SegmentKeys...)
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

			if rule.SegmentOperator == flipt.SegmentOperator_AND_SEGMENT_OPERATOR {
				evalRule.SegmentOperator = flipt.SegmentOperator_AND_SEGMENT_OPERATOR
			}

			evalRule.Segments = segments

			evalRules = append(evalRules, evalRule)

			for _, d := range r.Distributions {
				variant, found := findByKey(d.VariantKey, flag.Variants...)
				if !found {
					return errs.ErrInvalidf("flag %s/%s rule %d references unknown variant %q", doc.Namespace, flag.Key, rank, d.VariantKey)
				}

				id := uuid.Must(uuid.NewV4()).String()
				rule.Distributions = append(rule.Distributions, &flipt.Distribution{
					Id:        id,
					Rollout:   d.Rollout,
					RuleId:    rule.Id,
					VariantId: variant.Id,
					CreatedAt: ss.now,
					UpdatedAt: ss.now,
				})

				evalDists[evalRule.ID] = append(evalDists[evalRule.ID], &storage.EvaluationDistribution{
					ID:                id,
					Rollout:           d.Rollout,
					VariantID:         variant.Id,
					VariantKey:        variant.Key,
					VariantAttachment: variant.Attachment,
				})
			}

			ns.rules[rule.Id] = rule
		}

		ns.evalRules[f.Key] = evalRules

		evalRollouts := make([]*storage.EvaluationRollout, 0, len(f.Rollouts))
		for i, rollout := range f.Rollouts {
			rank := int32(i + 1)
			s := &storage.EvaluationRollout{
				NamespaceKey: namespaceKey,
				Rank:         rank,
			}

			flagRollout := &flipt.Rollout{
				Id:           uuid.Must(uuid.NewV4()).String(),
				Rank:         rank,
				FlagKey:      f.Key,
				NamespaceKey: namespaceKey,
				CreatedAt:    ss.now,
				UpdatedAt:    ss.now,
			}

			if rollout.Threshold != nil {
				s.Threshold = &storage.RolloutThreshold{
					Percentage: rollout.Threshold.Percentage,
					Value:      rollout.Threshold.Value,
				}
				s.RolloutType = flipt.RolloutType_THRESHOLD_ROLLOUT_TYPE

				flagRollout.Type = s.RolloutType
				flagRollout.Rule = &flipt.Rollout_Threshold{
					Threshold: &flipt.RolloutThreshold{
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

				segmentOperator := flipt.SegmentOperator_value[rollout.Segment.Operator]

				s.Segment = &storage.RolloutSegment{
					Segments:        segments,
					SegmentOperator: flipt.SegmentOperator(segmentOperator),
					Value:           rollout.Segment.Value,
				}

				s.RolloutType = flipt.RolloutType_SEGMENT_ROLLOUT_TYPE

				frs := &flipt.RolloutSegment{
					Value:           rollout.Segment.Value,
					SegmentOperator: flipt.SegmentOperator(segmentOperator),
				}

				if len(segmentKeys) == 1 {
					frs.SegmentKey = segmentKeys[0]
				} else {
					frs.SegmentKeys = segmentKeys
				}

				flagRollout.Type = s.RolloutType
				flagRollout.Rule = &flipt.Rollout_Segment{
					Segment: frs,
				}
			}

			ns.rollouts[flagRollout.Id] = flagRollout

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

func (ss *Snapshot) GetRule(ctx context.Context, p storage.NamespaceRequest, id string) (rule *flipt.Rule, _ error) {
	ns, err := ss.getNamespace(p.Namespace())
	if err != nil {
		return nil, err
	}

	var ok bool
	rule, ok = ns.rules[id]
	if !ok {
		return nil, errs.ErrNotFoundf(`rule "%s/%s"`, p.Namespace(), id)
	}

	return rule, nil
}

func (ss *Snapshot) ListRules(ctx context.Context, req *storage.ListRequest[storage.ResourceRequest]) (set storage.ResultSet[*flipt.Rule], _ error) {
	ns, err := ss.getNamespace(req.Predicate.Namespace())
	if err != nil {
		return set, err
	}

	rules := make([]*flipt.Rule, 0, len(ns.rules))
	for _, rule := range ns.rules {
		if rule.FlagKey == req.Predicate.Key {
			rules = append(rules, rule)
		}
	}

	return paginate(req.QueryParams, func(i, j int) bool {
		return rules[i].Rank < rules[j].Rank
	}, rules...)
}

func (ss *Snapshot) CountRules(ctx context.Context, req storage.ResourceRequest) (uint64, error) {
	ns, err := ss.getNamespace(req.Namespace())
	if err != nil {
		return 0, err
	}

	var count uint64 = 0
	for _, rule := range ns.rules {
		if rule.FlagKey == req.Key {
			count += 1
		}
	}

	return count, nil
}

func (ss *Snapshot) GetSegment(ctx context.Context, req storage.ResourceRequest) (*flipt.Segment, error) {
	ns, err := ss.getNamespace(req.Namespace())
	if err != nil {
		return nil, err
	}

	segment, ok := ns.segments[req.Key]
	if !ok {
		return nil, errs.ErrNotFoundf("segment %q", req)
	}

	return segment, nil
}

func (ss *Snapshot) ListSegments(ctx context.Context, req *storage.ListRequest[storage.NamespaceRequest]) (set storage.ResultSet[*flipt.Segment], err error) {
	ns, err := ss.getNamespace(req.Predicate.Namespace())
	if err != nil {
		return set, err
	}

	segments := make([]*flipt.Segment, 0, len(ns.segments))
	for _, segment := range ns.segments {
		segments = append(segments, segment)
	}

	return paginate(req.QueryParams, func(i, j int) bool {
		return segments[i].Key < segments[j].Key
	}, segments...)
}

func (ss *Snapshot) CountSegments(ctx context.Context, p storage.NamespaceRequest) (uint64, error) {
	ns, err := ss.getNamespace(p.Namespace())
	if err != nil {
		return 0, err
	}

	return uint64(len(ns.segments)), nil
}

func (ss *Snapshot) GetNamespace(ctx context.Context, p storage.NamespaceRequest) (*flipt.Namespace, error) {
	ns, err := ss.getNamespace(p.Namespace())
	if err != nil {
		return nil, err
	}

	return ns.resource, nil
}

func (ss *Snapshot) ListNamespaces(ctx context.Context, req *storage.ListRequest[storage.ReferenceRequest]) (set storage.ResultSet[*flipt.Namespace], err error) {
	ns := make([]*flipt.Namespace, 0, len(ss.ns))
	for _, n := range ss.ns {
		ns = append(ns, n.resource)
	}

	return paginate(req.QueryParams, func(i, j int) bool {
		return ns[i].Key < ns[j].Key
	}, ns...)
}

func (ss *Snapshot) CountNamespaces(ctx context.Context, _ storage.ReferenceRequest) (uint64, error) {
	return uint64(len(ss.ns)), nil
}

func (ss *Snapshot) GetFlag(ctx context.Context, req storage.ResourceRequest) (*flipt.Flag, error) {
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

func (ss *Snapshot) ListFlags(ctx context.Context, req *storage.ListRequest[storage.NamespaceRequest]) (set storage.ResultSet[*flipt.Flag], err error) {
	ns, err := ss.getNamespace(req.Predicate.Namespace())
	if err != nil {
		return set, err
	}

	flags := make([]*flipt.Flag, 0, len(ns.flags))
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

func (ss *Snapshot) GetEvaluationDistributions(ctx context.Context, rule storage.IDRequest) ([]*storage.EvaluationDistribution, error) {
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

func (ss *Snapshot) GetRollout(ctx context.Context, p storage.NamespaceRequest, id string) (*flipt.Rollout, error) {
	ns, err := ss.getNamespace(p.Namespace())
	if err != nil {
		return nil, err
	}

	rollout, ok := ns.rollouts[id]
	if !ok {
		return nil, errs.ErrNotFoundf(`rollout "%s/%s"`, p.Namespace(), id)
	}

	return rollout, nil
}

func (ss *Snapshot) ListRollouts(ctx context.Context, req *storage.ListRequest[storage.ResourceRequest]) (set storage.ResultSet[*flipt.Rollout], err error) {
	ns, err := ss.getNamespace(req.Predicate.Namespace())
	if err != nil {
		return set, err
	}

	rollouts := make([]*flipt.Rollout, 0)
	for _, rollout := range ns.rollouts {
		if rollout.FlagKey == req.Predicate.Key {
			rollouts = append(rollouts, rollout)
		}
	}

	return paginate(req.QueryParams, func(i, j int) bool {
		return rollouts[i].Rank < rollouts[j].Rank
	}, rollouts...)
}

func (ss *Snapshot) CountRollouts(ctx context.Context, flag storage.ResourceRequest) (uint64, error) {
	ns, err := ss.getNamespace(flag.Namespace())
	if err != nil {
		return 0, err
	}

	var count uint64 = 0
	for _, rollout := range ns.rollouts {
		if rollout.FlagKey == flag.Key {
			count += 1
		}
	}

	return count, nil
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
