package sqlite

import (
	"context"
	"database/sql"

	"errors"

	sq "github.com/Masterminds/squirrel"
	errs "github.com/markphelps/flipt/errors"
	flipt "github.com/markphelps/flipt/rpc/flipt"
	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/storage/sql/common"
	"github.com/mattn/go-sqlite3"
)

var _ storage.Store = &Store{}

// NewStore creates a new sqlite.Store
func NewStore(db *sql.DB) *Store {
	builder := sq.StatementBuilder.RunWith(sq.NewStmtCacher(db))

	return &Store{
		Store: common.NewStore(db, builder),
	}
}

// Store is a sqlite specific implementation of storage.Store
type Store struct {
	*common.Store
}

func (s *Store) String() string {
	return "sqlite"
}

func (s *Store) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	flag, err := s.Store.CreateFlag(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.Code == sqlite3.ErrConstraint {
			return nil, errs.ErrInvalidf("flag %q is not unique", r.Key)
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
				return nil, errs.ErrNotFoundf("flag %q", r.FlagKey)
			case sqlite3.ErrConstraintUnique:
				return nil, errs.ErrInvalidf("variant %q is not unique", r.Key)
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
			return nil, errs.ErrInvalidf("variant %q is not unique", r.Key)
		}

		return nil, err
	}

	return variant, nil
}

func (s *Store) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	segment, err := s.Store.CreateSegment(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.Code == sqlite3.ErrConstraint {
			return nil, errs.ErrInvalidf("segment %q is not unique", r.Key)
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
			return nil, errs.ErrNotFoundf("segment %q", r.SegmentKey)
		}

		return nil, err
	}

	return constraint, nil
}

func (s *Store) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	rule, err := s.Store.CreateRule(ctx, r)

	if err != nil {
		var serr sqlite3.Error

		if errors.As(err, &serr) && serr.Code == sqlite3.ErrConstraint {
			return nil, errs.ErrNotFoundf("flag %q or segment %q", r.FlagKey, r.SegmentKey)
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
			return nil, errs.ErrNotFoundf("rule %q", r.RuleId)
		}

		return nil, err
	}

	return dist, nil
}
