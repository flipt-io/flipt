package main

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/internal/server"
	"go.flipt.io/flipt/internal/storage"
	"go.flipt.io/flipt/internal/storage/sql"
	"go.flipt.io/flipt/internal/storage/sql/mysql"
	"go.flipt.io/flipt/internal/storage/sql/postgres"
	"go.flipt.io/flipt/internal/storage/sql/sqlite"
	sdk "go.flipt.io/flipt/sdk/go"
	sdkgrpc "go.flipt.io/flipt/sdk/go/grpc"
	sdkhttp "go.flipt.io/flipt/sdk/go/http"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var (
	exportFilename string
	exportAddress  string
	exportToken    string
)

func runExport(ctx context.Context, logger *zap.Logger, cfg *config.Config) error {
	ctx, cancel := context.WithCancel(ctx)

	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interrupt
		cancel()
	}()

	var lister ext.Lister

	// Switch on the presence of the export address
	// Use direct DB access if not supplied
	// Otherwise, use the Go SDK to access Flipt remotely.
	if exportAddress == "" {
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

		lister = server.New(logger, store)
	} else {
		addr, err := url.Parse(exportAddress)
		if err != nil {
			logger.Fatal("Export address is invalid", zap.Error(err))
		}

		var transport sdk.Transport
		switch addr.Scheme {
		case "http":
			transport = sdkhttp.NewTransport(exportAddress)
		case "grpc":
			conn, err := grpc.Dial(addr.Host,
				grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				logger.Fatal("Failed to dial Flipt", zap.Error(err))
			}

			transport = sdkgrpc.NewTransport(conn)
		}

		var opts []sdk.Option
		if exportToken != "" {
			opts = append(opts, sdk.WithClientTokenProvider(
				sdk.StaticClientTokenProvider(exportToken),
			))
		}

		client := sdk.New(transport, opts...)
		lister = client.Flipt()
	}

	// default to stdout
	var out io.WriteCloser = os.Stdout

	// export to file
	if exportFilename != "" {
		logger.Debug("exporting", zap.String("destination_path", exportFilename))

		var err error
		out, err = os.Create(exportFilename)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}

		fmt.Fprintf(out, "# exported by Flipt (%s) on %s\n\n", version, time.Now().UTC().Format(time.RFC3339))

		defer out.Close()
	}

	exporter := ext.NewExporter(lister, storage.DefaultNamespace)
	if err := exporter.Export(ctx, out); err != nil {
		return fmt.Errorf("exporting: %w", err)
	}

	return nil
}
