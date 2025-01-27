package fs

import (
	"context"
	"errors"
	"fmt"
	"path"

	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/rpc/flipt/core"
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

func (s *Store) GetFlag(ctx context.Context, req storage.ResourceRequest) (flag *core.Flag, err error) {
	return flag, s.viewer.View(ctx, req.Reference, func(ss storage.ReadOnlyStore) error {
		flag, err = ss.GetFlag(ctx, req)
		return err
	})
}

func (s *Store) ListFlags(ctx context.Context, req *storage.ListRequest[storage.NamespaceRequest]) (set storage.ResultSet[*core.Flag], err error) {
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

func (s *Store) GetEvaluationRules(ctx context.Context, flag storage.ResourceRequest) (rules []*storage.EvaluationRule, err error) {
	return rules, s.viewer.View(ctx, flag.Reference, func(ss storage.ReadOnlyStore) error {
		rules, err = ss.GetEvaluationRules(ctx, flag)
		return err
	})
}

func (s *Store) GetEvaluationDistributions(ctx context.Context, r storage.ResourceRequest, rule storage.IDRequest) (dists []*storage.EvaluationDistribution, err error) {
	return dists, s.viewer.View(ctx, rule.Reference, func(ss storage.ReadOnlyStore) error {
		dists, err = ss.GetEvaluationDistributions(ctx, r, rule)
		return err
	})
}

func (s *Store) GetEvaluationRollouts(ctx context.Context, flag storage.ResourceRequest) (rollouts []*storage.EvaluationRollout, err error) {
	return rollouts, s.viewer.View(ctx, flag.Reference, func(ss storage.ReadOnlyStore) error {
		rollouts, err = ss.GetEvaluationRollouts(ctx, flag)
		return err
	})
}
