package fs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"sort"
	"strconv"

	"github.com/gobwas/glob"
	"github.com/gofrs/uuid"
	ferrors "go.flipt.io/flipt/errors"
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

var ErrNotImplemented = errors.New("not implemented")

// FliptIndex represents the structure of a well-known file ".flipt.yml"
// at the root of an FS.
type FliptIndex struct {
	Version string   `yaml:"version,omitempty"`
	Include []string `yaml:"include,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
}

// storeSnapshot contains the structures necessary for serving
// flag state to a client.
type storeSnapshot struct {
	ns        map[string]*namespace
	evalDists map[string][]*storage.EvaluationDistribution
	now       *timestamppb.Timestamp
}

type namespace struct {
	resource  *flipt.Namespace
	flags     map[string]*flipt.Flag
	segments  map[string]*flipt.Segment
	rules     map[string]*flipt.Rule
	evalRules map[string][]*storage.EvaluationRule
}

func newNamespace(key, name string, created *timestamppb.Timestamp) *namespace {
	return &namespace{
		resource: &flipt.Namespace{
			Key:       key,
			Name:      name,
			CreatedAt: created,
			UpdatedAt: created,
		},
		flags:     map[string]*flipt.Flag{},
		segments:  map[string]*flipt.Segment{},
		rules:     map[string]*flipt.Rule{},
		evalRules: map[string][]*storage.EvaluationRule{},
	}
}

// snapshotFromFS is a convenience function for building a snapshot
// directly from an implementation of fs.FS using the list state files
// function to source the relevant Flipt configuration files.
func snapshotFromFS(logger *zap.Logger, fs fs.FS) (*storeSnapshot, error) {
	files, err := listStateFiles(logger, fs)
	if err != nil {
		return nil, err
	}

	logger.Debug("opening state files", zap.Strings("paths", files))

	var rds []io.Reader
	for _, file := range files {
		fi, err := fs.Open(file)
		if err != nil {
			return nil, err
		}

		defer fi.Close()
		rds = append(rds, fi)
	}

	return snapshotFromReaders(rds...)
}

// snapshotFromReaders constructs a storeSnapshot from the provided
// slice of io.Reader.
func snapshotFromReaders(sources ...io.Reader) (*storeSnapshot, error) {
	now := timestamppb.Now()
	s := storeSnapshot{
		ns: map[string]*namespace{
			defaultNs: newNamespace("default", "Default", now),
		},
		evalDists: map[string][]*storage.EvaluationDistribution{},
		now:       now,
	}

	for _, reader := range sources {
		doc := new(ext.Document)

		if err := yaml.NewDecoder(reader).Decode(doc); err != nil {
			return nil, err
		}

		// set namespace to default if empty in document
		if doc.Namespace == "" {
			doc.Namespace = "default"
		}

		if err := s.addDoc(doc); err != nil {
			return nil, err
		}

	}
	return &s, nil
}

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
		for i := range filenames {
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

func (ss *storeSnapshot) addDoc(doc *ext.Document) error {
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
		flag := &flipt.Flag{
			NamespaceKey: doc.Namespace,
			Key:          f.Key,
			Name:         f.Name,
			Description:  f.Description,
			Enabled:      f.Enabled,
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
		for _, r := range f.Rules {
			rule := &flipt.Rule{
				NamespaceKey: doc.Namespace,
				Id:           uuid.Must(uuid.NewV4()).String(),
				FlagKey:      f.Key,
				SegmentKey:   r.SegmentKey,
				Rank:         int32(r.Rank),
				CreatedAt:    ss.now,
				UpdatedAt:    ss.now,
			}

			evalRule := &storage.EvaluationRule{
				NamespaceKey: doc.Namespace,
				FlagKey:      f.Key,
				ID:           rule.Id,
				Rank:         rule.Rank,
				SegmentKey:   rule.SegmentKey,
			}

			segment := ns.segments[rule.SegmentKey]
			if segment == nil {
				return ferrors.ErrNotFoundf("segment %q in rule %d", rule.SegmentKey, rule.Rank)
			}

			evalRule.SegmentMatchType = segment.MatchType

			for _, constraint := range segment.Constraints {
				evalRule.Constraints = append(evalRule.Constraints, storage.EvaluationConstraint{
					Operator: constraint.Operator,
					Property: constraint.Property,
					Type:     constraint.Type,
					Value:    constraint.Value,
				})
			}

			evalRules = append(evalRules, evalRule)

			for _, d := range r.Distributions {
				variant, found := findByKey(d.VariantKey, flag.Variants...)
				if !found {
					continue
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
	}

	ss.ns[doc.Namespace] = ns

	ss.evalDists = evalDists

	return nil
}

func (ss storeSnapshot) String() string {
	return "snapshot"
}

func (ss *storeSnapshot) GetRule(ctx context.Context, namespaceKey string, id string) (rule *flipt.Rule, _ error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return nil, err
	}

	var ok bool
	rule, ok = ns.rules[id]
	if !ok {
		return nil, ferrors.ErrNotFoundf(`rule "%s/%s"`, namespaceKey, id)
	}

	return rule, nil
}

func (ss *storeSnapshot) ListRules(ctx context.Context, namespaceKey string, flagKey string, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Rule], _ error) {
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

	set = paginate(storage.NewQueryParams(opts...), func(i, j int) bool {
		return rules[i].Rank < rules[j].Rank
	}, rules...)

	return set, err
}

func (ss *storeSnapshot) CountRules(ctx context.Context, namespaceKey string) (uint64, error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return 0, err
	}

	return uint64(len(ns.rules)), nil
}

func (ss *storeSnapshot) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	return nil, ErrNotImplemented
}

func (ss *storeSnapshot) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	return nil, ErrNotImplemented
}

func (ss *storeSnapshot) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	return ErrNotImplemented
}

func (ss *storeSnapshot) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	return ErrNotImplemented
}

func (ss *storeSnapshot) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	return nil, ErrNotImplemented
}

func (ss *storeSnapshot) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	return nil, ErrNotImplemented
}

func (ss *storeSnapshot) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	return ErrNotImplemented
}

func (ss *storeSnapshot) GetSegment(ctx context.Context, namespaceKey string, key string) (*flipt.Segment, error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return nil, err
	}

	segment, ok := ns.segments[key]
	if !ok {
		return nil, ferrors.ErrNotFoundf(`segment "%s/%s"`, namespaceKey, key)
	}

	return segment, nil
}

func (ss *storeSnapshot) ListSegments(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Segment], err error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return set, err
	}

	segments := make([]*flipt.Segment, 0, len(ns.segments))
	for _, segment := range ns.segments {
		segments = append(segments, segment)
	}

	set = paginate(storage.NewQueryParams(opts...), func(i, j int) bool {
		return segments[i].Key < segments[j].Key
	}, segments...)

	return set, nil
}

func (ss *storeSnapshot) CountSegments(ctx context.Context, namespaceKey string) (uint64, error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return 0, err
	}

	return uint64(len(ns.segments)), nil
}

func (ss *storeSnapshot) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	return nil, ErrNotImplemented
}

func (ss *storeSnapshot) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	return nil, ErrNotImplemented
}

func (ss *storeSnapshot) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	return ErrNotImplemented
}

func (ss *storeSnapshot) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	return nil, ErrNotImplemented
}

func (ss *storeSnapshot) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	return nil, ErrNotImplemented
}

func (ss *storeSnapshot) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	return ErrNotImplemented
}

func (ss *storeSnapshot) GetNamespace(ctx context.Context, key string) (*flipt.Namespace, error) {
	ns, err := ss.getNamespace(key)
	if err != nil {
		return nil, err
	}

	return ns.resource, nil
}

func (ss *storeSnapshot) ListNamespaces(ctx context.Context, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Namespace], err error) {
	ns := make([]*flipt.Namespace, 0, len(ss.ns))
	for _, n := range ss.ns {
		ns = append(ns, n.resource)
	}

	set = paginate(storage.NewQueryParams(opts...), func(i, j int) bool {
		return ns[i].Key < ns[j].Key
	}, ns...)

	return set, err
}

func (ss *storeSnapshot) CountNamespaces(ctx context.Context) (uint64, error) {
	return uint64(len(ss.ns)), nil
}

func (ss *storeSnapshot) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	return nil, ErrNotImplemented
}

func (ss *storeSnapshot) UpdateNamespace(ctx context.Context, r *flipt.UpdateNamespaceRequest) (*flipt.Namespace, error) {
	return nil, ErrNotImplemented
}

func (ss *storeSnapshot) DeleteNamespace(ctx context.Context, r *flipt.DeleteNamespaceRequest) error {
	return ErrNotImplemented
}

func (ss *storeSnapshot) GetFlag(ctx context.Context, namespaceKey string, key string) (*flipt.Flag, error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return nil, err
	}

	flag, ok := ns.flags[key]
	if !ok {
		return nil, ferrors.ErrNotFoundf(`flag "%s/%s"`, namespaceKey, key)
	}

	return flag, nil
}

func (ss *storeSnapshot) ListFlags(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Flag], err error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return set, err
	}

	flags := make([]*flipt.Flag, 0, len(ns.flags))
	for _, flag := range ns.flags {
		flags = append(flags, flag)
	}

	set = paginate(storage.NewQueryParams(opts...), func(i, j int) bool {
		return flags[i].Key < flags[j].Key
	}, flags...)

	return set, nil
}

func (ss *storeSnapshot) CountFlags(ctx context.Context, namespaceKey string) (uint64, error) {
	ns, err := ss.getNamespace(namespaceKey)
	if err != nil {
		return 0, err
	}

	return uint64(len(ns.flags)), nil
}

func (ss *storeSnapshot) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	return nil, ErrNotImplemented
}

func (ss *storeSnapshot) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	return nil, ErrNotImplemented
}

func (ss *storeSnapshot) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	return ErrNotImplemented
}

func (ss *storeSnapshot) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	return nil, ErrNotImplemented
}

func (ss *storeSnapshot) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	return nil, ErrNotImplemented
}

func (ss *storeSnapshot) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	return ErrNotImplemented
}

func (ss *storeSnapshot) getNamespace(key string) (namespace, error) {
	ns, ok := ss.ns[key]
	if !ok {
		return namespace{}, ferrors.ErrNotFoundf("namespace %q", key)
	}

	return *ns, nil
}

func (ss *storeSnapshot) GetEvaluationRules(ctx context.Context, namespaceKey string, flagKey string) ([]*storage.EvaluationRule, error) {
	ns, ok := ss.ns[namespaceKey]
	if !ok {
		return nil, ferrors.ErrNotFoundf("namespaced %q", namespaceKey)
	}

	rules, ok := ns.evalRules[flagKey]
	if !ok {
		return nil, ferrors.ErrNotFoundf(`flag "%s/%s"`, namespaceKey, flagKey)
	}

	return rules, nil
}

func (ss *storeSnapshot) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	dists, ok := ss.evalDists[ruleID]
	if !ok {
		return nil, ferrors.ErrNotFoundf("rule %q", ruleID)
	}

	return dists, nil
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

func paginate[T any](params storage.QueryParams, less func(i, j int) bool, items ...T) storage.ResultSet[T] {
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
	if v, err := strconv.ParseInt(params.PageToken, 10, 64); err == nil {
		offset = int(v)
	}

	if offset >= len(set.Results) {
		return storage.ResultSet[T]{}
	}

	// 0 means no limit on page size (all items from offset)
	if params.Limit == 0 {
		set.Results = set.Results[offset:]
		return set
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

	return set
}
