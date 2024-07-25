package sql

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database"
	clickhouseMigrate "github.com/golang-migrate/migrate/v4/database/clickhouse"
	"github.com/golang-migrate/migrate/v4/database/cockroachdb"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/database/sqlite3"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"go.flipt.io/flipt/config/migrations"
	"go.flipt.io/flipt/internal/config"
	"go.uber.org/zap"
)

var expectedVersions = map[Driver]uint{
	SQLite:      14,
	LibSQL:      14, // libsql driver uses the same migrations as sqlite3
	Postgres:    15,
	MySQL:       14,
	CockroachDB: 12,
	Clickhouse:  3,
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
	case SQLite, LibSQL:
		dr, err = sqlite3.WithInstance(sql, &sqlite3.Config{})
	case Postgres:
		dr, err = pgx.WithInstance(sql, &pgx.Config{})
	case CockroachDB:
		dr, err = cockroachdb.WithInstance(sql, &cockroachdb.Config{})
	case MySQL:
		dr, err = mysql.WithInstance(sql, &mysql.Config{})
	}

	if err != nil {
		return nil, fmt.Errorf("getting db driver for: %s: %w", driver, err)
	}

	logger.Debug("using driver", zap.String("driver", driver.String()))

	return migratorHelper(logger, sql, driver, dr)
}

func migratorHelper(logger *zap.Logger, db *sql.DB, driver Driver, databaseDriver database.Driver) (*Migrator, error) {
	// source migrations from embedded config/migrations package
	// relative to the specific driver
	sourceDriver, err := iofs.New(migrations.FS, driver.Migrations())
	if err != nil {
		return nil, err
	}

	mm, err := migrate.NewWithInstance("iofs", sourceDriver, driver.Migrations(), databaseDriver)
	if err != nil {
		return nil, fmt.Errorf("creating migrate instance: %w", err)
	}

	return &Migrator{
		db:       db,
		migrator: mm,
		logger:   logger,
		driver:   driver,
	}, nil
}

// NewAnalyticsMigrator returns a migrator for analytics databases
func NewAnalyticsMigrator(cfg config.Config, logger *zap.Logger) (*Migrator, error) {
	sql, driver, err := openAnalytics(cfg)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
	}

	var dr database.Driver

	if driver == Clickhouse {
		options, err := cfg.Analytics.Storage.Clickhouse.Options()
		if err != nil {
			return nil, err
		}

		dr, err = clickhouseMigrate.WithInstance(sql, &clickhouseMigrate.Config{
			DatabaseName:          options.Auth.Database,
			MigrationsTableEngine: "MergeTree",
		})
		if err != nil {
			return nil, fmt.Errorf("getting db driver for: %s: %w", driver, err)
		}
	}

	logger.Debug("using driver", zap.String("driver", driver.String()))

	return migratorHelper(logger, sql, driver, dr)
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
