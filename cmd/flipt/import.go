package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/storage"
	"go.flipt.io/flipt/storage/sql"
	"go.flipt.io/flipt/storage/sql/mysql"
	"go.flipt.io/flipt/storage/sql/postgres"
	"go.flipt.io/flipt/storage/sql/sqlite"
	"go.uber.org/zap"
)

var (
	dropBeforeImport bool
	importStdin      bool
)

func runImport(ctx context.Context, logger *zap.Logger, args []string) error {
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interrupt
		cancel()
	}()

	db, driver, err := sql.Open(*cfg)
	if err != nil {
		return fmt.Errorf("opening db: %w", err)
	}

	defer db.Close()

	var store storage.Store

	switch driver {
	case sql.SQLite:
		store = sqlite.NewStore(db, logger)
	case sql.Postgres:
		store = postgres.NewStore(db, logger)
	case sql.MySQL:
		store = mysql.NewStore(db, logger)
	}

	var in io.ReadCloser = os.Stdin

	if !importStdin {
		importFilename := args[0]
		if importFilename == "" {
			return errors.New("import filename required")
		}

		f := filepath.Clean(importFilename)

		logger.Debug("importing", zap.String("source_path", f))

		in, err = os.Open(f)
		if err != nil {
			return fmt.Errorf("opening import file: %w", err)
		}
	}

	defer in.Close()

	// drop tables if specified
	if dropBeforeImport {
		logger.Debug("dropping tables before import")

		tables := []string{"schema_migrations", "distributions", "rules", "constraints", "variants", "segments", "flags"}

		for _, table := range tables {
			if _, err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)); err != nil {
				return fmt.Errorf("dropping tables: %w", err)
			}
		}
	}

	migrator, err := sql.NewMigrator(*cfg, logger)
	if err != nil {
		return err
	}

	defer migrator.Close()

	if err := migrator.Run(forceMigrate); err != nil {
		return err
	}

	if _, err := migrator.Close(); err != nil {
		return fmt.Errorf("closing migrator: %w", err)
	}

	importer := ext.NewImporter(store)
	if err := importer.Import(ctx, in); err != nil {
		return fmt.Errorf("importing: %w", err)
	}

	return nil
}
