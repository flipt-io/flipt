package storage

import (
	"flag"
	"os"
	"testing"

	"github.com/sirupsen/logrus"

	_ "github.com/golang-migrate/migrate/source/file"
	_ "github.com/mattn/go-sqlite3"
)

var (
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
	logger = logrus.New()

	if debug {
		logger.SetLevel(logrus.DebugLevel)
	}

	db, err := Open("sqlite3://" + testDBPath)
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

	err = db.Migrate("../config/migrations/")
	if err != nil {
		logger.Fatal(err)
	}

	flagStore = NewFlagStorage(logger, db)
	segmentStore = NewSegmentStorage(logger, db)
	ruleStore = NewRuleStorage(logger, db)

	return m.Run()
}
