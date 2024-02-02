package testing

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/docker/go-connections/nat"
	"github.com/golang-migrate/migrate/v4"
	clickhouseMigrate "github.com/golang-migrate/migrate/v4/database/clickhouse"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.flipt.io/flipt/config/migrations"
	fliptsql "go.flipt.io/flipt/internal/storage/sql"
)

type Database struct {
	Container testcontainers.Container
	DB        *sql.DB
	Driver    fliptsql.Driver
}

func Open() (*Database, error) {
	container, hostIP, port, err := NewAnalyticsDBContainer(context.Background())
	if err != nil {
		return nil, err
	}

	db := clickhouse.OpenDB(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", hostIP, port)},
	})

	mm, err := newMigrator(db, fliptsql.Clickhouse)
	if err != nil {
		return nil, err
	}

	// ensure we start with a clean slate
	if err := mm.Drop(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	mm, err = newMigrator(db, fliptsql.Clickhouse)
	if err != nil {
		return nil, err
	}

	if err := mm.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, err
	}

	return &Database{
		Container: container,
		DB:        db,
		Driver:    fliptsql.Clickhouse,
	}, nil
}

func newMigrator(db *sql.DB, driver fliptsql.Driver) (*migrate.Migrate, error) {
	dr, err := clickhouseMigrate.WithInstance(db, &clickhouseMigrate.Config{
		MigrationsTableEngine: "MergeTree",
	})
	if err != nil {
		return nil, err
	}

	sourceDriver, err := iofs.New(migrations.FS, fliptsql.Clickhouse.Migrations())
	if err != nil {
		return nil, fmt.Errorf("constructing migration source driver (db driver %q): %w", driver.Migrations(), err)
	}

	mm, err := migrate.NewWithInstance("iofs", sourceDriver, fliptsql.Clickhouse.Migrations(), dr)
	if err != nil {
		return nil, err
	}

	return mm, nil
}

func NewAnalyticsDBContainer(ctx context.Context) (testcontainers.Container, string, int, error) {
	port := nat.Port("9000/tcp")

	req := testcontainers.ContainerRequest{
		Image:        "clickhouse/clickhouse-server:24.1-alpine",
		ExposedPorts: []string{"9000/tcp"},
		WaitingFor:   wait.ForListeningPort(port),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", 0, err
	}

	if err := container.StartLogProducer(ctx); err != nil {
		return nil, "", 0, err
	}

	mappedPort, err := container.MappedPort(ctx, port)
	if err != nil {
		return nil, "", 0, err
	}

	hostIP, err := container.Host(ctx)
	if err != nil {
		return nil, "", 0, err
	}

	return container, hostIP, mappedPort.Int(), nil
}
