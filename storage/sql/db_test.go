package sql

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	ms "github.com/golang-migrate/migrate/database/mysql"
	pg "github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/database/sqlite3"
	"github.com/markphelps/flipt/config"
	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/storage/sql/mysql"
	"github.com/markphelps/flipt/storage/sql/postgres"
	"github.com/markphelps/flipt/storage/sql/sqlite"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "github.com/golang-migrate/migrate/source/file"
)

func TestOpen(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.DatabaseConfig
		driver  Driver
		wantErr bool
	}{
		{
			name: "sqlite url",
			cfg: config.DatabaseConfig{
				URL:             "file:flipt.db",
				MaxOpenConn:     5,
				ConnMaxLifetime: 30 * time.Minute,
			},
			driver: SQLite,
		},
		{
			name: "postres url",
			cfg: config.DatabaseConfig{
				URL: "postgres://postgres@localhost:5432/flipt?sslmode=disable",
			},
			driver: Postgres,
		},
		{
			name: "mysql url",
			cfg: config.DatabaseConfig{
				URL: "mysql://mysql@localhost:3306/flipt",
			},
			driver: MySQL,
		},
		{
			name: "invalid url",
			cfg: config.DatabaseConfig{
				URL: "http://a b",
			},
			wantErr: true,
		},
		{
			name: "unknown driver",
			cfg: config.DatabaseConfig{
				URL: "mongo://127.0.0.1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		var (
			cfg     = tt.cfg
			driver  = tt.driver
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			db, d, err := Open(config.Config{
				Database: cfg,
			})

			if wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, db)

			defer db.Close()

			assert.Equal(t, driver, d)
		})
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.DatabaseConfig
		dsn     string
		driver  Driver
		wantErr bool
	}{
		{
			name: "sqlite url",
			cfg: config.DatabaseConfig{
				URL: "file:flipt.db",
			},
			driver: SQLite,
			dsn:    "flipt.db?_fk=true&cache=shared",
		},
		{
			name: "sqlite",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabaseSQLite,
				Host:     "flipt.db",
			},
			driver: SQLite,
			dsn:    "flipt.db?_fk=true&cache=shared",
		},
		{
			name: "postres url",
			cfg: config.DatabaseConfig{
				URL: "postgres://postgres@localhost:5432/flipt?sslmode=disable",
			},
			driver: Postgres,
			dsn:    "dbname=flipt host=localhost port=5432 sslmode=disable user=postgres",
		},
		{
			name: "postgres no port",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabasePostgres,
				Name:     "flipt",
				Host:     "localhost",
				User:     "postgres",
			},
			driver: Postgres,
			dsn:    "dbname=flipt host=localhost user=postgres",
		},
		{
			name: "postgres no password",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabasePostgres,
				Name:     "flipt",
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
			},
			driver: Postgres,
			dsn:    "dbname=flipt host=localhost port=5432 user=postgres",
		},
		{
			name: "postgres with password",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabasePostgres,
				Name:     "flipt",
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Password: "foo",
			},
			driver: Postgres,
			dsn:    "dbname=flipt host=localhost password=foo port=5432 user=postgres",
		},
		{
			name: "mysql url",
			cfg: config.DatabaseConfig{
				URL: "mysql://mysql@localhost:3306/flipt",
			},
			driver: MySQL,
			dsn:    "mysql@tcp(localhost:3306)/flipt?multiStatements=true&parseTime=true&sql_mode=ANSI",
		},
		{
			name: "mysql no port",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabaseMySQL,
				Name:     "flipt",
				Host:     "localhost",
				User:     "mysql",
				Password: "foo",
			},
			driver: MySQL,
			dsn:    "mysql:foo@tcp(localhost:3306)/flipt?multiStatements=true&parseTime=true&sql_mode=ANSI",
		},
		{
			name: "mysql no password",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabaseMySQL,
				Name:     "flipt",
				Host:     "localhost",
				Port:     3306,
				User:     "mysql",
			},
			driver: MySQL,
			dsn:    "mysql@tcp(localhost:3306)/flipt?multiStatements=true&parseTime=true&sql_mode=ANSI",
		},
		{
			name: "mysql with password",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabaseMySQL,
				Name:     "flipt",
				Host:     "localhost",
				Port:     3306,
				User:     "mysql",
				Password: "foo",
			},
			driver: MySQL,
			dsn:    "mysql:foo@tcp(localhost:3306)/flipt?multiStatements=true&parseTime=true&sql_mode=ANSI",
		},
		{
			name: "invalid url",
			cfg: config.DatabaseConfig{
				URL: "http://a b",
			},
			wantErr: true,
		},
		{
			name: "unknown driver",
			cfg: config.DatabaseConfig{
				URL: "mongo://127.0.0.1",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		var (
			cfg     = tt.cfg
			driver  = tt.driver
			url     = tt.dsn
			wantErr = tt.wantErr
		)

		t.Run(tt.name, func(t *testing.T) {
			d, u, err := parse(config.Config{
				Database: cfg,
			}, false)

			if wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, driver, d)
			assert.Equal(t, url, u.DSN)
		})
	}
}

var store storage.Store

const defaultTestDBURL = "file:../../flipt_test.db"

func TestMain(m *testing.M) {
	// os.Exit skips defer calls
	// so we need to use another fn
	code, err := run(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func run(m *testing.M) (code int, err error) {

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		dbURL = defaultTestDBURL
	}

	cfg := config.Config{
		Database: config.DatabaseConfig{
			URL: dbURL,
		},
	}

	db, driver, err := open(cfg, true)
	if err != nil {
		return 1, err
	}

	var (
		dr   database.Driver
		stmt string

		tables = []string{"distributions", "rules", "constraints", "variants", "segments", "flags"}
	)

	switch driver {
	case SQLite:
		dr, err = sqlite3.WithInstance(db, &sqlite3.Config{})
		stmt = "DELETE FROM %s"
	case Postgres:
		dr, err = pg.WithInstance(db, &pg.Config{})
		stmt = "TRUNCATE TABLE %s CASCADE"
	case MySQL:
		dr, err = ms.WithInstance(db, &ms.Config{})
		stmt = "TRUNCATE TABLE %s"

		// https://stackoverflow.com/questions/5452760/how-to-truncate-a-foreign-key-constrained-table
		if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0;"); err != nil {
			return 1, fmt.Errorf("disabling foreign key checks: mysql: %w", err)
		}

	default:
		return 1, fmt.Errorf("unknown driver: %s", driver)
	}

	if err != nil {
		return 1, err
	}

	for _, t := range tables {
		_, _ = db.Exec(fmt.Sprintf(stmt, t))
	}

	f := filepath.Clean(fmt.Sprintf("../../config/migrations/%s", driver))

	mm, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", f), driver.String(), dr)
	if err != nil {
		return 1, err
	}

	if err := mm.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return 1, err
	}

	db, driver, err = open(cfg, false)
	if err != nil {
		return 1, err
	}

	defer db.Close()

	switch driver {
	case SQLite:
		store = sqlite.NewStore(db)
	case Postgres:
		store = postgres.NewStore(db)
	case MySQL:
		if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 1;"); err != nil {
			return 1, fmt.Errorf("enabling foreign key checks: mysql: %w", err)
		}

		store = mysql.NewStore(db)
	}

	return m.Run(), nil
}
