package db

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/database/sqlite3"
	"github.com/markphelps/flipt/config"
	"github.com/markphelps/flipt/errors"
	"github.com/sirupsen/logrus"
)

var ErrMigrationsPending = errors.New("migrations pending")

// Migrator is responsible for migrating the database schema
type Migrator struct {
	cfg      *config.Config
	sql      *sql.DB
	migrator *migrate.Migrate
	logger   logrus.FieldLogger
}

// NewMigrator creates a new Migrator
func NewMigrator(cfg *config.Config, logger logrus.FieldLogger) (*Migrator, error) {
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
		cfg:      cfg,
		sql:      sql,
		migrator: mm,
		logger:   logger,
	}, nil
}

// Close closes the migrator
func (m *Migrator) Close() error {
	return m.sql.Close()
}

// Check checks for any pending migrations; if this
// is the first run, it migrates the database
func (m *Migrator) Check(wantVersion uint) error {
	v, _, err := m.migrator.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("getting current migrations version: %w", err)
	}

	// if first run, go ahead and run all migrations
	// otherwise exit and inform user to run manually if migrations are pending
	if err == migrate.ErrNilVersion {
		m.logger.Debug("no previous migrations run; running now")
		if err := m.Run(); err != nil {
			return fmt.Errorf("running migrations: %w", err)
		}
	} else if v < wantVersion {
		m.logger.Debugf("migrations pending: [current version=%d, want version=%d]", v, wantVersion)
		return ErrMigrationsPending
	}
	return nil
}

// Run runs any pending migrations
func (m *Migrator) Run() error {
	m.logger.Info("running migrations...")

	if err := m.migrator.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	m.logger.Info("finished migrations")
	return nil
}
