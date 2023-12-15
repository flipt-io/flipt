package fs

import (
	"context"
	"errors"
	"fmt"
	"path"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
)

var (
	_ storage.Store = (*Store)(nil)

	// ErrNotImplemented is returned when a method has intentionally not been implemented
	// This is usually reserved for the store write actions when the store is read-only
	// but still needs to implement storage.Store
	ErrNotImplemented = errors.New("not implemented")
)

// SnapshotStore is a type which has a single function View.
// View is a functional transaction interface for reading a snapshot
// during the lifetime of a supplied function.
type SnapshotStore interface {
	// View accepts a function which takes a *StoreSnapshot.
	// The SnapshotStore will supply a snapshot which is valid
	// for the lifetime of the provided function call.
	View(func(storage.ReadOnlyStore) error) error
	fmt.Stringer
}

// Store embeds a StoreSnapshot and wraps the Store methods with a read-write mutex
// to synchronize reads with atomic replacements of the embedded snapshot.
type Store struct {
	viewer SnapshotStore
}

func NewStore(viewer SnapshotStore) *Store {
	return &Store{viewer: viewer}
}

func (s *Store) String() string {
	return path.Join("declarative", s.viewer.String())
}

func (s *Store) GetFlag(ctx context.Context, namespaceKey string, key string) (flag *flipt.Flag, err error) {
	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return flag, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		flag, err = ss.GetFlag(ctx, namespaceKey, key)
		return err
	})
}

func (s *Store) ListFlags(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Flag], err error) {
	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return set, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		set, err = ss.ListFlags(ctx, namespaceKey, opts...)
		return err
	})
}

func (s *Store) CountFlags(ctx context.Context, namespaceKey string) (count uint64, err error) {
	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return count, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		count, err = ss.CountFlags(ctx, namespaceKey)
		return err
	})
}

func (s *Store) GetRule(ctx context.Context, namespaceKey string, id string) (rule *flipt.Rule, err error) {
	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return rule, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		rule, err = ss.GetRule(ctx, namespaceKey, id)
		return err
	})
}

func (s *Store) ListRules(ctx context.Context, namespaceKey string, flagKey string, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Rule], err error) {
	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return set, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		set, err = ss.ListRules(ctx, namespaceKey, flagKey, opts...)
		return err
	})
}

func (s *Store) CountRules(ctx context.Context, namespaceKey, flagKey string) (count uint64, err error) {
	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return count, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		count, err = ss.CountRules(ctx, namespaceKey, flagKey)
		return err
	})
}

func (s *Store) GetSegment(ctx context.Context, namespaceKey string, key string) (segment *flipt.Segment, err error) {
	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return segment, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		segment, err = ss.GetSegment(ctx, namespaceKey, key)
		return err
	})
}

func (s *Store) ListSegments(ctx context.Context, namespaceKey string, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Segment], err error) {
	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return set, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		set, err = ss.ListSegments(ctx, namespaceKey, opts...)
		return err
	})
}

func (s *Store) CountSegments(ctx context.Context, namespaceKey string) (count uint64, err error) {
	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return count, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		count, err = ss.CountSegments(ctx, namespaceKey)
		return err
	})
}

func (s *Store) GetEvaluationRules(ctx context.Context, namespaceKey string, flagKey string) (rules []*storage.EvaluationRule, err error) {
	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return rules, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		rules, err = ss.GetEvaluationRules(ctx, namespaceKey, flagKey)
		return err
	})
}

func (s *Store) GetEvaluationDistributions(ctx context.Context, ruleID string) (dists []*storage.EvaluationDistribution, err error) {
	return dists, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		dists, err = ss.GetEvaluationDistributions(ctx, ruleID)
		return err
	})
}

func (s *Store) GetEvaluationRollouts(ctx context.Context, namespaceKey, flagKey string) (rollouts []*storage.EvaluationRollout, err error) {
	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return rollouts, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		rollouts, err = ss.GetEvaluationRollouts(ctx, namespaceKey, flagKey)
		return err
	})
}

