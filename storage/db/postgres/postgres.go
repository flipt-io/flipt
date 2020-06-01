package postgres

import (
	"context"
	"database/sql"

	"errors"

	"github.com/lib/pq"
	errs "github.com/markphelps/flipt/errors"

	sq "github.com/Masterminds/squirrel"
)

const (
	pgConstraintForeignKey = "foreign_key_violation"
	pgConstraintUnique     = "unique_violation"
)

type Driver struct{}

func (d *Driver) Create(ctx context.Context, query sq.InsertBuilder) error {
	var perr pq.Error

	if _, err := query.ExecContext(ctx); err != nil {
		if errors.As(err, &perr) {
			switch perr.Code.Name() {
			case pgConstraintUnique:
				return errs.ErrUniqueViolation
			case pgConstraintForeignKey:
				return errs.ErrForeignKeyViolation
			}
		}

		return err
	}

	return nil
}

func (d *Driver) Update(ctx context.Context, query sq.UpdateBuilder) (sql.Result, error) {
	var perr pq.Error

	res, err := query.ExecContext(ctx)
	if err != nil {
		if errors.As(err, &perr) {
			switch perr.Code.Name() {
			case pgConstraintUnique:
				return nil, errs.ErrUniqueViolation
			}
		}

		return nil, err
	}

	return res, nil
}
