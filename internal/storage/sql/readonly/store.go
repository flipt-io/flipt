package readonly

import (
	"context"
	"fmt"

	"go.flipt.io/flipt/internal/storage"
	flipt "go.flipt.io/flipt/rpc/flipt"
)

var _ storage.Store = (*ReadOnlyStore)(nil)

// ReadOnlyStore is a wrapper around the storage.Store that prevents write operations for sql storage.
type ReadOnlyStore struct {
	storage.Store
}

func NewStore(store storage.Store) *ReadOnlyStore {
	return &ReadOnlyStore{Store: store}
}

func (s *ReadOnlyStore) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) UpdateNamespace(ctx context.Context, r *flipt.UpdateNamespaceRequest) (*flipt.Namespace, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) DeleteNamespace(ctx context.Context, r *flipt.DeleteNamespaceRequest) error {
	return storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) DeleteFlag(ctx context.Context, r *flipt.DeleteFlagRequest) error {
	return storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) DeleteVariant(ctx context.Context, r *flipt.DeleteVariantRequest) error {
	return storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) UpdateSegment(ctx context.Context, r *flipt.UpdateSegmentRequest) (*flipt.Segment, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) DeleteSegment(ctx context.Context, r *flipt.DeleteSegmentRequest) error {
	return storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) UpdateConstraint(ctx context.Context, r *flipt.UpdateConstraintRequest) (*flipt.Constraint, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) DeleteConstraint(ctx context.Context, r *flipt.DeleteConstraintRequest) error {
	return storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) DeleteRule(ctx context.Context, r *flipt.DeleteRuleRequest) error {
	return storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) OrderRules(ctx context.Context, r *flipt.OrderRulesRequest) error {
	return storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) UpdateDistribution(ctx context.Context, r *flipt.UpdateDistributionRequest) (*flipt.Distribution, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) DeleteDistribution(ctx context.Context, r *flipt.DeleteDistributionRequest) error {
	return storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (*flipt.Rollout, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) UpdateRollout(ctx context.Context, r *flipt.UpdateRolloutRequest) (*flipt.Rollout, error) {
	return nil, storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) DeleteRollout(ctx context.Context, r *flipt.DeleteRolloutRequest) error {
	return storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) OrderRollouts(ctx context.Context, r *flipt.OrderRolloutsRequest) error {
	return storage.ErrReadOnlyStore
}

func (s *ReadOnlyStore) String() string {
	return fmt.Sprintf("readonly:%s", s.Store.String())
}
