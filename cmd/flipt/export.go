package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/sql"
	"go.flipt.io/flipt/internal/storage/sql/mysql"
	"go.flipt.io/flipt/internal/storage/sql/postgres"
	"go.flipt.io/flipt/internal/storage/sql/sqlite"
	"go.uber.org/zap"
)

var exportFilename string

func runExport(ctx context.Context, logger *zap.Logger) error {
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
	case sql.Postgres, sql.CockroachDB:
		store = postgres.NewStore(db, logger)
	case sql.MySQL:
		store = mysql.NewStore(db, logger)
	}

	// default to stdout
	var out io.WriteCloser = os.Stdout

	// export to file
	if exportFilename != "" {
		logger.Debug("exporting", zap.String("destination_path", exportFilename))

		out, err = os.Create(exportFilename)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}

		fmt.Fprintf(out, "# exported by Flipt (%s) on %s\n\n", version, time.Now().UTC().Format(time.RFC3339))
	}

	defer out.Close()

	exporter := ext.NewExporter(store)
	if err := exporter.Export(ctx, out); err != nil {
		return fmt.Errorf("exporting: %w", err)
	}

	return nil
}
