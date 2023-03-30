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

	"github.com/spf13/cobra"
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

type exportCommand struct {
	filename string
	address  string
	token    string
}

func newExportCommand() *cobra.Command {
	export := &exportCommand{}

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export flags/segments/rules to file/stdout",
		RunE:  export.run,
	}

	cmd.Flags().StringVarP(
		&export.filename,
		"output", "o",
		"",
		"export to filename (default STDOUT)",
	)

	cmd.Flags().StringVarP(
		&export.address,
		"export-from-address", "",
		"",
		"address of remote Flipt instance to export from (defaults to direct DB export if not supplied)",
	)

	cmd.Flags().StringVarP(
		&export.token,
		"export-from-token", "",
		"",
		"client token used to authenticate access to remote Flipt instance when exporting.",
	)

	return cmd
}

func (c *exportCommand) run(cmd *cobra.Command, _ []string) error {
	// Switch on the presence of the export address
	// Use direct DB access if not supplied
	// Otherwise, use the Go SDK to access Flipt remotely.
	if c.address == "" {
		logger, cfg := buildConfig()

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

		return c.export(cmd.Context(), logger, server.New(logger, store))
	}

	logger := zap.Must(zap.NewDevelopment())

	addr, err := url.Parse(c.address)
	if err != nil {
		logger.Fatal("Export address is invalid", zap.Error(err))
	}

	var transport sdk.Transport
	switch addr.Scheme {
	case "http":
		transport = sdkhttp.NewTransport(c.address)
	case "grpc":
		conn, err := grpc.Dial(addr.Host,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Fatal("Failed to dial Flipt", zap.Error(err))
		}

		transport = sdkgrpc.NewTransport(conn)
	}

	var opts []sdk.Option
	if c.token != "" {
		opts = append(opts, sdk.WithClientTokenProvider(
			sdk.StaticClientTokenProvider(c.token),
		))
	}

	client := sdk.New(transport, opts...)

	return c.export(cmd.Context(), logger, client.Flipt())
}

func (c *exportCommand) export(ctx context.Context, logger *zap.Logger, lister ext.Lister) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-interrupt
		cancel()
	}()

	// default to stdout
	var out io.WriteCloser = os.Stdout

	// export to file
	if c.filename != "" {
		logger.Debug("exporting", zap.String("destination_path", c.filename))

		var err error
		out, err = os.Create(c.filename)
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
