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
	dbURL  string

	flagStore    FlagStore
	segmentStore SegmentStore
	ruleStore    RuleStore
)

const (
	testDBPath       = "../flipt_test.db"
	defaultTestDBURL = "file:" + testDBPath
)

func init() {
	flag.BoolVar(&debug, "debug", false, "enable debug logging")
	flag.StringVar(&dbURL, "db", defaultTestDBURL, "url for db")
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

	db, err := Open(dbURL)
	if err != nil {
		logger.Fatal(err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			logger.Fatal(err)
		}

		if err := os.Remove(testDBPath); !os.IsNotExist(err) && err != nil {
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
