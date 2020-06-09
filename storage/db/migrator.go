package db

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/mysql"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/database/sqlite3"
	"github.com/markphelps/flipt/config"
	"github.com/markphelps/flipt/errors"
)

var ErrMigrationsNilVersion = errors.New("migrations nil version")

// Migrator is responsible for migrating the database schema
type Migrator struct {
	migrator *migrate.Migrate
}

// NewMigrator creates a new Migrator
func NewMigrator(cfg *config.Config) (*Migrator, error) {
	sql, driver, err := Open(cfg.Database.URL)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	var dr database.Driver

	switch driver {
	case SQLite:
		dr, err = sqlite3.WithInstance(sql, &sqlite3.Config{})
	case Postgres:
		dr, err = postgres.WithInstance(sql, &postgres.Config{})
	case MySQL:
		dr, err = mysql.WithInstance(sql, &mysql.Config{})
	}

	if err != nil {
		return nil, fmt.Errorf("getting db driver for: %s: %w", driver, err)
	}

	f := filepath.Clean(fmt.Sprintf("%s/%s", cfg.Database.MigrationsPath, driver))

	mm, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", f), driver.String(), dr)
	if err != nil {
		return nil, fmt.Errorf("opening migrations: %w", err)
	}

	return &Migrator{
		migrator: mm,
	}, nil
}

// Close closes the source and db
func (m *Migrator) Close() (source, db error) {
	return m.migrator.Close()
}

// CurrentVersion returns the current migration version
func (m *Migrator) CurrentVersion() (uint, error) {
	v, _, err := m.migrator.Version()

	// migrations never run
	if err == migrate.ErrNilVersion {
		return 0, ErrMigrationsNilVersion
	}

	if err != nil {
		return 0, fmt.Errorf("getting current migrations version: %w", err)
	}

	return v, nil
}

// Run runs any pending migrations
func (m *Migrator) Run() error {
	if err := m.migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("running migrations: %w", err)
	}

	return nil
}
