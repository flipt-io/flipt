package fs

import (
	"context"
	"errors"
	"fmt"
	"path"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
)

var (
	_ storage.Store = (*Store)(nil)

	// ErrNotImplemented is returned when a method has intentionally not been implemented
	// This is usually reserved for the store write actions when the store is read-only
	// but still needs to implement storage.Store
	ErrNotImplemented = errors.New("not implemented")
)

// ReferencedSnapshotStore is a type which has a single function View.
// View is a functional transaction interface for reading a snapshot
// during the lifetime of a supplied function.
type ReferencedSnapshotStore interface {
	// View accepts a function which takes a storage.ReadOnlyStore (*StoreSnapshot).
	// The ReferencedSnapshotStore will supply a snapshot which is valid
	// for the lifetime of the provided function call.
	// It expects a reference to be supplied on calls to View.
	// Empty reference will be regarded as the default reference by the store.
	View(context.Context, storage.Reference, func(storage.ReadOnlyStore) error) error
	fmt.Stringer
}

// SnapshotStore is a type which has a single function View.
// View is a functional transaction interface for reading a snapshot
// during the lifetime of a supplied function.
type SnapshotStore interface {
	// View accepts a function which takes a storage.ReadOnlyStore (*SnapshotStore).
	// The SnapshotStore will supply a snapshot which is valid
	// for the lifetime of the provided function call.
	View(context.Context, func(storage.ReadOnlyStore) error) error
	fmt.Stringer
}

// SingleReferenceSnapshotStore implements ReferencedSnapshotStore but delegates to a SnapshotStore implementation.
// Calls with View with a non-empty reference log a warning with the requested reference.
type SingleReferenceSnapshotStore struct {
	SnapshotStore

	logger *zap.Logger
}

// NewSingleReferenceStore adapts a SnapshotStore implementation into
// a ReferencedSnapshotStore implementation which expects to be called with
// an empty reference.
func NewSingleReferenceStore(logger *zap.Logger, s SnapshotStore) *SingleReferenceSnapshotStore {
	return &SingleReferenceSnapshotStore{s, logger}
}

// View delegates to the embedded snapshot store implementation and drops the reference.
// It logs a warning if a non-empty reference is supplied as the wrapped store
// isn't intended for use with custom references.
func (s *SingleReferenceSnapshotStore) View(ctx context.Context, ref storage.Reference, fn func(storage.ReadOnlyStore) error) error {
	if ref != "" {
		s.logger.Warn("single reference store called with non-empty reference", zap.String("reference", string(ref)))
	}

	return s.SnapshotStore.View(ctx, fn)
}

// Store embeds a StoreSnapshot and wraps the Store methods with a read-write mutex
// to synchronize reads with atomic replacements of the embedded snapshot.
type Store struct {
	viewer ReferencedSnapshotStore
}

func NewStore(viewer ReferencedSnapshotStore) *Store {
	return &Store{viewer: viewer}
}

func (s *Store) String() string {
	return path.Join("declarative", s.viewer.String())
}

func (s *Store) GetFlag(ctx context.Context, req storage.ResourceRequest) (flag *flipt.Flag, err error) {
	return flag, s.viewer.View(ctx, req.Reference, func(ss storage.ReadOnlyStore) error {
		flag, err = ss.GetFlag(ctx, req)
		return err
	})
}

func (s *Store) ListFlags(ctx context.Context, req *storage.ListRequest[storage.NamespaceRequest]) (set storage.ResultSet[*flipt.Flag], err error) {
	return set, s.viewer.View(ctx, req.Predicate.Reference, func(ss storage.ReadOnlyStore) error {
		set, err = ss.ListFlags(ctx, req)
		return err
	})
}

func (s *Store) CountFlags(ctx context.Context, p storage.NamespaceRequest) (count uint64, err error) {
	return count, s.viewer.View(ctx, p.Reference, func(ss storage.ReadOnlyStore) error {
		count, err = ss.CountFlags(ctx, p)
		return err
	})
}

func (s *Store) GetRule(ctx context.Context, p storage.NamespaceRequest, id string) (rule *flipt.Rule, err error) {
	return rule, s.viewer.View(ctx, p.Reference, func(ss storage.ReadOnlyStore) error {
		rule, err = ss.GetRule(ctx, p, id)
		return err
	})
}

func (s *Store) ListRules(ctx context.Context, req *storage.ListRequest[storage.ResourceRequest]) (set storage.ResultSet[*flipt.Rule], err error) {
	return set, s.viewer.View(ctx, req.Predicate.Reference, func(ss storage.ReadOnlyStore) error {
		set, err = ss.ListRules(ctx, req)
		return err
	})
}

