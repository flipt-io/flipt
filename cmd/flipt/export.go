package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/markphelps/flipt/internal/ext"
	"github.com/markphelps/flipt/storage"
	"github.com/markphelps/flipt/storage/sql"
	"github.com/markphelps/flipt/storage/sql/mysql"
	"github.com/markphelps/flipt/storage/sql/postgres"
	"github.com/markphelps/flipt/storage/sql/sqlite"
)

var exportFilename string

func runExport(_ []string) error {
	ctx := context.Background()
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
		store = sqlite.NewStore(db)
	case sql.Postgres:
		store = postgres.NewStore(db)
	case sql.MySQL:
		store = mysql.NewStore(db)
	}

	// default to stdout
	var out io.WriteCloser = os.Stdout

	// export to file
	if exportFilename != "" {
		l.Debugf("exporting to %q", exportFilename)

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
