package sql

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	"github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.flipt.io/flipt/config/migrations"
	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap"
)

var expectedVersions = map[Driver]uint{
	SQLite:      10,
	Postgres:    8,
	MySQL:       6,
	CockroachDB: 5,
}

// Migrator is responsible for migrating the database schema
type Migrator struct {
	db       *sql.DB
	driver   Driver
	logger   *zap.Logger
	migrator *migrate.Migrate
}

// NewMigrator creates a new Migrator
func NewMigrator(cfg config.Config, logger *zap.Logger) (*Migrator, error) {
	sql, driver, err := open(cfg, Options{migrate: true})
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	var dr database.Driver

	switch driver {
	case SQLite:
		dr, err = sqlite3.WithInstance(sql, &sqlite3.Config{})
	case Postgres:
		dr, err = postgres.WithInstance(sql, &postgres.Config{})
	case CockroachDB:
		dr, err = cockroachdb.WithInstance(sql, &cockroachdb.Config{})
	case MySQL:
		dr, err = mysql.WithInstance(sql, &mysql.Config{})
	}

	logger.Debug("using driver", zap.String("driver", driver.String()))

	if err != nil {
		return nil, fmt.Errorf("getting db driver for: %s: %w", driver, err)
	}

	// source migrations from embedded config/migrations package
	// relative to the specific driver
	sourceDriver, err := iofs.New(migrations.FS, driver.String())
	if err != nil {
		return nil, err
	}

	mm, err := migrate.NewWithInstance("iofs", sourceDriver, driver.String(), dr)
	if err != nil {
		return nil, fmt.Errorf("creating migrate instance: %w", err)
	}

	return &Migrator{
		db:       sql,
		migrator: mm,
		logger:   logger,
		driver:   driver,
	}, nil
}

// Close closes the source and db
func (m *Migrator) Close() (source, db error) {
	return m.migrator.Close()
}

// Up runs any pending migrations
func (m *Migrator) Up(force bool) error {
	canAutoMigrate := force

	// check if any migrations are pending
	currentVersion, _, err := m.migrator.Version()

	if err != nil {
		if !errors.Is(err, migrate.ErrNilVersion) {
			return fmt.Errorf("getting current migrations version: %w", err)
		}

		m.logger.Debug("first run, running migrations...")

		// if first run then it's safe to migrate
		if err := m.migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("running migrations: %w", err)
		}

		m.logger.Debug("migrations complete")
		return nil
	}

	expectedVersion := expectedVersions[m.driver]

	if currentVersion < expectedVersion {
		if !canAutoMigrate {
			return errors.New("migrations pending, please backup your database and run `flipt migrate`")
		}

		m.logger.Debug("current migration", zap.Uint("current_version", currentVersion), zap.Uint("expected_version", expectedVersion))

		m.logger.Debug("running migrations...")

		if err := m.migrator.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("running migrations: %w", err)
		}

		m.logger.Debug("migrations complete")
		return nil
	}

	m.logger.Debug("migrations up to date")
	return nil
}

// Drop drops the database
func (m *Migrator) Drop() error {
	m.logger.Debug("running drop ...")

	switch m.driver {
	case SQLite:
		// disable foreign keys for sqlite to avoid errors when dropping tables
		// https://www.sqlite.org/foreignkeys.html#fk_enable
		// we dont need to worry about re-enabling them since we're dropping the db
		// and the connection will be closed
		_, _ = m.db.Exec("PRAGMA foreign_keys = OFF")
	case MySQL:
		// https://stackoverflow.com/questions/5452760/how-to-truncate-a-foreign-key-constrained-table
		_, _ = m.db.Exec("SET FOREIGN_KEY_CHECKS = 0;")
	}

	if err := m.migrator.Drop(); err != nil {
		return fmt.Errorf("dropping: %w", err)
	}

	m.logger.Debug("drop complete")
	return nil
}