func (s *Store) CountRules(ctx context.Context, flag storage.ResourceRequest) (count uint64, err error) {
	return count, s.viewer.View(ctx, flag.Reference, func(ss storage.ReadOnlyStore) error {
		count, err = ss.CountRules(ctx, flag)
		return err
	})
}

func (s *Store) GetSegment(ctx context.Context, p storage.ResourceRequest) (segment *flipt.Segment, err error) {
	return segment, s.viewer.View(ctx, p.Reference, func(ss storage.ReadOnlyStore) error {
		segment, err = ss.GetSegment(ctx, p)
		return err
	})
}

func (s *Store) ListSegments(ctx context.Context, req *storage.ListRequest[storage.NamespaceRequest]) (set storage.ResultSet[*flipt.Segment], err error) {
	return set, s.viewer.View(ctx, req.Predicate.Reference, func(ss storage.ReadOnlyStore) error {
		set, err = ss.ListSegments(ctx, req)
		return err
	})
}

func (s *Store) CountSegments(ctx context.Context, p storage.NamespaceRequest) (count uint64, err error) {
	return count, s.viewer.View(ctx, p.Reference, func(ss storage.ReadOnlyStore) error {
		count, err = ss.CountSegments(ctx, p)
		return err
	})
}

func (s *Store) GetEvaluationRules(ctx context.Context, flag storage.ResourceRequest) (rules []*storage.EvaluationRule, err error) {
	return rules, s.viewer.View(ctx, flag.Reference, func(ss storage.ReadOnlyStore) error {
		rules, err = ss.GetEvaluationRules(ctx, flag)
		return err
	})
}

func (s *Store) GetEvaluationDistributions(ctx context.Context, rule storage.IDRequest) (dists []*storage.EvaluationDistribution, err error) {
	return dists, s.viewer.View(ctx, rule.Reference, func(ss storage.ReadOnlyStore) error {
		dists, err = ss.GetEvaluationDistributions(ctx, rule)
		return err
	})
}

func (s *Store) GetEvaluationRollouts(ctx context.Context, flag storage.ResourceRequest) (rollouts []*storage.EvaluationRollout, err error) {
	return rollouts, s.viewer.View(ctx, flag.Reference, func(ss storage.ReadOnlyStore) error {
		rollouts, err = ss.GetEvaluationRollouts(ctx, flag)
		return err
	})
}

func (s *Store) GetNamespace(ctx context.Context, p storage.NamespaceRequest) (ns *flipt.Namespace, err error) {
	return ns, s.viewer.View(ctx, p.Reference, func(ss storage.ReadOnlyStore) error {
		ns, err = ss.GetNamespace(ctx, p)
		return err
	})
}

func (s *Store) ListNamespaces(ctx context.Context, req *storage.ListRequest[storage.ReferenceRequest]) (set storage.ResultSet[*flipt.Namespace], err error) {
	return set, s.viewer.View(ctx, req.Predicate.Reference, func(ss storage.ReadOnlyStore) error {
		set, err = ss.ListNamespaces(ctx, req)
		return err
	})
}

func (s *Store) CountNamespaces(ctx context.Context, req storage.ReferenceRequest) (count uint64, err error) {
	return count, s.viewer.View(ctx, req.Reference, func(ss storage.ReadOnlyStore) error {
		count, err = ss.CountNamespaces(ctx, req)
		return err
	})
}

func (s *Store) GetRollout(ctx context.Context, p storage.NamespaceRequest, id string) (rollout *flipt.Rollout, err error) {
	return rollout, s.viewer.View(ctx, p.Reference, func(ss storage.ReadOnlyStore) error {
		rollout, err = ss.GetRollout(ctx, p, id)
		return err
	})
}

func (s *Store) ListRollouts(ctx context.Context, req *storage.ListRequest[storage.ResourceRequest]) (set storage.ResultSet[*flipt.Rollout], err error) {
	return set, s.viewer.View(ctx, req.Predicate.Reference, func(ss storage.ReadOnlyStore) error {
		set, err = ss.ListRollouts(ctx, req)
		return err
	})
}

func (s *Store) CountRollouts(ctx context.Context, flag storage.ResourceRequest) (count uint64, err error) {
	return count, s.viewer.View(ctx, flag.Reference, func(ss storage.ReadOnlyStore) error {
		count, err = ss.CountRollouts(ctx, flag)
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

func (s *Store) GetVersion(ctx context.Context, ns storage.NamespaceRequest) (version string, err error) {
	return version, s.viewer.View(ctx, ns.Reference, func(ss storage.ReadOnlyStore) error {
		version, err = ss.GetVersion(ctx, ns)
		return err
	})
}
