package fs

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"google.golang.org/protobuf/types/known/timestamppb"
	"gopkg.in/yaml.v2"
)

var ErrNotImplemented = errors.New("not implemented")

type storeSnapshot struct {
	ns        map[string]namespace
	evalDists map[string][]*storage.EvaluationDistribution
}

type namespace struct {
	resource  *flipt.Namespace
	flags     map[string]*flipt.Flag
	segments  map[string]*flipt.Segment
	rules     map[string]*flipt.Rule
	evalRules map[string][]*storage.EvaluationRule
}

func newNamespace(key string) namespace {
	return namespace{
		resource:  &flipt.Namespace{Key: key, Name: strings.Title(key)},
		flags:     map[string]*flipt.Flag{},
		segments:  map[string]*flipt.Segment{},
		rules:     map[string]*flipt.Rule{},
		evalRules: map[string][]*storage.EvaluationRule{},
	}
}

func newStoreSnapshot(source fs.FS) (*storeSnapshot, []string, error) {
	store := storeSnapshot{
		ns: map[string]namespace{
			"default": {
				resource: &flipt.Namespace{
					Key:  "default",
					Name: "Default",
				},
			},
		},
		evalDists: map[string][]*storage.EvaluationDistribution{},
	}

	dirs := []string{}
	if err := fs.WalkDir(source, ".", func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			// skip the git directory
			if path == ".git" {
				return fs.SkipDir
			}

			dirs = append(dirs, path)

			return nil
		}

		if filepath.Base(path) != "features.yaml" {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}

		var (
			dec = yaml.NewDecoder(file)
			doc = new(ext.Document)
		)

		if err := dec.Decode(doc); err != nil {
			return fmt.Errorf("unmarshalling document: %w", err)
		}

		return store.addDoc(doc)
	}); err != nil {
		return nil, nil, err
	}

	return &store, dirs, nil
}

func (f *storeSnapshot) GetRule(ctx context.Context, namespaceKey string, id string) (rule *flipt.Rule, _ error) {
	ns, err := f.getNamespace(namespaceKey)
	if err != nil {
		return nil, err
	}

	var ok bool
	rule, ok = ns.rules[id]
	if !ok {
		return nil, errors.ErrNotFoundf(`rule "%s/%s"`, namespaceKey, id)
	}

	return rule, nil
}

func (f *storeSnapshot) ListRules(ctx context.Context, namespaceKey string, flagKey string, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Rule], _ error) {
	ns, err := f.getNamespace(namespaceKey)
	if err != nil {
		return set, err
	}

	rules := make([]*flipt.Rule, 0, len(ns.rules))
	for _, rule := range ns.rules {
		rules = append(rules, rule)
	}

	set, err = paginate(ctx, storage.NewQueryParams(opts...), func(i, j int) bool {
		return rules[i].Rank < rules[j].Rank
	}, rules...)

	return set, err
}

func (f *storeSnapshot) CountRules(ctx context.Context, namespaceKey string) (uint64, error) {
	ns, err := f.getNamespace(namespaceKey)
	if err != nil {
		return 0, err
	}

	return uint64(len(ns.rules)), nil
}

func (f *storeSnapshot) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	return nil, ErrNotImplemented
}

func (f *storeSnapshot) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	return nil, ErrNotImplemented
}

func (f *storeSnapshot) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	return ErrNotImplemented
}

func (f *storeSnapshot) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	return ErrNotImplemented
}

func (f *storeSnapshot) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	return nil, ErrNotImplemented
}

func (f *storeSnapshot) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	return nil, ErrNotImplemented
}

func (f *storeSnapshot) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	return ErrNotImplemented
}

func (f *storeSnapshot) GetSegment(ctx context.Context, namespaceKey string, key string) (*flipt.Segment, error) {
	ns, err := f.getNamespace(namespaceKey)
	if err != nil {
		return nil, err
	}

	segment, ok := ns.segments[key]
	if !ok {
		return nil, errors.ErrNotFoundf(`segment "%s/%s"`, namespaceKey, key)
	}

	return segment, nil
}

func (f *storeSnapshot) ListSegments(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Segment], err error) {
	ns, err := f.getNamespace(namespaceKey)
	if err != nil {
		return set, err
	}

	segments := make([]*flipt.Segment, 0, len(ns.segments))
	for _, segment := range ns.segments {
		segments = append(segments, segment)
	}

	set, err = paginate(ctx, storage.NewQueryParams(opts...), func(i, j int) bool {
		return segments[i].CreatedAt.AsTime().Before(segments[j].CreatedAt.AsTime())
	}, segments...)

	return set, nil
}

func (f *storeSnapshot) CountSegments(ctx context.Context, namespaceKey string) (uint64, error) {
	ns, err := f.getNamespace(namespaceKey)
	if err != nil {
		return 0, err
	}

	return uint64(len(ns.segments)), nil
}

func (f *storeSnapshot) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	return nil, ErrNotImplemented
}

func (f *storeSnapshot) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	return nil, ErrNotImplemented
}

func (f *storeSnapshot) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	return ErrNotImplemented
}

func (f *storeSnapshot) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	return nil, ErrNotImplemented
}

func (f *storeSnapshot) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	return nil, ErrNotImplemented
}

func (f *storeSnapshot) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	return ErrNotImplemented
}

func (f *storeSnapshot) GetNamespace(ctx context.Context, key string) (*flipt.Namespace, error) {
	ns, err := f.getNamespace(key)
	if err != nil {
		return nil, err
	}

	return ns.resource, nil
}

