package sql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/sql/mysql"
	"go.flipt.io/flipt/internal/storage/sql/postgres"
	"go.flipt.io/flipt/internal/storage/sql/sqlite"
	"go.uber.org/zap/zaptest"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	ms "github.com/golang-migrate/migrate/database/mysql"
	pg "github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/database/sqlite3"
	_ "github.com/golang-migrate/migrate/source/file"
)

func TestOpen(t *testing.T) {
	logger := zaptest.NewLogger(t)

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
			name: "postgres url",
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
			name: "cockroachdb url",
			cfg: config.DatabaseConfig{
				URL: "cockroachdb://cockroachdb@localhost:26257/flipt?sslmode=disable",
			},
			driver: CockroachDB,
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
			}, logger)

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
	logger := zaptest.NewLogger(t)

	tests := []struct {
		name    string
		cfg     config.DatabaseConfig
		dsn     string
		driver  Driver
		options options
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
			name: "postgres url",
			cfg: config.DatabaseConfig{
				URL: "postgres://postgres@localhost:5432/flipt?sslmode=disable",
			},
			driver: Postgres,
			dsn:    "dbname=flipt host=localhost port=5432 sslmode=disable user=postgres",
		},
		{
			name: "postgres no disable sslmode",
			cfg: config.DatabaseConfig{
				URL: "postgres://postgres@localhost:5432/flipt",
			},
			driver: Postgres,
			dsn:    "dbname=flipt host=localhost port=5432 user=postgres",
		},
		{
			name: "postgres disable sslmode via opts",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabasePostgres,
				Name:     "flipt",
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
			},
			options: options{
				sslDisabled: true,
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
			name: "mysql no ANSI sql mode via opts",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabaseMySQL,
				Name:     "flipt",
				Host:     "localhost",
				Port:     3306,
				User:     "mysql",
			},
			options: options{
				migrate: true,
			},
			driver: MySQL,
			dsn:    "mysql@tcp(localhost:3306)/flipt?multiStatements=true&parseTime=true",
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
			name: "cockroachdb url",
			cfg: config.DatabaseConfig{
				URL: "cockroachdb://cockroachdb@localhost:26257/flipt?sslmode=disable",
			},
			driver: CockroachDB,
			dsn:    "postgres://cockroachdb@localhost:26257/flipt?sslmode=disable",
		},
		{
			name: "cockroachdb url alternative (cockroach://",
			cfg: config.DatabaseConfig{
				URL: "cockroach://cockroachdb@localhost:26257/flipt?sslmode=disable",
			},
			driver: CockroachDB,
			dsn:    "postgres://cockroachdb@localhost:26257/flipt?sslmode=disable",
		},
		{
			name: "cockroachdb url alternative (crdb://",
			cfg: config.DatabaseConfig{
				URL: "crdb://cockroachdb@localhost:26257/flipt?sslmode=disable",
			},
			driver: CockroachDB,
			dsn:    "postgres://cockroachdb@localhost:26257/flipt?sslmode=disable",
		},
		{
			name: "cockroachdb default disable sslmode",
			// cockroachdb defaults to sslmode=disable
			// https://www.cockroachlabs.com/docs/stable/connection-parameters.html#additional-connection-parameters
			cfg: config.DatabaseConfig{
				URL: "cockroachdb://cockroachdb@localhost:26257/flipt",
			},
			driver: CockroachDB,
			dsn:    "postgres://cockroachdb@localhost:26257/flipt?sslmode=disable",
		},
		{
			name: "cockroachdb disable sslmode via opts",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabaseCockroachDB,
				Name:     "flipt",
				Host:     "localhost",
				Port:     26257,
				User:     "cockroachdb",
			},
			options: options{
				sslDisabled: true,
			},
			driver: CockroachDB,
			dsn:    "postgres://cockroachdb@localhost:26257/flipt?sslmode=disable",
		},
		{
			name: "cockroachdb no port",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabaseCockroachDB,
				Name:     "flipt",
				Host:     "localhost",
				User:     "cockroachdb",
			},
			driver: CockroachDB,
			dsn:    "postgres://cockroachdb@localhost:26257/flipt?sslmode=disable",
		},
		{
			name: "cockroachdb no password",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabaseCockroachDB,
				Name:     "flipt",
				Host:     "localhost",
				Port:     26257,
				User:     "cockroachdb",
			},
			driver: CockroachDB,
			dsn:    "postgres://cockroachdb@localhost:26257/flipt?sslmode=disable",
		},
		{
			name: "cockroachdb with password",
			cfg: config.DatabaseConfig{
				Protocol: config.DatabaseCockroachDB,
				Name:     "flipt",
				Host:     "localhost",
				Port:     26257,
				User:     "cockroachdb",
				Password: "foo",
			},
			driver: CockroachDB,
			dsn:    "postgres://cockroachdb:foo@localhost:26257/flipt?sslmode=disable",
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
		tt := tt

		var (
			cfg     = tt.cfg
			driver  = tt.driver
			url     = tt.dsn
			wantErr = tt.wantErr
			opts    = tt.options
		)

		t.Run(tt.name, func(t *testing.T) {
			d, u, err := parse(config.Config{
				Database: cfg,
			}, logger, opts)

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

func TestDBTestSuite(t *testing.T) {
	suite.Run(t, new(DBTestSuite))
}

const defaultTestDBURL = "file:../../flipt_test.db"

type dbContainer struct {
	testcontainers.Container
	host string
	port int
}

type DBTestSuite struct {
	suite.Suite
	db            *sql.DB
	store         storage.Store
	driver        Driver
	testcontainer *dbContainer
}

var dd string

func TestMain(m *testing.M) {
	dd = os.Getenv("FLIPT_TEST_DATABASE_PROTOCOL")
	os.Exit(m.Run())
}

func (s *DBTestSuite) SetupSuite() {
	setup := func() error {
		var (
			logger = zaptest.NewLogger(s.T())
			proto  config.DatabaseProtocol
		)

		switch dd {
		case "postgres":
			proto = config.DatabasePostgres
		case "mysql":
			proto = config.DatabaseMySQL
		default:
			proto = config.DatabaseSQLite
		}

		cfg := config.Config{
			Database: config.DatabaseConfig{
				Protocol: proto,
				URL:      defaultTestDBURL,
			},
		}

		if proto != config.DatabaseSQLite {
			dbContainer, err := newDBContainer(s.T(), context.Background(), proto)
			if err != nil {
				return fmt.Errorf("creating db container: %w", err)
			}

			cfg.Database.URL = ""
			cfg.Database.Host = dbContainer.host
			cfg.Database.Port = dbContainer.port
			cfg.Database.Name = "flipt_test"
			cfg.Database.User = "flipt"
			cfg.Database.Password = "password"

			s.testcontainer = dbContainer
		}

		db, driver, err := open(cfg, logger, options{migrate: true, sslDisabled: true})
		if err != nil {
			return fmt.Errorf("opening db: %w", err)
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
				return fmt.Errorf("disabling foreign key checks: %w", err)
			}

		default:
			return fmt.Errorf("unknown driver: %s", proto)
		}

		if err != nil {
			return fmt.Errorf("creating driver: %w", err)
		}

		for _, t := range tables {
			_, _ = db.Exec(fmt.Sprintf(stmt, t))
		}

		f := filepath.Clean(fmt.Sprintf("../../../config/migrations/%s", driver))

		mm, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", f), driver.String(), dr)
		if err != nil {
			return fmt.Errorf("creating migrate instance: %w", err)
		}

		if err := mm.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("running migrations: %w", err)
		}

		if err := db.Close(); err != nil {
			return fmt.Errorf("closing db: %w", err)
		}

		// re-open db and enable ANSI mode for MySQL
		db, driver, err = open(cfg, logger, options{migrate: false, sslDisabled: true})
		if err != nil {
			return fmt.Errorf("opening db: %w", err)
		}

		s.db = db
		s.driver = driver

		var store storage.Store

		switch driver {
		case SQLite:
			store = sqlite.NewStore(db, logger)
		case Postgres:
			store = postgres.NewStore(db, logger)
		case MySQL:
			if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 1;"); err != nil {
				return fmt.Errorf("enabling foreign key checks: %w", err)
			}

			store = mysql.NewStore(db, logger)
		}

		s.store = store
		return nil
	}

	s.Require().NoError(setup())
}

