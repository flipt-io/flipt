package storage

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	sq "github.com/Masterminds/squirrel"
	migrate "github.com/golang-migrate/migrate"
	sqlite3 "github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db     *sql.DB
	logger *logrus.Logger
	debug  bool

	flagStore    FlagStore
	segmentStore SegmentStore
	ruleStore    RuleStore
)

const testDBPath = "../flipt_test.db"

func TestMain(m *testing.M) {
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.Parse()
	// os.Exit skips defer calls
	// so we need to use another fn
	os.Exit(run(m))
}

func run(m *testing.M) int {
	var err error

	logger = logrus.New()

	if debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	db, err = sql.Open("sqlite3", fmt.Sprintf("%s?_fk=true", testDBPath))
	if err != nil {
		logger.Fatal(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Fatal(err)
		}

		if err := os.Remove(testDBPath); err != nil {
			logger.Fatal(err)
		}
	}()

	driver, err := sqlite3.WithInstance(db, &sqlite3.Config{})
	if err != nil {
		logger.Fatal(err)
	}

	mm, err := migrate.NewWithDatabaseInstance("file://../config/migrations", "sqlite3", driver)
	if err != nil {
		logger.Fatal(err)
	}

	if err := mm.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Fatal(err)
	}

	var (
		builder = sq.StatementBuilder.RunWith(db)
		tx      = sq.NewStmtCacheProxy(db)
	)

	flagStore = NewFlagStorage(logger, builder)
	segmentStore = NewSegmentStorage(logger, builder)
	ruleStore = NewRuleStorage(logger, tx, builder)

	return m.Run()
}
