package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/database/sqlite3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/golang-migrate/migrate/source/file"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		url     string
		driver  Driver
		wantErr bool
	}{
		{
			name:   "sqlite",
			input:  "file:flipt.db",
			driver: SQLite,
			url:    "file:flipt.db?_fk=true&cache=shared",
		},
		{
			name:   "postres",
			input:  "postgres://postgres@localhost:5432/flipt?sslmode=disable",
			driver: Postgres,
			url:    "postgres://postgres@localhost:5432/flipt?sslmode=disable",
		},
		{
			name:    "invalid url",
			input:   "http://a b",
			wantErr: true,
		},
		{
			name:    "unknown driver ",
			input:   "mongo://127.0.0.1",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		var (
			input   = tt.input
			driver  = tt.driver
			url     = tt.url
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			d, u, err := parse(input)

			if wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, driver, d)
			assert.Equal(t, url, u.String())
		})
	}
}

var (
	logger *logrus.Logger

	flagStore    *FlagStorage
	segmentStore *SegmentStorage
	ruleStore    *RuleStorage
	evaluator    *EvaluatorStorage
)

const defaultTestDBURL = "file:../flipt_test.db"

func TestMain(m *testing.M) {
	// os.Exit skips defer calls
	// so we need to use another fn
	os.Exit(run(m))
}

func run(m *testing.M) int {
	logger = logrus.New()

	debug := os.Getenv("DEBUG")
	if debug == "" {
		logger.Level = logrus.DebugLevel
		logger.Out = ioutil.Discard
	}

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = defaultTestDBURL
	}

	db, driver, err := Open(dbURL)
	if err != nil {
		logger.Fatal(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Fatal(err)
		}
	}()

	var (
		dr      database.Driver
		builder sq.StatementBuilderType
		stmt    string

		tables = []string{"distributions", "rules", "constraints", "variants", "segments", "flags"}
	)

	switch driver {
	case SQLite:
		dr, err = sqlite3.WithInstance(db, &sqlite3.Config{})
		builder = sq.StatementBuilder.RunWith(sq.NewStmtCacher(db))

		stmt = "DELETE FROM %s"
	case Postgres:
		dr, err = postgres.WithInstance(db, &postgres.Config{})
		builder = sq.StatementBuilder.PlaceholderFormat(sq.Dollar).RunWith(sq.NewStmtCacher(db))

		stmt = "TRUNCATE TABLE %s CASCADE"
	}

	for _, t := range tables {
		_, _ = db.Exec(fmt.Sprintf(stmt, t))
	}

	if err != nil {
		logger.Fatal(err)
	}

	f := filepath.Clean(fmt.Sprintf("../config/migrations/%s", driver))

	mm, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", f), driver.String(), dr)
	if err != nil {
		logger.Fatal(err)
	}

	logger.Info("running migrations...")

	if err := mm.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Fatal(err)
	}

	flagStore = NewFlagStorage(logger, builder)
	segmentStore = NewSegmentStorage(logger, builder)
	ruleStore = NewRuleStorage(logger, builder, db)
	evaluator = NewEvaluatorStorage(logger, builder)

	return m.Run()
}
