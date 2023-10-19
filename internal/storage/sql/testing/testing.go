package testing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.flipt.io/flipt/config/migrations"
	"go.flipt.io/flipt/internal/config"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"

	ms "github.com/golang-migrate/migrate/v4/database/mysql"
	pg "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
)

const defaultTestDBPrefix = "flipt_*.db"

type Database struct {
	DB        *sql.DB
	Driver    fliptsql.Driver
	Container *DBContainer

	cleanup func()
}

func (d *Database) Shutdown(ctx context.Context) {
	if d.DB != nil {
		d.DB.Close()
	}

	if d.Container != nil {
		_ = d.Container.StopLogProducer()
		_ = d.Container.Terminate(ctx)
	}

	if d.cleanup != nil {
		d.cleanup()
	}
}

func Open() (*Database, error) {
	var proto config.DatabaseProtocol

	switch os.Getenv("FLIPT_TEST_DATABASE_PROTOCOL") {
	case "cockroachdb", "cockroach":
		proto = config.DatabaseCockroachDB
	case "postgres":
		proto = config.DatabasePostgres
	case "mysql":
		proto = config.DatabaseMySQL
	case "libsql":
		proto = config.DatabaseLibSQL
	default:
		proto = config.DatabaseSQLite
	}

	cfg := config.Config{
		Database: config.DatabaseConfig{
			Protocol: proto,
		},
	}

	var (
		username, password, dbName string
		useTestContainer           bool
		cleanup                    func()
	)

	if url := os.Getenv("FLIPT_TEST_DB_URL"); len(url) > 0 {
		// FLIPT_TEST_DB_URL takes precedent if set.
		// It assumes the database is already running at the target URL.
		// It does not attempt to create an instance of the DB or do any cleanup.
		cfg.Database.URL = url
	} else {
		// Otherwise, depending on the value of FLIPT_TEST_DATABASE_PROTOCOL a test database
		// is created and destroyed for the lifecycle of the test.
		switch proto {
		case config.DatabaseSQLite:
			dbPath := createTempDBPath()
			cfg.Database.URL = "file:" + dbPath
			cleanup = func() {
				_ = os.Remove(dbPath)
			}
		case config.DatabaseLibSQL:
			dbPath := createTempDBPath()
			cfg.Database.URL = "libsql://file:" + dbPath
			cleanup = func() {
				_ = os.Remove(dbPath)
			}
		case config.DatabaseCockroachDB:
			useTestContainer = true
			username = "root"
			password = ""
			dbName = "defaultdb"
		default:
			useTestContainer = true
			username = "flipt"
			password = "password"
			dbName = "flipt_test"
		}
	}

	var (
		container *DBContainer
		err       error
	)

	if useTestContainer {
		container, err = NewDBContainer(context.Background(), proto)
		if err != nil {
			return nil, fmt.Errorf("creating db container: %w", err)
		}

		cfg.Database.URL = ""
		cfg.Database.Host = container.Host
		cfg.Database.Port = container.Port
		cfg.Database.Name = dbName
		cfg.Database.User = username
		cfg.Database.Password = password
		cfg.Database.ConnMaxLifetime = 1 * time.Minute
	}

	db, driver, err := fliptsql.Open(cfg, fliptsql.WithMigrate, fliptsql.WithSSLDisabled)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	mm, err := newMigrator(db, driver)
	if err != nil {
		return nil, fmt.Errorf("creating migrate instance: %w", err)
	}

	// run drop to clear target DB (incase we're reusing)
	if err := mm.Drop(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("running drop: %w", err)
	}

	if err := db.Close(); err != nil {
		return nil, fmt.Errorf("closing db: %w", err)
	}

	// need new instance after drop
	db, driver, err = fliptsql.Open(cfg, fliptsql.WithMigrate, fliptsql.WithSSLDisabled)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	mm, err = newMigrator(db, driver)
	if err != nil {
		return nil, fmt.Errorf("creating migrate instance: %w", err)
	}

	if err := mm.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("running migrations: %w", err)
	}

	if err := db.Close(); err != nil {
		return nil, fmt.Errorf("closing db: %w", err)
	}

	// re-open db and enable ANSI mode for MySQL
	db, driver, err = fliptsql.Open(cfg, fliptsql.WithSSLDisabled)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	db.SetConnMaxLifetime(2 * time.Minute)
	db.SetConnMaxIdleTime(time.Minute)

	// 2 minute timeout attempting to establish first connection
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &Database{
		DB:        db,
		Driver:    driver,
		Container: container,
		cleanup:   cleanup,
	}, nil
}

