package sqlite

import (
	"context"
	"database/sql"

	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/mattn/go-sqlite3"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/sql/common"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
)

var _ storage.Store = &Store{}

// NewStore creates a new sqlite.Store
func NewStore(db *sql.DB, builder sq.StatementBuilderType, logger *zap.Logger) *Store {
	return &Store{
		Store: common.NewStore(db, builder, logger),
	}
}

// Store is a sqlite specific implementation of storage.Store
type Store struct {
	*common.Store
}

func (s *Store) String() string {
	return "sqlite"
}

func (s *Store) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	namespace, err := s.Store.CreateNamespace(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.ExtendedCode == sqlite3.ErrConstraintPrimaryKey {
			return nil, errs.ErrInvalidf(`namespace "%s" is not unique`, r.Key)
		}

		return nil, err
	}

	return namespace, nil
}

func (s *Store) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	flag, err := s.Store.CreateFlag(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) {
			switch serr.ExtendedCode {
			case sqlite3.ErrConstraintForeignKey:
				return nil, errs.ErrNotFoundf("namespace %q", r.NamespaceKey)
			case sqlite3.ErrConstraintPrimaryKey:
				return nil, errs.ErrInvalidf(`flag "%s/%s" is not unique`, r.NamespaceKey, r.Key)
			}
		}

		return nil, err
	}

	return flag, nil
}

func (s *Store) UpdateFlag(ctx context.Context, r *flipt.UpdateFlagRequest) (*flipt.Flag, error) {
	flag, err := s.Store.UpdateFlag(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.ExtendedCode == sqlite3.ErrConstraintForeignKey {
			if r.DefaultVariantId != "" {
				return nil, errs.ErrInvalidf(`variant %q not found for flag "%s/%s"`, r.DefaultVariantId, r.NamespaceKey, r.Key)
			}

			return nil, errs.ErrInvalidf(`flag "%s/%s" not found`, r.NamespaceKey, r.Key)
		}

		return nil, err
	}

	return flag, nil
}

func (s *Store) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	variant, err := s.Store.CreateVariant(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) {
			switch serr.ExtendedCode {
			case sqlite3.ErrConstraintForeignKey:
				return nil, errs.ErrNotFoundf(`flag "%s/%s"`, r.NamespaceKey, r.FlagKey)
			case sqlite3.ErrConstraintUnique:
				return nil, errs.ErrInvalidf(`variant %q is not unique for flag "%s/%s"`, r.Key, r.NamespaceKey, r.FlagKey)
			}
		}

		return nil, err
	}

	return variant, nil
}

func (s *Store) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	variant, err := s.Store.UpdateVariant(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.Code == sqlite3.ErrConstraint {
			return nil, errs.ErrInvalidf(`variant %q is not unique for flag "%s/%s"`, r.Key, r.NamespaceKey, r.FlagKey)
		}

		return nil, err
	}

	return variant, nil
}

func (s *Store) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	segment, err := s.Store.CreateSegment(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) {
			switch serr.ExtendedCode {
			case sqlite3.ErrConstraintForeignKey:
				return nil, errs.ErrNotFoundf("namespace %q", r.NamespaceKey)
			case sqlite3.ErrConstraintPrimaryKey:
				return nil, errs.ErrInvalidf(`segment "%s/%s" is not unique`, r.NamespaceKey, r.Key)
			}
		}

		return nil, err
	}

	return segment, nil
}

func (s *Store) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	constraint, err := s.Store.CreateConstraint(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.Code == sqlite3.ErrConstraint {
			return nil, errs.ErrNotFoundf(`segment "%s/%s"`, r.NamespaceKey, r.SegmentKey)
		}

		return nil, err
	}

	return constraint, nil
}

func (s *Store) CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (*flipt.Rollout, error) {
	rollout, err := s.Store.CreateRollout(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.Code == sqlite3.ErrConstraint {
			if segment := r.GetSegment(); segment != nil {
				return nil, errs.ErrNotFoundf(`flag "%s/%s or segment %s"`, r.NamespaceKey, r.FlagKey, segment.SegmentKey)
			}
			return nil, errs.ErrNotFoundf(`flag "%s/%s"`, r.NamespaceKey, r.FlagKey)
		}

		return nil, err
	}

	return rollout, nil
}

func (s *Store) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	rule, err := s.Store.CreateRule(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.Code == sqlite3.ErrConstraint {
			return nil, errs.ErrNotFoundf(`flag "%s/%s" or segment "%s/%s"`, r.NamespaceKey, r.FlagKey, r.NamespaceKey, r.SegmentKey)
		}

		return nil, err
	}

	return rule, nil
}

func (s *Store) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	rule, err := s.Store.UpdateRule(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.Code == sqlite3.ErrConstraint {
			return nil, errs.ErrNotFoundf(`rule "%s/%s"`, r.NamespaceKey, r.Id)
		}

		return nil, err
	}

	return rule, nil
}

func (s *Store) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	dist, err := s.Store.CreateDistribution(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.Code == sqlite3.ErrConstraint {
			return nil, errs.ErrNotFoundf("variant %q, rule %q, flag %q in namespace %q", r.VariantId, r.RuleId, r.FlagKey, r.NamespaceKey)
		}

		return nil, err
	}

	return dist, nil
}
