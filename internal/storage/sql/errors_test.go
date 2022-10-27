package sql

import (
	"database/sql"
	"fmt"
	"testing"

	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"
	"github.com/mattn/go-sqlite3"
	"github.com/stretchr/testify/require"
	"go.flipt.io/flipt/internal/storage"
)

func Test_AdaptError(t *testing.T) {
	for _, test := range []struct {
		driver   Driver
		inputErr error
		// if outputErrIs nil then test will ensure input is returned
		outputErrIs error
	}{
		// No driver
		{},
		// All drivers
		{
			driver:      SQLite,
			inputErr:    sql.ErrNoRows,
			outputErrIs: storage.ErrNotFound,
		},
		{
			driver:      SQLite,
			inputErr:    fmt.Errorf("wrapped no rows: %w", sql.ErrNoRows),
			outputErrIs: storage.ErrNotFound,
		},
		{
			driver:      Postgres,
			inputErr:    sql.ErrNoRows,
			outputErrIs: storage.ErrNotFound,
		},
		{
			driver:      Postgres,
			inputErr:    fmt.Errorf("wrapped no rows: %w", sql.ErrNoRows),
			outputErrIs: storage.ErrNotFound,
		},
		{
			driver:      MySQL,
			inputErr:    sql.ErrNoRows,
			outputErrIs: storage.ErrNotFound,
		},
		{
			driver:      MySQL,
			inputErr:    fmt.Errorf("wrapped no rows: %w", sql.ErrNoRows),
			outputErrIs: storage.ErrNotFound,
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
			outputErrIs: storage.ErrInvalid,
		},
		{
			driver: SQLite,
			inputErr: sqlite3.Error{
				Code:         sqlite3.ErrConstraint,
				ExtendedCode: sqlite3.ErrConstraintForeignKey,
			},
			outputErrIs: storage.ErrNotFound,
		},
		{
			driver: SQLite,
			inputErr: sqlite3.Error{
				Code:         sqlite3.ErrConstraint,
				ExtendedCode: sqlite3.ErrConstraintUnique,
			},
			outputErrIs: storage.ErrInvalid,
		},
		// Postgres
		// Unchanged errors
		{driver: Postgres},
		{
			driver:   Postgres,
			inputErr: &pq.Error{},
		},
		{
			driver:   Postgres,
			inputErr: &pq.Error{Code: pq.ErrorCode("01000")},
		},
		// Adjusted errors
		{
			driver: Postgres,
			// foreign_key_violation
			inputErr:    &pq.Error{Code: pq.ErrorCode("23503")},
			outputErrIs: storage.ErrNotFound,
		},
		{
			driver: Postgres,
			// unique_violation
			inputErr:    &pq.Error{Code: pq.ErrorCode("23505")},
			outputErrIs: storage.ErrInvalid,
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
			outputErrIs: storage.ErrNotFound,
		},
		{
			driver: MySQL,
			// unique_violation
			inputErr:    &mysql.MySQLError{Number: uint16(1062)},
			outputErrIs: storage.ErrInvalid,
		},
	} {
		test := test
		name := fmt.Sprintf("(%v).AdaptError(%T) == %T", test.driver, test.inputErr, test.outputErrIs)
		t.Run(name, func(t *testing.T) {
			err := test.driver.AdaptError(test.inputErr)
			if test.outputErrIs == nil {
				// given the output expectation is nil we ensure the input error
				// is returned unchanged
				require.Equal(t, test.inputErr, err, "input error was changed unexpectedly")
				return
			}

			// otherwise, we ensure returned error matches via errors.Is
			require.ErrorIs(t, err, test.outputErrIs)
		})
	}
}