func (f *storeSnapshot) ListNamespaces(ctx context.Context, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Namespace], err error) {
	set.Results = make([]*flipt.Namespace, 0, len(f.ns))
	for _, ns := range f.ns {
		set.Results = append(set.Results, ns.resource)
	}

	set, err = paginate(ctx, storage.NewQueryParams(opts...), func(i, j int) bool {
		return set.Results[i].CreatedAt.AsTime().Before(set.Results[j].CreatedAt.AsTime())
	}, set.Results...)

	return set, err
}

func (f *storeSnapshot) CountNamespaces(ctx context.Context) (uint64, error) {
	return uint64(len(f.ns)), nil
}

func (f *storeSnapshot) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	return nil, ErrNotImplemented
}

func (f *storeSnapshot) UpdateNamespace(ctx context.Context, r *flipt.UpdateNamespaceRequest) (*flipt.Namespace, error) {
	return nil, ErrNotImplemented
}

func (f *storeSnapshot) DeleteNamespace(ctx context.Context, r *flipt.DeleteNamespaceRequest) error {
	return ErrNotImplemented
}

func (f *storeSnapshot) GetFlag(ctx context.Context, namespaceKey string, key string) (*flipt.Flag, error) {
	ns, err := f.getNamespace(namespaceKey)
	if err != nil {
		return nil, err
	}

	flag, ok := ns.flags[key]
	if !ok {
		return nil, errors.ErrNotFoundf(`flag "%s/%s"`, namespaceKey, key)
	}

	return flag, nil
}

func (f *storeSnapshot) ListFlags(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Flag], err error) {
	ns, err := f.getNamespace(namespaceKey)
	if err != nil {
		return set, err
	}

	flags := make([]*flipt.Flag, 0, len(ns.flags))
	for _, flag := range ns.flags {
		flags = append(flags, flag)
	}

	set, err = paginate(ctx, storage.NewQueryParams(opts...), func(i, j int) bool {
		return flags[i].CreatedAt.AsTime().Before(flags[j].CreatedAt.AsTime())
	}, flags...)

	return set, nil
}

func (f *storeSnapshot) CountFlags(ctx context.Context, namespaceKey string) (uint64, error) {
	ns, err := f.getNamespace(namespaceKey)
	if err != nil {
		return 0, err
	}
	return uint64(len(ns.flags)), nil
}

func (f *storeSnapshot) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	return nil, ErrNotImplemented
}

func (f *storeSnapshot) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	return nil, ErrNotImplemented
}

func (f *storeSnapshot) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	return ErrNotImplemented
}

func (f *storeSnapshot) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	return nil, ErrNotImplemented
}

func (f *storeSnapshot) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	return nil, ErrNotImplemented
}

func (f *storeSnapshot) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	return ErrNotImplemented
}

func (e *storeSnapshot) getNamespace(key string) (namespace, error) {
	ns, ok := e.ns[key]
	if !ok {
		return namespace{}, errors.ErrNotFoundf("namespace %q", key)
	}

	return ns, nil
}

// GetEvaluationRules returns rules applicable to flagKey provided
func (e *storeSnapshot) GetEvaluationRules(ctx context.Context, namespaceKey string, flagKey string) ([]*storage.EvaluationRule, error) {
	ns, ok := e.ns[namespaceKey]
	if !ok {
		return nil, errors.ErrNotFoundf("namespaced %q", namespaceKey)
	}

	rules, ok := ns.evalRules[flagKey]
	if !ok {
		return nil, errors.ErrNotFoundf(`flag "%s/%s"`, namespaceKey, flagKey)
	}

	return rules, nil
}

func (e *storeSnapshot) GetEvaluationDistributions(ctx context.Context, ruleID string) ([]*storage.EvaluationDistribution, error) {
	dists, ok := e.evalDists[ruleID]
	if !ok {
		return nil, errors.ErrNotFoundf("rule %q", ruleID)
	}

	return dists, nil
}

func (e *storeSnapshot) addDoc(doc *ext.Document) error {
	ns := newNamespace(doc.Namespace)
	evalDists := map[string][]*storage.EvaluationDistribution{}

	now := timestamppb.Now()

	for _, s := range doc.Segments {
		matchType, _ := flipt.MatchType_value[s.MatchType]
		segment := &flipt.Segment{
			NamespaceKey: doc.Namespace,
			Name:         s.Name,
			Key:          s.Key,
			Description:  s.Description,
			MatchType:    flipt.MatchType(matchType),
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		for _, constraint := range s.Constraints {
			constraintType, _ := flipt.ComparisonType_value[constraint.Type]
			segment.Constraints = append(segment.Constraints, &flipt.Constraint{
				Id:        uuid.Must(uuid.NewV4()).String(),
				Operator:  constraint.Operator,
				Property:  constraint.Property,
				Type:      flipt.ComparisonType(constraintType),
				Value:     constraint.Value,
				CreatedAt: now,
				UpdatedAt: now,
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
			CreatedAt:    now,
			UpdatedAt:    now,
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
				CreatedAt:    now,
				UpdatedAt:    now,
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
				CreatedAt:    now,
				UpdatedAt:    now,
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
				return errors.ErrNotFoundf("segment %q in rule %d", rule.SegmentKey, rule.Rank)
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
					CreatedAt: now,
					UpdatedAt: now,
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

	// set state on snapshot
	e.ns[doc.Namespace] = ns
	e.evalDists = evalDists

	return nil
}

func ruleID(ns, flag string, id int) string {
	return fmt.Sprintf("%s/%s/%d", ns, flag, id)
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

func paginate[T any](ctx context.Context, params storage.QueryParams, less func(i, j int) bool, items ...T) (storage.ResultSet[T], error) {
	var set storage.ResultSet[T]

	// copy all auths into slice
	set.Results = make([]T, 0, len(items))
	for _, res := range items {
		set.Results = append(set.Results, res)
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
