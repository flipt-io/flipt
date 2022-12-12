package postgres

import (
	"context"
	"database/sql"

	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/lib/pq"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/sql/common"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
)

const (
	constraintForeignKeyErr = "foreign_key_violation"
	constraintUniqueErr     = "unique_violation"
)

var _ storage.Store = &Store{}

func NewStore(db *sql.DB, logger *zap.Logger) *Store {
	builder := sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(sq.NewStmtCacher(db))

	return &Store{
		Store: common.NewStore(db, builder, logger),
	}
}

type Store struct {
	*common.Store
}

func (s *Store) String() string {
	return "postgres"
}

func (s *Store) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	flag, err := s.Store.CreateFlag(ctx, r)

	if err != nil {
		var perr *pq.Error

		if errors.As(err, &perr) && perr.Code.Name() == constraintUniqueErr {
			return nil, errs.NewErrorf[errs.ErrInvalid]("flag %q is not unique", r.Key)
		}

		return nil, err
	}

	return flag, nil
}

func (s *Store) CreateVariant(ctx context.Context, r *flipt.CreateVariantRequest) (*flipt.Variant, error) {
	variant, err := s.Store.CreateVariant(ctx, r)

	if err != nil {
		var perr *pq.Error

		if errors.As(err, &perr) {
			switch perr.Code.Name() {
			case constraintForeignKeyErr:
				return nil, errs.NewErrorf[errs.ErrNotFound]("flag %q", r.FlagKey)
			case constraintUniqueErr:
				return nil, errs.NewErrorf[errs.ErrInvalid]("variant %q is not unique", r.Key)
			}
		}

		return nil, err
	}

	return variant, nil
}

func (s *Store) UpdateVariant(ctx context.Context, r *flipt.UpdateVariantRequest) (*flipt.Variant, error) {
	variant, err := s.Store.UpdateVariant(ctx, r)

	if err != nil {
		var perr *pq.Error

		if errors.As(err, &perr) && perr.Code.Name() == constraintUniqueErr {
			return nil, errs.NewErrorf[errs.ErrInvalid]("variant %q is not unique", r.Key)
		}

		return nil, err
	}

	return variant, nil
}

func (s *Store) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	segment, err := s.Store.CreateSegment(ctx, r)

	if err != nil {
		var perr *pq.Error

		if errors.As(err, &perr) && perr.Code.Name() == constraintUniqueErr {
			return nil, errs.NewErrorf[errs.ErrInvalid]("segment %q is not unique", r.Key)
		}

		return nil, err
	}

	return segment, nil
}

func (s *Store) CreateConstraint(ctx context.Context, r *flipt.CreateConstraintRequest) (*flipt.Constraint, error) {
	constraint, err := s.Store.CreateConstraint(ctx, r)

	if err != nil {
		var perr *pq.Error

		if errors.As(err, &perr) && perr.Code.Name() == constraintForeignKeyErr {
			return nil, errs.NewErrorf[errs.ErrNotFound]("segment %q", r.SegmentKey)
		}

		return nil, err
	}

	return constraint, nil
}

func (s *Store) CreateRule(ctx context.Context, r *flipt.CreateRuleRequest) (*flipt.Rule, error) {
	rule, err := s.Store.CreateRule(ctx, r)

	if err != nil {
		var perr *pq.Error

		if errors.As(err, &perr) && perr.Code.Name() == constraintForeignKeyErr {
			return nil, errs.NewErrorf[errs.ErrNotFound]("flag %q or segment %q", r.FlagKey, r.SegmentKey)
		}

		return nil, err
	}

	return rule, nil
}

func (s *Store) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	dist, err := s.Store.CreateDistribution(ctx, r)

	if err != nil {
		var perr *pq.Error

		if errors.As(err, &perr) && perr.Code.Name() == constraintForeignKeyErr {
			return nil, errs.NewErrorf[errs.ErrNotFound]("rule %q", r.RuleId)
		}

		return nil, err
	}

	return dist, nil
}
