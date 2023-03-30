package main

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
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
	var (
		// default to stdout
		out    io.Writer = os.Stdout
		logger           = zap.Must(zap.NewDevelopment())
	)

	// export to file
	if c.filename != "" {
		logger.Debug("exporting", zap.String("destination_path", c.filename))

		fi, err := os.Create(c.filename)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}

		defer fi.Close()

		fmt.Fprintf(fi, "# exported by Flipt (%s) on %s\n\n", version, time.Now().UTC().Format(time.RFC3339))

		out = fi
	}

	// Use client when remote address is configured.
	if c.address != "" {
		return export(cmd.Context(), out, fliptClient(logger, c.address, c.token))
	}

	server, err := fliptServer(false)
	if err != nil {
		return err
	}

	return export(cmd.Context(), out, server)
}

func export(ctx context.Context, dst io.Writer, lister ext.Lister) error {
	return ext.NewExporter(lister, storage.DefaultNamespace).Export(ctx, dst)
}

func fliptServer(dropBeforeMigrate bool) (*server.Server, error) {
	// Otherwise, go direct to the DB using Flipt configuration file.
	logger, cfg := buildConfig()

	db, driver, err := sql.Open(*cfg)
	if err != nil {
		return nil, fmt.Errorf("opening db: %w", err)
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

	migrator, err := sql.NewMigrator(*cfg, logger)
	if err != nil {
		return nil, err
	}

	defer migrator.Close()

	// drop tables if specified
	if dropBeforeMigrate {
		logger.Debug("dropping tables")

		if err := migrator.Down(); err != nil {
			return nil, fmt.Errorf("attempting to drop: %w", err)
		}
	}

	if err := migrator.Up(forceMigrate); err != nil {
		return nil, err
	}

	if _, err := migrator.Close(); err != nil {
		return nil, fmt.Errorf("closing migrator: %w", err)
	}

	return server.New(logger, store), nil
}

func fliptClient(logger *zap.Logger, address, token string) *sdk.Flipt {
	addr, err := url.Parse(address)
	if err != nil {
		logger.Fatal("Export address is invalid", zap.Error(err))
	}

	var transport sdk.Transport
	switch addr.Scheme {
	case "http":
		transport = sdkhttp.NewTransport(address)
	case "grpc":
		conn, err := grpc.Dial(addr.Host,
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			logger.Fatal("Failed to dial Flipt", zap.Error(err))
		}

		transport = sdkgrpc.NewTransport(conn)
	}

	var opts []sdk.Option
	if token != "" {
		opts = append(opts, sdk.WithClientTokenProvider(
			sdk.StaticClientTokenProvider(token),
		))
	}

	return sdk.New(transport, opts...).Flipt()
}
