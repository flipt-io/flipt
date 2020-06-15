package db

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/mysql"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/database/sqlite3"
	"github.com/markphelps/flipt/config"
)

var expectedVersions = map[Driver]uint{
	SQLite:   2,
	Postgres: 2,
	MySQL:    0,
}

// Migrator is responsible for migrating the database schema
type Migrator struct {
	driver   Driver
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
		driver:   driver,
	}, nil
}

// Close closes the source and db
func (m *Migrator) Close() (source, db error) {
	return m.migrator.Close()
}

// Run runs any pending migrations
func (m *Migrator) Run(force bool) error {
	canAutoMigrate := force

	// check if any migrations are pending
	currentVersion, _, err := m.migrator.Version()

	if err != nil {
		if !errors.Is(err, migrate.ErrNilVersion) {
			return fmt.Errorf("getting current migrations version: %w", err)
		}

		// if first run then it's safe to migrate
		canAutoMigrate = true
	}

	expectedVersion := expectedVersions[m.driver]

	if currentVersion < expectedVersion {
		if !canAutoMigrate {
			return errors.New("migrations pending, please backup your database and run `flipt migrate`")
		}

		if err := m.migrator.Up(); err != nil && err != migrate.ErrNoChange {
			return fmt.Errorf("running migrations: %w", err)
		}
	}

	return nil
}