func (s *Store) GetNamespace(ctx context.Context, key string) (ns *flipt.Namespace, err error) {
	if key == "" {
		key = flipt.DefaultNamespace
	}

	return ns, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		ns, err = ss.GetNamespace(ctx, key)
		return err
	})
}

func (s *Store) ListNamespaces(ctx context.Context, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Namespace], err error) {
	return set, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		set, err = ss.ListNamespaces(ctx, opts...)
		return err
	})
}

func (s *Store) CountNamespaces(ctx context.Context) (count uint64, err error) {
	return count, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		count, err = ss.CountNamespaces(ctx)
		return err
	})
}

func (s *Store) GetRollout(ctx context.Context, namespaceKey, id string) (rollout *flipt.Rollout, err error) {
	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return rollout, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		rollout, err = ss.GetRollout(ctx, namespaceKey, id)
		return err
	})
}

func (s *Store) ListRollouts(ctx context.Context, namespaceKey, flagKey string, opts ...storage.QueryOption) (set storage.ResultSet[*flipt.Rollout], err error) {
	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return set, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		set, err = ss.ListRollouts(ctx, namespaceKey, flagKey, opts...)
		return err
	})
}

func (s *Store) CountRollouts(ctx context.Context, namespaceKey, flagKey string) (count uint64, err error) {
	if namespaceKey == "" {
		namespaceKey = flipt.DefaultNamespace
	}

	return count, s.viewer.View(func(ss storage.ReadOnlyStore) error {
		count, err = ss.CountRollouts(ctx, namespaceKey, flagKey)
		return err
	})
}

// unimplemented write paths below

func (s *Store) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	return nil, ErrNotImplemented
}

func (s *Store) UpdateNamespace(ctx context.Context, r *flipt.UpdateNamespaceRequest) (*flipt.Namespace, error) {
	return nil, ErrNotImplemented
}

func (s *Store) DeleteNamespace(ctx context.Context, r *flipt.DeleteNamespaceRequest) error {
	return ErrNotImplemented
}

func (s *Store) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	return nil, ErrNotImplemented
}

func (s *Store) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	return nil, ErrNotImplemented
}

func (s *Store) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	return ErrNotImplemented
}

func (s *Store) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	return nil, ErrNotImplemented
}

func (s *Store) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	return nil, ErrNotImplemented
}

func (s *Store) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	return ErrNotImplemented
}

func (s *Store) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	return nil, ErrNotImplemented
}

func (s *Store) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	return nil, ErrNotImplemented
}

func (s *Store) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	return ErrNotImplemented
}

func (s *Store) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	return nil, ErrNotImplemented
}

func (s *Store) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	return nil, ErrNotImplemented
}

func (s *Store) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	return ErrNotImplemented
}

func (s *Store) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	return nil, ErrNotImplemented
}

func (s *Store) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	return nil, ErrNotImplemented
}

func (s *Store) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	return ErrNotImplemented
}

func (s *Store) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	return ErrNotImplemented
}

func (s *Store) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	return nil, ErrNotImplemented
}

func (s *Store) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	return nil, ErrNotImplemented
}

func (s *Store) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	return ErrNotImplemented
}

func (s *Store) CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (*flipt.Rollout, error) {
	return nil, ErrNotImplemented
}

func (s *Store) UpdateRollout(ctx context.Context, r *flipt.UpdateRolloutRequest) (*flipt.Rollout, error) {
	return nil, ErrNotImplemented
}

func (s *Store) DeleteRollout(ctx context.Context, r *flipt.DeleteRolloutRequest) error {
	return ErrNotImplemented
}

func (s *Store) OrderRollouts(ctx context.Context, r *flipt.OrderRolloutsRequest) error {
	return ErrNotImplemented
}
