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

	"github.com/gobwas/glob"
	"github.com/gofrs/uuid"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/cue"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v3"
)

const (
	indexFile = ".flipt.yml"
	defaultNs = "default"
)

var (
	_                 storage.Store = (*StoreSnapshot)(nil)
	ErrNotImplemented               = errors.New("not implemented")
)

// FliptIndex represents the structure of a well-known file ".flipt.yml"
// at the root of an FS.
type FliptIndex struct {
	Version string   `yaml:"version,omitempty"`
	Include []string `yaml:"include,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
}

// StoreSnapshot contains the structures necessary for serving
// flag state to a client.
type StoreSnapshot struct {
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
}

func newNamespace(key, name string, created *timestamppb.Timestamp) *namespace {
	return &namespace{
		resource: &flipt.Namespace{
			Key:       key,
			Name:      name,
			CreatedAt: created,
			UpdatedAt: created,
		},
		flags:        map[string]*flipt.Flag{},
		segments:     map[string]*flipt.Segment{},
		rules:        map[string]*flipt.Rule{},
		rollouts:     map[string]*flipt.Rollout{},
		evalRules:    map[string][]*storage.EvaluationRule{},
		evalRollouts: map[string][]*storage.EvaluationRollout{},
	}
}

// SnapshotFromFS is a convenience function for building a snapshot
// directly from an implementation of fs.FS using the list state files
// function to source the relevant Flipt configuration files.
func SnapshotFromFS(logger *zap.Logger, fs fs.FS) (*StoreSnapshot, error) {
	files, err := listStateFiles(logger, fs)
	if err != nil {
		return nil, err
	}

	logger.Debug("opening state files", zap.Strings("paths", files))

	return SnapshotFromPaths(logger, fs, files...)
}

// SnapshotFromPaths constructs a StoreSnapshot from the provided
// slice of paths resolved against the provided fs.FS.
func SnapshotFromPaths(logger *zap.Logger, ffs fs.FS, paths ...string) (*StoreSnapshot, error) {
	var files []fs.File
	for _, file := range paths {
		fi, err := ffs.Open(file)
		if err != nil {
			return nil, err
		}

		files = append(files, fi)
	}

	return SnapshotFromFiles(logger, files...)
}

// SnapshotFromFiles constructs a StoreSnapshot from the provided slice
// of fs.File implementations.
func SnapshotFromFiles(logger *zap.Logger, files ...fs.File) (*StoreSnapshot, error) {
	now := flipt.Now()
	s := StoreSnapshot{
		ns: map[string]*namespace{
			defaultNs: newNamespace("default", "Default", now),
		},
		evalDists: map[string][]*storage.EvaluationDistribution{},
		now:       now,
	}

	for _, fi := range files {
		defer fi.Close()

		docs, err := documentsFromFile(fi)
		if err != nil {
			return nil, err
		}

		for _, doc := range docs {
			if err := s.addDoc(doc); err != nil {
				return nil, err
			}
		}
	}

	return &s, nil
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

		docs, err := documentsFromFile(fi)
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

// documentsFromFile parses and validates a document from a single fs.File instance
func documentsFromFile(fi fs.File) ([]*ext.Document, error) {
	validator, err := cue.NewFeaturesValidator()
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
		if doc.Namespace == "" {
			doc.Namespace = "default"
		}
		docs = append(docs, doc)
	}

	return docs, nil
}

// listStateFiles lists all the file paths in a provided fs.FS containing Flipt feature state
func listStateFiles(logger *zap.Logger, source fs.FS) ([]string, error) {
	// This is the default variable + value for the FliptIndex. It will preserve its value if
	// a .flipt.yml can not be read for whatever reason.
	idx := FliptIndex{
		Version: "1.0",
		Include: []string{
			"**features.yml", "**features.yaml", "**.features.yml", "**.features.yaml",
		},
	}

	// Read index file
	inFile, err := source.Open(indexFile)
	if err == nil {
		if derr := yaml.NewDecoder(inFile).Decode(&idx); derr != nil {
			return nil, fmt.Errorf("yaml: %w", derr)
		}
	}

	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return nil, err
		} else {
			logger.Debug("index file does not exist, defaulting...", zap.String("file", indexFile), zap.Error(err))
		}
	}

	var includes []glob.Glob
	for _, g := range idx.Include {
		glob, err := glob.Compile(g)
		if err != nil {
			return nil, fmt.Errorf("compiling include glob: %w", err)
		}

		includes = append(includes, glob)
	}

	filenames := make([]string, 0)
	if err := fs.WalkDir(source, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		for _, glob := range includes {
			if glob.Match(path) {
				filenames = append(filenames, path)
				return nil
			}
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if len(idx.Exclude) > 0 {
		var excludes []glob.Glob
		for _, g := range idx.Exclude {
			glob, err := glob.Compile(g)
			if err != nil {
				return nil, fmt.Errorf("compiling include glob: %w", err)
			}

			excludes = append(excludes, glob)
		}

	OUTER:
		for i := len(filenames) - 1; i >= 0; i-- {
			for _, glob := range excludes {
				if glob.Match(filenames[i]) {
					filenames = append(filenames[:i], filenames[i+1:]...)
					continue OUTER
				}
			}
		}
	}

	return filenames, nil
}

func (ss *StoreSnapshot) addDoc(doc *ext.Document) error {
	ns := ss.ns[doc.Namespace]
	if ns == nil {
		ns = newNamespace(doc.Namespace, doc.Namespace, ss.now)
	}

	evalDists := map[string][]*storage.EvaluationDistribution{}
	if len(ss.evalDists) > 0 {
		evalDists = ss.evalDists
	}

	for _, s := range doc.Segments {
		matchType := flipt.MatchType_value[s.MatchType]
		segment := &flipt.Segment{
			NamespaceKey: doc.Namespace,
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
				NamespaceKey: doc.Namespace,
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
			NamespaceKey: doc.Namespace,
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

			flag.Variants = append(flag.Variants, &flipt.Variant{
				Id:           uuid.Must(uuid.NewV4()).String(),
				NamespaceKey: doc.Namespace,
				Key:          v.Key,
				Name:         v.Name,
				Description:  v.Description,
				Attachment:   string(attachment),
				CreatedAt:    ss.now,
				UpdatedAt:    ss.now,
			})
		}

		ns.flags[f.Key] = flag

		evalRules := []*storage.EvaluationRule{}
		for i, r := range f.Rules {
			rank := int32(i + 1)
			rule := &flipt.Rule{
				NamespaceKey: doc.Namespace,
				Id:           uuid.Must(uuid.NewV4()).String(),
				FlagKey:      f.Key,
				Rank:         rank,
				CreatedAt:    ss.now,
				UpdatedAt:    ss.now,
			}

			evalRule := &storage.EvaluationRule{
				NamespaceKey: doc.Namespace,
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
				NamespaceKey: doc.Namespace,
				Rank:         rank,
			}

			flagRollout := &flipt.Rollout{
				Id:           uuid.Must(uuid.NewV4()).String(),
				Rank:         rank,
				FlagKey:      f.Key,
				NamespaceKey: doc.Namespace,
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

	ss.ns[doc.Namespace] = ns

	ss.evalDists = evalDists

	return nil
}

func (ss StoreSnapshot) String() string {
	return "snapshot"
}

func (ss *StoreSnapshot) GetRule(ctx context.Context, namespaceKey string, id string) (rule *flipt.Rule, _ error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return nil, err
	}

	var ok bool
	rule, ok = ns.rules[id]
	if !ok {
		return nil, errs.ErrNotFoundf(`rule "%s/%s"`, namespaceKey, id)
	}

	return rule, nil
}

func (ss *StoreSnapshot) ListRules(ctx context.Context, namespaceKey string, flagKey string, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Rule], _ error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return set, err
	}

	rules := make([]*flipt.Rule, 0, len(ns.rules))
	for _, rule := range ns.rules {
		if rule.FlagKey == flagKey {
			rules = append(rules, rule)
		}
	}

	return paginate(storage.NewQueryParams(opts...), func(i, j int) bool {
		return rules[i].Rank < rules[j].Rank
	}, rules...)
}

func (ss *StoreSnapshot) CountRules(ctx context.Context, namespaceKey, flagKey string) (uint64, error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return 0, err
	}

	var count uint64 = 0
	for _, rule := range ns.rules {
		if rule.FlagKey == flagKey {
			count += 1
		}
	}

	return count, nil
}

func (ss *StoreSnapshot) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	return ErrNotImplemented
}

func (ss *StoreSnapshot) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	return ErrNotImplemented
}

func (ss *StoreSnapshot) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	return ErrNotImplemented
}

func (ss *StoreSnapshot) GetSegment(ctx context.Context, namespaceKey string, key string) (*flipt.Segment, error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return nil, err
	}

	segment, ok := ns.segments[key]
	if !ok {
		return nil, errs.ErrNotFoundf(`segment "%s/%s"`, namespaceKey, key)
	}

	return segment, nil
}

func (ss *StoreSnapshot) ListSegments(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Segment], err error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return set, err
	}

	segments := make([]*flipt.Segment, 0, len(ns.segments))
	for _, segment := range ns.segments {
		segments = append(segments, segment)
	}

	return paginate(storage.NewQueryParams(opts...), func(i, j int) bool {
		return segments[i].Key < segments[j].Key
	}, segments...)
}

func (ss *StoreSnapshot) CountSegments(ctx context.Context, namespaceKey string) (uint64, error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return 0, err
	}

	return uint64(len(ns.segments)), nil
}

func (ss *StoreSnapshot) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	return ErrNotImplemented
}

func (ss *StoreSnapshot) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	return ErrNotImplemented
}

func (ss *StoreSnapshot) GetNamespace(ctx context.Context, key string) (*flipt.Namespace, error) {
	ns, err := ss.getNamespace(key)
	if err != nil {
		return nil, err
	}

	return ns.resource, nil
}

func (ss *StoreSnapshot) ListNamespaces(ctx context.Context, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Namespace], err error) {
	ns := make([]*flipt.Namespace, 0, len(ss.ns))
	for _, n := range ss.ns {
		ns = append(ns, n.resource)
	}

	return paginate(storage.NewQueryParams(opts...), func(i, j int) bool {
		return ns[i].Key < ns[j].Key
	}, ns...)
}

func (ss *StoreSnapshot) CountNamespaces(ctx context.Context) (uint64, error) {
	return uint64(len(ss.ns)), nil
}

func (ss *StoreSnapshot) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) UpdateNamespace(ctx context.Context, r *flipt.UpdateNamespaceRequest) (*flipt.Namespace, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) DeleteNamespace(ctx context.Context, r *flipt.DeleteNamespaceRequest) error {
	return ErrNotImplemented
}

func (ss *StoreSnapshot) GetFlag(ctx context.Context, namespaceKey string, key string) (*flipt.Flag, error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return nil, err
	}

	flag, ok := ns.flags[key]
	if !ok {
		return nil, errs.ErrNotFoundf(`flag "%s/%s"`, namespaceKey, key)
	}

	return flag, nil
}

func (ss *StoreSnapshot) ListFlags(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Flag], err error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return set, err
	}

	flags := make([]*flipt.Flag, 0, len(ns.flags))
	for _, flag := range ns.flags {
		flags = append(flags, flag)
	}

	return paginate(storage.NewQueryParams(opts...), func(i, j int) bool {
		return flags[i].Key < flags[j].Key
	}, flags...)
}

func (ss *StoreSnapshot) CountFlags(ctx context.Context, namespaceKey string) (uint64, error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return 0, err
	}

	return uint64(len(ns.flags)), nil
}

func (ss *StoreSnapshot) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	return ErrNotImplemented
}

func (ss *StoreSnapshot) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	return ErrNotImplemented
}

func (ss *StoreSnapshot) GetEvaluationRules(ctx context.Context, namespaceKey string, flagKey string) ([]*storage.EvaluationRule, error) {
	ns, ok := ss.ns[namespaceKey]
	if !ok {
		return nil, errs.ErrNotFoundf("namespaced %q", namespaceKey)
	}

	rules, ok := ns.evalRules[flagKey]
	if !ok {
		return nil, errs.ErrNotFoundf(`flag "%s/%s"`, namespaceKey, flagKey)
	}

	return rules, nil
}

func (ss *StoreSnapshot) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	dists, ok := ss.evalDists[ruleID]
	if !ok {
		return nil, errs.ErrNotFoundf("rule %q", ruleID)
	}

	return dists, nil
}

func (ss *StoreSnapshot) GetEvaluationRollouts(ctx context.Context, namespaceKey, flagKey string) ([]*storage.EvaluationRollout, error) {
	ns, ok := ss.ns[namespaceKey]
	if !ok {
		return nil, errs.ErrNotFoundf("namespaced %q", namespaceKey)
	}

	rollouts, ok := ns.evalRollouts[flagKey]
	if !ok {
		return nil, errs.ErrNotFoundf(`flag "%s/%s"`, namespaceKey, flagKey)
	}

	return rollouts, nil
}

func (ss *StoreSnapshot) GetRollout(ctx context.Context, namespaceKey, id string) (*flipt.Rollout, error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return nil, err
	}

	rollout, ok := ns.rollouts[id]
	if !ok {
		return nil, errs.ErrNotFoundf(`rollout "%s/%s"`, namespaceKey, id)
	}

	return rollout, nil
}

func (ss *StoreSnapshot) ListRollouts(ctx context.Context, namespaceKey, flagKey string, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Rollout], err error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return set, err
	}

	rollouts := make([]*flipt.Rollout, 0)
	for _, rollout := range ns.rollouts {
		if rollout.FlagKey == flagKey {
			rollouts = append(rollouts, rollout)
		}
	}

	return paginate(storage.NewQueryParams(opts...), func(i, j int) bool {
		return rollouts[i].Rank < rollouts[j].Rank
	}, rollouts...)
}

func (ss *StoreSnapshot) CountRollouts(ctx context.Context, namespaceKey, flagKey string) (uint64, error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return 0, err
	}

	var count uint64 = 0
	for _, rollout := range ns.rollouts {
		if rollout.FlagKey == flagKey {
			count += 1
		}
	}

	return count, nil
}

func (ss *StoreSnapshot) CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (*flipt.Rollout, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) UpdateRollout(ctx context.Context, r *flipt.UpdateRolloutRequest) (*flipt.Rollout, error) {
	return nil, ErrNotImplemented
}

func (ss *StoreSnapshot) DeleteRollout(ctx context.Context, r *flipt.DeleteRolloutRequest) error {
	return ErrNotImplemented
}

func (ss *StoreSnapshot) OrderRollouts(ctx context.Context, r *flipt.OrderRolloutsRequest) error {
	return ErrNotImplemented
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

func (ss *StoreSnapshot) getNamespace(key string) (namespace, error) {
	ns, ok := ss.ns[key]
	if !ok {
		return namespace{}, errs.ErrNotFoundf("namespace %q", key)
	}

	return *ns, nil
}
