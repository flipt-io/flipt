package testing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"

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

const defaultTestDBURL = "file:../../flipt_test.db"

type Database struct {
	DB        *sql.DB
	Driver    fliptsql.Driver
	Container *DBContainer
}

func (d *Database) Shutdown(ctx context.Context) {
	if d.DB != nil {
		d.DB.Close()
	}

	if d.Container != nil {
		_ = d.Container.Terminate(ctx)
	}
}

func Open() (*Database, error) {
	var proto config.DatabaseProtocol

	switch os.Getenv("FLIPT_TEST_DATABASE_PROTOCOL") {
	case "cockroachdb":
		proto = config.DatabaseCockroachDB
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

	var (
		username, password, dbName string
		useTestContainer           bool
	)

	switch proto {
	case config.DatabaseSQLite:
		// no-op
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
	}

	db, driver, err := fliptsql.Open(cfg, fliptsql.WithMigrate, fliptsql.WithSSLDisabled)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	var (
		dr   database.Driver
		stmt string

		tables = []string{"distributions", "rules", "constraints", "variants", "segments", "flags", "authentications"}
	)

	switch driver {
	case fliptsql.SQLite:
		dr, err = sqlite3.WithInstance(db, &sqlite3.Config{})
		stmt = "DELETE FROM %s"
	case fliptsql.Postgres:
		dr, err = pg.WithInstance(db, &pg.Config{})
		stmt = "TRUNCATE TABLE %s CASCADE"
	case fliptsql.CockroachDB:
		dr, err = cockroachdb.WithInstance(db, &cockroachdb.Config{})
		stmt = "TRUNCATE TABLE %s CASCADE"
	case fliptsql.MySQL:
		dr, err = ms.WithInstance(db, &ms.Config{})
		stmt = "TRUNCATE TABLE %s"

		// https://stackoverflow.com/questions/5452760/how-to-truncate-a-foreign-key-constrained-table
		if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 0;"); err != nil {
			return nil, fmt.Errorf("disabling foreign key checks: %w", err)
		}

	default:
		return nil, fmt.Errorf("unknown driver: %s", proto)
	}

	if err != nil {
		return nil, fmt.Errorf("creating driver: %w", err)
	}

	for _, t := range tables {
		_, _ = db.Exec(fmt.Sprintf(stmt, t))
	}

	// source migrations from embedded config/migrations package
	// relative to the specific driver
	sourceDriver, err := iofs.New(migrations.FS, driver.String())
	if err != nil {
		return nil, fmt.Errorf("constructing migration source driver (db driver %q): %w", driver.String(), err)
	}

	mm, err := migrate.NewWithInstance("iofs", sourceDriver, driver.String(), dr)
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

	if driver == fliptsql.MySQL {
		if _, err := db.Exec("SET FOREIGN_KEY_CHECKS = 1;"); err != nil {
			return nil, fmt.Errorf("enabling foreign key checks: %w", err)
		}
	}

	return &Database{
		DB:        db,
		Driver:    driver,
		Container: container,
	}, nil
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
			WaitingFor:   wait.ForListeningPort(port),
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
			WaitingFor:   wait.ForHTTP("/health").WithPort("8080"),
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
