package storage

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	migrate "github.com/golang-migrate/migrate"
	sqlite3 "github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db     *sql.DB
	logger logrus.FieldLogger
	tables = []string{"constraints", "distributions", "flags", "rules", "segments", "variants"}

	flagRepo    FlagRepository
	segmentRepo SegmentRepository
	ruleRepo    RuleRepository
)

func TestMain(m *testing.M) {
	var err error

	logger = logrus.New()
	db, err = sql.Open("sqlite3", "../flipt_test.db?_fk=true")
	if err != nil {
		logger.Fatal(err)
	}

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

	flagRepo = NewFlagStorage(logger, db)
	segmentRepo = NewSegmentStorage(logger, db)
	ruleRepo = NewRuleStorage(logger, db)

	err = truncate()
	if err != nil {
		logger.Fatal(err)
	}

	defer func() {
		db.Close()
	}()

	code := m.Run()

	os.Exit(code)
}

func truncate() error {
	for _, table := range tables {
		truncate := fmt.Sprintf("DELETE FROM %s;", table)

		if _, err := db.Exec(truncate); err != nil {
			return err
		}

		if testing.Verbose() {
			logger.Println(truncate)
		}
	}

	return nil
}
