package main

import (
	"fmt"
	"path/filepath"

	"github.com/golang-migrate/migrate"
	"github.com/golang-migrate/migrate/database"
	"github.com/golang-migrate/migrate/database/postgres"
	"github.com/golang-migrate/migrate/database/sqlite3"
	"github.com/markphelps/flipt/storage/db"
)

func runMigrations() error {
	sql, driver, err := db.Open(cfg.Database.URL)
	if err != nil {
		return fmt.Errorf("opening db: %w", err)
	}

	defer sql.Close()

	var dr database.Driver

	switch driver {
	case db.SQLite:
		dr, err = sqlite3.WithInstance(sql, &sqlite3.Config{})
	case db.Postgres:
		dr, err = postgres.WithInstance(sql, &postgres.Config{})
	}

	if err != nil {
		return fmt.Errorf("getting db driver for: %s: %w", driver, err)
	}

	f := filepath.Clean(fmt.Sprintf("%s/%s", cfg.Database.MigrationsPath, driver))

	mm, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", f), driver.String(), dr)
	if err != nil {
		return fmt.Errorf("opening migrations: %w", err)
	}

	logger.Info("running migrations...")

	if err := mm.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	logger.Info("finished migrations")

	return nil
}