func newMigrator(db *sql.DB, driver fliptsql.Driver) (*migrate.Migrate, error) {
	var (
		dr  database.Driver
		err error
	)

	switch driver {
	case fliptsql.SQLite, fliptsql.LibSQL:
		dr, err = sqlite3.WithInstance(db, &sqlite3.Config{})
	case fliptsql.Postgres:
		dr, err = pg.WithInstance(db, &pg.Config{})
	case fliptsql.CockroachDB:
		dr, err = cockroachdb.WithInstance(db, &cockroachdb.Config{})
	case fliptsql.MySQL:
		dr, err = ms.WithInstance(db, &ms.Config{})

	default:
		return nil, fmt.Errorf("unknown driver: %s", driver)
	}

	if err != nil {
		return nil, fmt.Errorf("creating driver: %w", err)
	}

	// source migrations from embedded config/migrations package
	// relative to the specific driver
	sourceDriver, err := iofs.New(migrations.FS, driver.Migrations())
	if err != nil {
		return nil, fmt.Errorf("constructing migration source driver (db driver %q): %w", driver.String(), err)
	}

	return migrate.NewWithInstance("iofs", sourceDriver, driver.Migrations(), dr)
}

type DBContainer struct {
	testcontainers.Container
	Host string
	Port int
}

func NewDBContainer(ctx context.Context, proto config.DatabaseProtocol) (*DBContainer, error) {
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
			WaitingFor: wait.ForSQL(port, "postgres", func(host string, port nat.Port) string {
				return fmt.Sprintf("postgres://flipt:password@%s:%s/flipt_test?sslmode=disable", host, port.Port())
			}),
			Env: map[string]string{
				"POSTGRES_USER":     "flipt",
				"POSTGRES_PASSWORD": "password",
				"POSTGRES_DB":       "flipt_test",
			},
		}
	case config.DatabaseCockroachDB:
		port = nat.Port("26257/tcp")
		req = testcontainers.ContainerRequest{
			Image:        "cockroachdb/cockroach:latest-v21.2",
			ExposedPorts: []string{"26257/tcp", "8080/tcp"},
			WaitingFor: wait.ForSQL(port, "postgres", func(host string, port nat.Port) string {
				return fmt.Sprintf("postgres://root@%s:%s/defaultdb?sslmode=disable", host, port.Port())
			}),
			Env: map[string]string{
				"COCKROACH_USER":     "root",
				"COCKROACH_DATABASE": "defaultdb",
			},
			Cmd: []string{"start-single-node", "--insecure"},
		}
	case config.DatabaseMySQL:
		port = nat.Port("3306/tcp")
		req = testcontainers.ContainerRequest{
			Image:        "mysql:8",
			ExposedPorts: []string{"3306/tcp"},
			WaitingFor: wait.ForSQL(port, "mysql", func(host string, port nat.Port) string {
				return fmt.Sprintf("flipt:password@tcp(%s:%s)/flipt_test?multiStatements=true", host, port.Port())
			}),
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

	if err := container.StartLogProducer(ctx); err != nil {
		return nil, err
	}

	verbose, _ := strconv.ParseBool(strings.TrimSpace(os.Getenv("FLIPT_TEST_DATABASE_VERBOSE")))
	if verbose {
		var logger testContainerLogger
		container.FollowOutput(&logger)
	}

	mappedPort, err := container.MappedPort(ctx, port)
	if err != nil {
		return nil, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, err
	}

	return &DBContainer{Container: container, Host: hostIP, Port: mappedPort.Int()}, nil
}

type testContainerLogger struct{}

func (t testContainerLogger) Accept(entry testcontainers.Log) {
	log.Println(entry.LogType, ":", string(entry.Content))
}

func createTempDBPath() string {
	fi, err := os.CreateTemp("", defaultTestDBPrefix)
	if err != nil {
		panic(err)
	}
	_ = fi.Close()
	return fi.Name()
}
