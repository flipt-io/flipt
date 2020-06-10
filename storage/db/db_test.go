package db

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/mysql"
	pg "github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/database/sqlite3"
	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/storage/db/postgres"
	"github.com/markphelps/flipt/storage/db/sqlite"
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
			name:   "mysql",
			input:  "mysql://mysql@localhost:3306/flipt",
			driver: MySQL,
			url:    "mysql://mysql@localhost:3306/flipt?multiStatements=true",
		},
		{
			name:    "invalid url",
			input:   "http://a b",
			wantErr: true,
		},
		{
			name:    "unknown driver",
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

var store storage.Store

const defaultTestDBURL = "file:../../flipt_test.db"

func TestMain(m *testing.M) {
	// os.Exit skips defer calls
	// so we need to use another fn
	os.Exit(run(m))
}

func run(m *testing.M) int {
	logger := logrus.New()
	//	logger.SetOutput(ioutil.Discard)

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
		dr   database.Driver
		stmt string

		tables = []string{"distributions", "rules", "constraints", "variants", "segments", "flags"}
	)

	switch driver {
	case SQLite:
		dr, err = sqlite3.WithInstance(db, &sqlite3.Config{})

		stmt = "DELETE FROM %s"
		store = sqlite.NewStore(db)
	case Postgres:
		dr, err = pg.WithInstance(db, &pg.Config{})

		stmt = "TRUNCATE TABLE %s CASCADE"
		store = postgres.NewStore(db)
	case MySQL:
		dr, err = mysql.WithInstance(db, &mysql.Config{})

		stmt = "TRUNCATE TABLE %s"
	default:
		logger.Fatalf("unknown driver: %s", driver)
	}

	if err != nil {
		logger.Fatal(err)
	}

	for _, t := range tables {
		_, _ = db.Exec(fmt.Sprintf(stmt, t))
	}

	f := filepath.Clean(fmt.Sprintf("../../config/migrations/%s", driver))

	mm, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", f), driver.String(), dr)
	if err != nil {
		logger.Fatal(err)
	}

	if err := mm.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Fatal(err)
	}

	return m.Run()
}