func (s *DBTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if s.testcontainer != nil {
		_ = s.testcontainer.Terminate(shutdownCtx)
	}
}

func newDBContainer(t *testing.T, ctx context.Context, proto config.DatabaseProtocol) (*dbContainer, error) {
	t.Helper()

	if testing.Short() {
		t.Skipf("skipping running %s tests in short mode", proto.String())
	}

	var (
		req  testcontainers.ContainerRequest
		port nat.Port
	)

	switch proto {
	case config.DatabasePostgres:
		port = nat.Port("5432/tcp")
		req = testcontainers.ContainerRequest{
			Image:        "postgres:11.2",
			ExposedPorts: []string{"5432/tcp"},
			WaitingFor:   wait.ForListeningPort(port),
			Env: map[string]string{
				"POSTGRES_USER":     "flipt",
				"POSTGRES_PASSWORD": "password",
				"POSTGRES_DB":       "flipt_test",
			},
		}
	case config.DatabaseMySQL:
		port = nat.Port("3306/tcp")
		req = testcontainers.ContainerRequest{
			Image:        "mysql:8",
			ExposedPorts: []string{"3306/tcp"},
			WaitingFor:   wait.ForListeningPort(port),
			Env: map[string]string{
				"MYSQL_USER":                 "flipt",
				"MYSQL_PASSWORD":             "password",
				"MYSQL_DATABASE":             "flipt_test",
				"MYSQL_ALLOW_EMPTY_PASSWORD": "true",
			},
		}
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, err
	}

	mappedPort, err := container.MappedPort(ctx, port)
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	return &dbContainer{Container: container, host: hostIP, port: mappedPort.Int()}, nil
}
