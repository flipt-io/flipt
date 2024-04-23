package sql

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/errors"
)

func errPtr[E error](e E) *E {
	return &e
}

func Test_AdaptError(t *testing.T) {
	for _, test := range []struct {
		driver   Driver
		inputErr error
		// if outputErrAs nil then test will ensure input is returned
		outputErrAs any
	}{
		// No driver
		{},
		// All drivers
		{
			driver:      SQLite,
			inputErr:    sql.ErrNoRows,
			outputErrAs: errPtr(errors.ErrNotFound("")),
		},
		{
			driver:      SQLite,
			inputErr:    fmt.Errorf("wrapped no rows: %w", sql.ErrNoRows),
			outputErrAs: errPtr(errors.ErrNotFound("")),
		},
		{
			driver:      Postgres,
			inputErr:    sql.ErrNoRows,
			outputErrAs: errPtr(errors.ErrNotFound("")),
		},
		{
			driver:      Postgres,
			inputErr:    fmt.Errorf("wrapped no rows: %w", sql.ErrNoRows),
			outputErrAs: errPtr(errors.ErrNotFound("")),
		},
		{
			driver:      MySQL,
			inputErr:    sql.ErrNoRows,
			outputErrAs: errPtr(errors.ErrNotFound("")),
		},
		{
			driver:      MySQL,
			inputErr:    fmt.Errorf("wrapped no rows: %w", sql.ErrNoRows),
			outputErrAs: errPtr(errors.ErrNotFound("")),
		},
		// SQLite
		// Unchanged errors
		{driver: SQLite},
		{
			driver:   SQLite,
			inputErr: sqlite3.Error{},
		},
		{
			driver:   SQLite,
			inputErr: sqlite3.Error{Code: sqlite3.ErrTooBig},
		},
		// Adjusted errors
		{
			driver: SQLite,
			inputErr: sqlite3.Error{
				Code:         sqlite3.ErrConstraint,
				ExtendedCode: sqlite3.ErrConstraintCheck,
			},
			outputErrAs: errPtr(errors.ErrInvalid("")),
		},
		{
			driver: SQLite,
			inputErr: sqlite3.Error{
				Code:         sqlite3.ErrConstraint,
				ExtendedCode: sqlite3.ErrConstraintForeignKey,
			},
			outputErrAs: errPtr(errors.ErrNotFound("")),
		},
		{
			driver: SQLite,
			inputErr: sqlite3.Error{
				Code:         sqlite3.ErrConstraint,
				ExtendedCode: sqlite3.ErrConstraintUnique,
			},
			outputErrAs: errPtr(errors.ErrInvalid("")),
		},
		// Postgres
		// Unchanged errors
		{driver: Postgres},
		{
			driver:   Postgres,
			inputErr: &pgconn.PgError{},
		},
		{
			driver:   Postgres,
			inputErr: &pgconn.PgError{Code: "01000"},
		},
		// Adjusted errors
		{
			driver: Postgres,
			// foreign_key_violation
			inputErr:    &pgconn.PgError{Code: "23503"},
			outputErrAs: errPtr(errors.ErrNotFound("")),
		},
		{
			driver: Postgres,
			// unique_violation
			inputErr:    &pgconn.PgError{Code: "23505"},
			outputErrAs: errPtr(errors.ErrInvalid("")),
		},
		{
			driver: Postgres,
			// connection error
			inputErr:    &pgconn.ConnectError{},
			outputErrAs: errPtr(errConnectionFailed),
		},
		// MySQL
		// Unchanged errors
		{driver: MySQL},
		{
			driver:   MySQL,
			inputErr: &mysql.MySQLError{},
		},
		{
			driver:   MySQL,
			inputErr: &mysql.MySQLError{Number: uint16(1000)},
		},
		// Adjusted errors
		{
			driver: MySQL,
			// foreign_key_violation
			inputErr:    &mysql.MySQLError{Number: uint16(1452)},
			outputErrAs: errPtr(errors.ErrNotFound("")),
		},
		{
			driver: MySQL,
			// unique_violation
			inputErr:    &mysql.MySQLError{Number: uint16(1062)},
			outputErrAs: errPtr(errors.ErrInvalid("")),
		},
	} {
		test := test

		outputs := test.outputErrAs
		if outputs == nil {
			outputs = test.inputErr
		}

		name := fmt.Sprintf("(%v).AdaptError(%v) == %T", test.driver, test.inputErr, outputs)

		t.Run(name, func(t *testing.T) {
			err := test.driver.AdaptError(test.inputErr)
			if test.outputErrAs == nil {
				// given the output expectation is nil we ensure the input error
				// is returned unchanged
				require.Equal(t, test.inputErr, err, "input error was changed unexpectedly")
				return
			}

			// otherwise, we ensure returned error matches via errors.Is
			require.ErrorAs(t, err, test.outputErrAs)
		})
	}
}
