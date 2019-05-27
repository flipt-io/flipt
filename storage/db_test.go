// +build db

package storage

import (
	"flag"
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	_ "github.com/golang-migrate/migrate/source/file"
)

var (
	logger *logrus.Logger
	debug  bool
	dbURL  string

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

	db, err := Open(dbURL)
	if err != nil {
		logger.Fatal(err)
	}

	if err := db.truncate(); err != nil {
		logger.Fatal(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Fatal(err)
		}
	}()

	err = db.Migrate("../config/migrations/")
	if err != nil {
		logger.Fatal(err)
	}

	flagStore = NewFlagStorage(logger, db)
	segmentStore = NewSegmentStorage(logger, db)
	ruleStore = NewRuleStorage(logger, db)

	return m.Run()
}
