package storage

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/database/sqlite3"
	"github.com/sirupsen/logrus"

	_ "github.com/golang-migrate/migrate/source/file"
)

var (
	logger *logrus.Logger
	debug  bool

	flagStore    FlagStore
	segmentStore SegmentStore
	ruleStore    RuleStore
)

const defaultTestDBURL = "file:../flipt_test.db"

func init() {
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.Parse()
}

func TestMain(m *testing.M) {
	// os.Exit skips defer calls
	// so we need to use another fn
	os.Exit(run(m))
}

func run(m *testing.M) int {
	logger = logrus.New()

	if debug {
		logger.SetLevel(logrus.DebugLevel)
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

	return m.Run()
}
