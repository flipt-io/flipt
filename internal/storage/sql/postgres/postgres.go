package postgres

import (
	"context"
	"database/sql"

	"errors"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgconn"
	errs "go.flipt.io/flipt/errors"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/sql/common"
	flipt "go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
)

const (
	constraintForeignKeyErr = "23503" // "foreign_key_violation"
	constraintUniqueErr     = "23505" // "unique_violation"
)

var _ storage.Store = &Store{}

func NewStore(db *sql.DB, builder sq.StatementBuilderType, logger *zap.Logger) *Store {
	return &Store{
		Store: common.NewStore(db, builder.PlaceholderFormat(sq.Dollar), logger),
	}
}

type Store struct {
	*common.Store
}

func (s *Store) String() string {
	return "postgres"
}

func (s *Store) CreateNamespace(ctx context.Context, r *flipt.CreateNamespaceRequest) (*flipt.Namespace, error) {
	namespace, err := s.Store.CreateNamespace(ctx, r)

	if err != nil {
		var perr *pgconn.PgError

		if errors.As(err, &perr) && perr.Code == constraintUniqueErr {
			return nil, errs.ErrInvalidf(`namespace %q is not unique`, r.Key)
		}

		return nil, err
	}

	return namespace, nil
}

func (s *Store) CreateFlag(ctx context.Context, r *flipt.CreateFlagRequest) (*flipt.Flag, error) {
	flag, err := s.Store.CreateFlag(ctx, r)

	if err != nil {
		var perr *pgconn.PgError

		if errors.As(err, &perr) {
			switch perr.Code {
			case constraintForeignKeyErr:
				return nil, errs.ErrNotFoundf("namespace %q", r.NamespaceKey)
			case constraintUniqueErr:
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
		var perr *pgconn.PgError

		if errors.As(err, &perr) && perr.Code == constraintForeignKeyErr {
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
		var perr *pgconn.PgError

		if errors.As(err, &perr) {
			switch perr.Code {
			case constraintForeignKeyErr:
				return nil, errs.ErrNotFoundf(`flag "%s/%s"`, r.NamespaceKey, r.FlagKey)
			case constraintUniqueErr:
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
		var perr *pgconn.PgError

		if errors.As(err, &perr) && perr.Code == constraintUniqueErr {
			return nil, errs.ErrInvalidf(`variant %q is not unique for flag "%s/%s"`, r.Key, r.NamespaceKey, r.FlagKey)
		}

		return nil, err
	}

	return variant, nil
}

func (s *Store) CreateSegment(ctx context.Context, r *flipt.CreateSegmentRequest) (*flipt.Segment, error) {
	segment, err := s.Store.CreateSegment(ctx, r)

	if err != nil {
		var perr *pgconn.PgError

		if errors.As(err, &perr) {
			switch perr.Code {
			case constraintForeignKeyErr:
				return nil, errs.ErrNotFoundf("namespace %q", r.NamespaceKey)
			case constraintUniqueErr:
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
		var perr *pgconn.PgError

		if errors.As(err, &perr) && perr.Code == constraintForeignKeyErr {
			return nil, errs.ErrNotFoundf(`segment "%s/%s"`, r.NamespaceKey, r.SegmentKey)
		}

		return nil, err
	}

	return constraint, nil
}

func (s *Store) CreateRollout(ctx context.Context, r *flipt.CreateRolloutRequest) (*flipt.Rollout, error) {
	rollout, err := s.Store.CreateRollout(ctx, r)

	if err != nil {
		var perr *pgconn.PgError

		if errors.As(err, &perr) && perr.Code == constraintForeignKeyErr {
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
		var perr *pgconn.PgError

		if errors.As(err, &perr) && perr.Code == constraintForeignKeyErr {
			return nil, errs.ErrNotFoundf(`flag "%s/%s" or segment "%s/%s"`, r.NamespaceKey, r.FlagKey, r.NamespaceKey, r.SegmentKey)
		}

		return nil, err
	}

	return rule, nil
}

func (s *Store) UpdateRule(ctx context.Context, r *flipt.UpdateRuleRequest) (*flipt.Rule, error) {
	rule, err := s.Store.UpdateRule(ctx, r)

	if err != nil {
		var perr *pgconn.PgError

		if errors.As(err, &perr) && perr.Code == constraintForeignKeyErr {
			return nil, errs.ErrNotFoundf(`rule "%s/%s"`, r.NamespaceKey, r.Id)
		}

		return nil, err
	}

	return rule, nil
}

func (s *Store) CreateDistribution(ctx context.Context, r *flipt.CreateDistributionRequest) (*flipt.Distribution, error) {
	dist, err := s.Store.CreateDistribution(ctx, r)

	if err != nil {
		var perr *pgconn.PgError

		if errors.As(err, &perr) && perr.Code == constraintForeignKeyErr {
			return nil, errs.ErrNotFoundf("variant %q, rule %q, flag %q in namespace %q", r.VariantId, r.RuleId, r.FlagKey, r.NamespaceKey)
		}

		return nil, err
	}

	return dist, nil
}
