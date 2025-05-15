package fs

import (
	"bytes"
	"context"
	"crypto/sha1" //nolint:gosec
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"path/filepath"
	"sort"
	"strconv"

	"github.com/google/uuid"
	"go.flipt.io/flipt/core/validation"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/containers"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/internal/server/environments"
	"go.flipt.io/flipt/internal/storage"
	configcoreflipt "go.flipt.io/flipt/internal/storage/environments/fs/flipt"
	"go.flipt.io/flipt/rpc/flipt"
	"go.flipt.io/flipt/rpc/flipt/core"
	"go.flipt.io/flipt/rpc/v2/evaluation"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

var _ storage.ReadOnlyStore = (*Snapshot)(nil)

// Snapshot contains the structures necessary for serving
// flag state to a evaluation.
type Snapshot struct {
	ns        map[string]*namespace
	evalDists map[string][]*storage.EvaluationDistribution
	evalSnap  *evaluation.EvaluationSnapshot
	now       *timestamppb.Timestamp
}

type namespace struct {
	flags        map[string]*core.Flag
	segments     map[string]*core.Segment
	evalRules    map[string][]*storage.EvaluationRule
	evalRollouts map[string][]*storage.EvaluationRollout
	etag         string
}

func newNamespace() *namespace {
	return &namespace{
		flags:        map[string]*core.Flag{},
		segments:     map[string]*core.Segment{},
		evalRules:    map[string][]*storage.EvaluationRule{},
		evalRollouts: map[string][]*storage.EvaluationRollout{},
	}
}

type SnapshotOption struct {
	validatorOption []validation.FeaturesValidatorOption
}

func newSnapshotOption(opts ...containers.Option[SnapshotOption]) SnapshotOption {
	so := SnapshotOption{}
	containers.ApplyAll(&so, opts...)
	return so
}

func WithValidatorOption(opts ...validation.FeaturesValidatorOption) containers.Option[SnapshotOption] {
	return func(so *SnapshotOption) {
		so.validatorOption = opts
	}
}

type SnapshotBuilder struct {
	logger          *zap.Logger
	dependencyGraph *configcoreflipt.DependencyGraph
	opts            []containers.Option[SnapshotOption]
}

func NewSnapshotBuilder(logger *zap.Logger, dependencyGraph *configcoreflipt.DependencyGraph, opts ...containers.Option[SnapshotOption]) *SnapshotBuilder {
	return &SnapshotBuilder{
		logger:          logger,
		dependencyGraph: dependencyGraph,
		opts:            opts,
	}
}

// SnapshotFromFS is a convenience function for building a snapshot
// directly from an implementation of fs.FS using the list state files
// function to source the relevant Flipt configuration files.
func (b *SnapshotBuilder) SnapshotFromFS(conf *Config, src fs.FS) (*Snapshot, error) {
	paths, err := conf.List(src)
	if err != nil {
		return nil, err
	}

	return b.SnapshotFromPaths(src, paths)
}

// SnapshotFromPaths constructs a StoreSnapshot from the provided
// slice of paths resolved against the provided fs.FS.
func (b *SnapshotBuilder) SnapshotFromPaths(ffs fs.FS, paths []string) (*Snapshot, error) {
	b.logger.Debug("opening state files", zap.Strings("paths", paths))

	var files []fs.File
	for _, file := range paths {
		fi, err := ffs.Open(file)
		if err != nil {
			return nil, err
		}

		files = append(files, fi)
	}

	return b.SnapshotFromFiles(files)
}

// SnapshotFromFiles constructs a StoreSnapshot from the provided slice
// of fs.File implementations.
func (b *SnapshotBuilder) SnapshotFromFiles(files []fs.File) (*Snapshot, error) {
	var (
		so = newSnapshotOption(b.opts...)
		s  = EmptySnapshot()
	)

	for _, fi := range files {
		defer fi.Close()

		docs, err := documentsFromFile(fi, so)
		if err != nil {
			return nil, err
		}

		for _, doc := range docs {
			if err := s.addDoc(doc, b.dependencyGraph); err != nil {
				return nil, err
			}
		}
	}

	return s, nil
}

// WalkDocuments walks all the Flipt feature documents found in the target fs.FS
// based on either the default index file or an index file located in the root
func WalkDocuments(logger *zap.Logger, conf *Config, src fs.FS, fn func(*ext.Document) error) error {
	paths, err := conf.List(src)
	if err != nil {
		return err
	}

	so := newSnapshotOption()
	for _, file := range paths {
		logger.Debug("opening state file", zap.String("path", file))

		fi, err := src.Open(file)
		if err != nil {
			return err
		}
		defer fi.Close()

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

func EmptySnapshot() *Snapshot {
	return &Snapshot{
		ns: map[string]*namespace{
			flipt.DefaultNamespace: newNamespace(),
		},
		evalDists: map[string][]*storage.EvaluationDistribution{},
		evalSnap: &evaluation.EvaluationSnapshot{
			Namespaces: map[string]*evaluation.EvaluationNamespaceSnapshot{},
		},
		now: flipt.Now(),
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

	var (
		buf    = &bytes.Buffer{}
		reader = io.TeeReader(fi, buf)
		docs   []*ext.Document
		extn   = filepath.Ext(stat.Name())
	)

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

	var (
		hash = sha1.New() //nolint:gosec
		_, _ = hash.Write(buf.Bytes())
		// etag is the sha1 hash of the document
		etag = fmt.Sprintf("%x", hash.Sum(nil))
	)

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

		doc.Etag = etag

		docs = append(docs, doc)
	}

	return docs, nil
}

// addDoc is the heart of the snapshot building code
// Currently it is full of duplication and complexity due to surving three very similar
// needs that could be collapsed and some dropped once we remove support for legacy
// codepaths (v1 types / eval).
// The snapshot generated contains all the necessary state to serve server-side
// evaluation as well as returning entire snapshot state for client-side evaluation.
func (s *Snapshot) addDoc(doc *ext.Document, dependencyGraph *configcoreflipt.DependencyGraph) error {
	var (
		namespaceKey = doc.Namespace.GetKey()
		ns           = s.ns[namespaceKey]
		snap         = s.evalSnap.Namespaces[namespaceKey]
	)

	if ns == nil {
		ns = newNamespace()
		s.ns[namespaceKey] = ns
	}

	if snap == nil {
		snap = &evaluation.EvaluationNamespaceSnapshot{
			Namespace: &evaluation.EvaluationNamespace{
				Key: namespaceKey,
			},
			Flags: make([]*evaluation.EvaluationFlag, 0, len(doc.Flags)),
		}
		s.evalSnap.Namespaces[namespaceKey] = snap
	}

	evalDists := map[string][]*storage.EvaluationDistribution{}
	if len(s.evalDists) > 0 {
		evalDists = s.evalDists
	}

	evalSnapSegments := map[string]*evaluation.EvaluationSegment{}
	for _, s := range doc.Segments {
		var (
			matchType = core.MatchType_value[s.MatchType]
			segment   = &core.Segment{
				Name:        s.Name,
				Key:         s.Key,
				Description: s.Description,
				MatchType:   core.MatchType(matchType),
			}
			evalSnapSegment = &evaluation.EvaluationSegment{
				Key:         s.Key,
				Name:        s.Name,
				Description: s.Description,
				MatchType:   toEvaluationSegmentMatchType(core.MatchType(matchType)),
			}
		)

		for _, constraint := range s.Constraints {
			constraintType := core.ComparisonType(core.ComparisonType_value[constraint.Type])
			segment.Constraints = append(segment.Constraints, &core.Constraint{
				Operator:    constraint.Operator,
				Property:    constraint.Property,
				Type:        constraintType,
				Value:       constraint.Value,
				Description: constraint.Description,
			})

			evalSnapSegment.Constraints = append(evalSnapSegment.Constraints, &evaluation.EvaluationConstraint{
				Type:     toEvaluationConstraintComparisonType(constraintType),
				Operator: constraint.Operator,
				Property: constraint.Property,
				Value:    constraint.Value,
			})
		}

		ns.segments[segment.Key] = segment
		evalSnapSegments[segment.Key] = evalSnapSegment
	}

	for _, f := range doc.Flags {
		var (
			flagType = core.FlagType_value[f.Type]
			flag     = &core.Flag{
				Key:         f.Key,
				Name:        f.Name,
				Description: f.Description,
				Enabled:     f.Enabled,
				Type:        core.FlagType(flagType),
			}

			// evaluation snapshot

			evalSnapFlag = &evaluation.EvaluationFlag{
				Key:         f.Key,
				Name:        f.Name,
				Description: f.Description,
				Enabled:     f.Enabled,
				Type:        toEvaluationFlagType(f.Type),
				CreatedAt:   s.now,
				UpdatedAt:   s.now,
			}
		)

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

				attachment, err := json.Marshal(v.Attachment)
				if err != nil {
					return err
				}

				// evaluation

				evalSnapFlag.DefaultVariant = &evaluation.EvaluationVariant{
					Id:         path.Join(flag.Key, variant.Key),
					Key:        v.Key,
					Attachment: string(attachment),
				}
			}
		}

		ns.flags[f.Key] = flag
		snap.Flags = append(snap.Flags, evalSnapFlag)

		evalRules := []*storage.EvaluationRule{}
		for i, r := range f.Rules {
			var (
				rank = int32(i + 1)
				rule = &core.Rule{
					Segments: r.Segment.Keys,
				}
				evalRule = &storage.EvaluationRule{
					ID:           uuid.NewString(),
					NamespaceKey: namespaceKey,
					FlagKey:      f.Key,
					Rank:         rank,
				}

				evalSnapRule = &evaluation.EvaluationRule{
					Id:   evalRule.ID,
					Rank: evalRule.Rank,
				}
			)

			segmentOperator := core.SegmentOperator(core.SegmentOperator_value[r.Segment.Operator])
			rule.SegmentOperator = segmentOperator
			evalRule.SegmentOperator = segmentOperator
			evalSnapRule.SegmentOperator = toEvaluationSegmentOperator(rule.SegmentOperator)

			var (
				segmentKeys = []string{}
				segments    = make(map[string]*storage.EvaluationSegment)
			)

			if count := len(rule.Segments); count > 0 {
				segmentKeys = append(segmentKeys, rule.Segments...)
				evalSnapRule.Segments = make([]*evaluation.EvaluationSegment, 0, count)
			}

			for _, segmentKey := range segmentKeys {
				segment := ns.segments[segmentKey]
				if segment == nil {
					return errs.ErrInvalidf("flag %s/%s rule %d references unknown segment %q", doc.Namespace, flag.Key, rank, segmentKey)
				}

				// track dependency between flag and segment
				dependencyGraph.AddDependency(configcoreflipt.ResourceID{
					Namespace: doc.Namespace.GetKey(),
					Key:       flag.Key,
					Type:      environments.FlagResourceType,
				}, configcoreflipt.ResourceID{
					Namespace: doc.Namespace.GetKey(),
					Key:       segmentKey,
					Type:      environments.SegmentResourceType,
				})

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

				if segment, ok := evalSnapSegments[segmentKey]; ok {
					evalSnapRule.Segments = append(evalSnapRule.Segments, segment)
				}
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

				evalSnapRule.Distributions = append(evalSnapRule.Distributions, &evaluation.EvaluationDistribution{
					RuleId:            evalRule.ID,
					Rollout:           d.Rollout,
					VariantKey:        variant.Key,
					VariantAttachment: string(attachment),
				})
			}

			flag.Rules = append(flag.Rules, rule)
			evalSnapFlag.Rules = append(evalSnapFlag.Rules, evalSnapRule)
		}

		ns.evalRules[f.Key] = evalRules

		evalRollouts := make([]*storage.EvaluationRollout, 0, len(f.Rollouts))
		for i, rollout := range f.Rollouts {
			var (
				rank        = int32(i + 1)
				flagRollout = &core.Rollout{
					Description: rollout.Description,
				}

				evalRollout = &storage.EvaluationRollout{
					NamespaceKey: namespaceKey,
					Rank:         rank,
				}

				evalSnapRollout = &evaluation.EvaluationRollout{
					Rank: rank,
				}
			)

			if rollout.Threshold != nil {
				evalRollout.Threshold = &storage.RolloutThreshold{
					Percentage: rollout.Threshold.Percentage,
					Value:      rollout.Threshold.Value,
				}
				evalRollout.RolloutType = core.RolloutType_THRESHOLD_ROLLOUT_TYPE

				flagRollout.Type = evalRollout.RolloutType
				flagRollout.Rule = &core.Rollout_Threshold{
					Threshold: &core.RolloutThreshold{
						Percentage: rollout.Threshold.Percentage,
						Value:      rollout.Threshold.Value,
					},
				}

				evalSnapRollout.Rule = &evaluation.EvaluationRollout_Threshold{
					Threshold: &evaluation.EvaluationRolloutThreshold{
						Percentage: rollout.Threshold.Percentage,
						Value:      rollout.Threshold.Value,
					},
				}
				evalSnapRollout.Type = evaluation.EvaluationRolloutType_THRESHOLD_ROLLOUT_TYPE
			} else if rollout.Segment != nil {
				var (
					segmentKeys = []string{}
					segments    = make(map[string]*storage.EvaluationSegment)
				)

				if len(rollout.Segment.Keys) > 0 {
					segmentKeys = append(segmentKeys, rollout.Segment.Keys...)
				}

				evalSnapRolloutSegment := &evaluation.EvaluationRollout_Segment{
					Segment: &evaluation.EvaluationRolloutSegment{
						Value: rollout.Segment.Value,
					},
				}
				evalSnapRollout.Rule = evalSnapRolloutSegment

				for _, segmentKey := range segmentKeys {
					segment, ok := ns.segments[segmentKey]
					if !ok {
						return errs.ErrInvalidf("flag %s/%s rule %d references unknown segment %q", doc.Namespace, flag.Key, rank, segmentKey)
					}

					// track dependency between flag and segment
					dependencyGraph.AddDependency(configcoreflipt.ResourceID{
						Namespace: doc.Namespace.GetKey(),
						Key:       flag.Key,
						Type:      environments.FlagResourceType,
					}, configcoreflipt.ResourceID{
						Namespace: doc.Namespace.GetKey(),
						Key:       segmentKey,
						Type:      environments.SegmentResourceType,
					})

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

					if segment, ok := evalSnapSegments[segmentKey]; ok {
						evalSnapRolloutSegment.Segment.Segments = append(evalSnapRolloutSegment.Segment.Segments,
							segment)
					}
				}

				segmentOperator := core.SegmentOperator(core.SegmentOperator_value[rollout.Segment.Operator])
				evalRollout.Segment = &storage.RolloutSegment{
					Segments:        segments,
					SegmentOperator: segmentOperator,
					Value:           rollout.Segment.Value,
				}

				evalRollout.RolloutType = core.RolloutType_SEGMENT_ROLLOUT_TYPE

				evalSnapRolloutSegment.Segment.SegmentOperator = toEvaluationSegmentOperator(segmentOperator)
				evalSnapRollout.Type = evaluation.EvaluationRolloutType_SEGMENT_ROLLOUT_TYPE

				frs := &core.RolloutSegment{
					Value:           rollout.Segment.Value,
					SegmentOperator: segmentOperator,
					Segments:        segmentKeys,
				}

				flagRollout.Type = evalRollout.RolloutType
				flagRollout.Rule = &core.Rollout_Segment{
					Segment: frs,
				}
			}

			flag.Rollouts = append(flag.Rollouts, flagRollout)

			evalRollouts = append(evalRollouts, evalRollout)
			evalSnapFlag.Rollouts = append(evalSnapFlag.Rollouts, evalSnapRollout)
		}

		ns.evalRollouts[f.Key] = evalRollouts
	}

	ns.etag = doc.Etag
	s.ns[namespaceKey] = ns
	snap.Digest = doc.Etag
	s.evalSnap.Namespaces[namespaceKey] = snap
	s.evalDists = evalDists

	return nil
}

func (s Snapshot) String() string {
	return "snapshot"
}

func (s *Snapshot) GetFlag(ctx context.Context, req storage.ResourceRequest) (*core.Flag, error) {
	ns, err := s.getNamespace(req.Namespace())
	if err != nil {
		return nil, err
	}

	flag, ok := ns.flags[req.Key]
	if !ok {
		return nil, errs.ErrNotFoundf("flag %q", req)
	}

	return flag, nil
}

func (s *Snapshot) ListFlags(ctx context.Context, req *storage.ListRequest[storage.NamespaceRequest]) (set storage.ResultSet[*core.Flag], err error) {
	ns, err := s.getNamespace(req.Predicate.Namespace())
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

func (s *Snapshot) CountFlags(ctx context.Context, p storage.NamespaceRequest) (uint64, error) {
	ns, err := s.getNamespace(p.Namespace())
	if err != nil {
		return 0, err
	}

	return uint64(len(ns.flags)), nil
}

func (s *Snapshot) GetEvaluationRules(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRule, error) {
	ns, ok := s.ns[flag.Namespace()]
	if !ok {
		return nil, errs.ErrNotFoundf("namespace %q", flag.NamespaceRequest)
	}

	rules, ok := ns.evalRules[flag.Key]
	if !ok {
		return nil, errs.ErrNotFoundf("flag %q", flag)
	}

	return rules, nil
}

func (s *Snapshot) GetEvaluationDistributions(ctx context.Context, flag storage.ResourceRequest, rule storage.IDRequest) ([]*storage.EvaluationDistribution, error) {
	_, ok := s.ns[flag.Namespace()]
	if !ok {
		return nil, errs.ErrNotFoundf("namespace %q", flag.NamespaceRequest)
	}

	dists, ok := s.evalDists[rule.ID]
	if !ok {
		return []*storage.EvaluationDistribution{}, nil
	}

	return dists, nil
}

func (s *Snapshot) GetEvaluationRollouts(ctx context.Context, flag storage.ResourceRequest) ([]*storage.EvaluationRollout, error) {
	ns, ok := s.ns[flag.Namespace()]
	if !ok {
		return nil, errs.ErrNotFoundf("namespace %q", flag.NamespaceRequest)
	}

	rollouts, ok := ns.evalRollouts[flag.Key]
	if !ok {
		return nil, errs.ErrNotFoundf("flag %q", flag)
	}

	return rollouts, nil
}

func (s *Snapshot) EvaluationNamespaceSnapshot(_ context.Context, ns string) (*evaluation.EvaluationNamespaceSnapshot, error) {
	snap, ok := s.evalSnap.Namespaces[ns]
	if !ok {
		return nil, errs.ErrNotFoundf("evaluation snapshot for namespace %q", ns)
	}

	return snap, nil
}

func (s *Snapshot) EvaluationSnapshot(_ context.Context) (*evaluation.EvaluationSnapshot, error) {
	return s.evalSnap, nil
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

func (s *Snapshot) getNamespace(key string) (namespace, error) {
	ns, ok := s.ns[key]
	if !ok {
		return namespace{}, errs.ErrNotFoundf("namespace %q", key)
	}

	return *ns, nil
}

func toEvaluationFlagType(typ string) evaluation.EvaluationFlagType {
	switch core.FlagType(core.FlagType_value[typ]) {
	case core.FlagType_BOOLEAN_FLAG_TYPE:
		return evaluation.EvaluationFlagType_BOOLEAN_FLAG_TYPE
	case core.FlagType_VARIANT_FLAG_TYPE:
		return evaluation.EvaluationFlagType_VARIANT_FLAG_TYPE
	}
	return evaluation.EvaluationFlagType_VARIANT_FLAG_TYPE
}

func toEvaluationSegmentMatchType(s core.MatchType) evaluation.EvaluationSegmentMatchType {
	switch s {
	case core.MatchType_ANY_MATCH_TYPE:
		return evaluation.EvaluationSegmentMatchType_ANY_SEGMENT_MATCH_TYPE
	case core.MatchType_ALL_MATCH_TYPE:
		return evaluation.EvaluationSegmentMatchType_ALL_SEGMENT_MATCH_TYPE
	}
	return evaluation.EvaluationSegmentMatchType_ANY_SEGMENT_MATCH_TYPE
}

func toEvaluationSegmentOperator(s core.SegmentOperator) evaluation.EvaluationSegmentOperator {
	switch s {
	case core.SegmentOperator_OR_SEGMENT_OPERATOR:
		return evaluation.EvaluationSegmentOperator_OR_SEGMENT_OPERATOR
	case core.SegmentOperator_AND_SEGMENT_OPERATOR:
		return evaluation.EvaluationSegmentOperator_AND_SEGMENT_OPERATOR
	}
	return evaluation.EvaluationSegmentOperator_OR_SEGMENT_OPERATOR
}

func toEvaluationConstraintComparisonType(c core.ComparisonType) evaluation.EvaluationConstraintComparisonType {
	switch c {
	case core.ComparisonType_STRING_COMPARISON_TYPE:
		return evaluation.EvaluationConstraintComparisonType_STRING_CONSTRAINT_COMPARISON_TYPE
	case core.ComparisonType_NUMBER_COMPARISON_TYPE:
		return evaluation.EvaluationConstraintComparisonType_NUMBER_CONSTRAINT_COMPARISON_TYPE
	case core.ComparisonType_DATETIME_COMPARISON_TYPE:
		return evaluation.EvaluationConstraintComparisonType_DATETIME_CONSTRAINT_COMPARISON_TYPE
	case core.ComparisonType_BOOLEAN_COMPARISON_TYPE:
		return evaluation.EvaluationConstraintComparisonType_BOOLEAN_CONSTRAINT_COMPARISON_TYPE
	case core.ComparisonType_ENTITY_ID_COMPARISON_TYPE:
		return evaluation.EvaluationConstraintComparisonType_ENTITY_ID_CONSTRAINT_COMPARISON_TYPE
	}
	return evaluation.EvaluationConstraintComparisonType_UNKNOWN_CONSTRAINT_COMPARISON_TYPE
}
