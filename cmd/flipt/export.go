package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/rpc/flipt"
	"go.uber.org/zap"
)

type exportCommand struct {
	filename      string
	address       string
	token         string
	namespaces    string // comma delimited list of namespaces
	allNamespaces bool
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
		"address", "a",
		"",
		"address of remote Flipt instance to export from (defaults to direct DB export if not supplied)",
	)

	cmd.Flags().StringVarP(
		&export.token,
		"token", "t",
		"",
		"client token used to authenticate access to remote Flipt instance when exporting.",
	)

	cmd.Flags().StringVarP(
		&export.namespaces,
		"namespaces", "n",
		flipt.DefaultNamespace,
		"comma delimited list of namespaces to export from.",
	)

	cmd.Flags().BoolVar(
		&export.allNamespaces,
		"all-namespaces",
		false,
		"boolean flag to specify if every namespace should be exported.",
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
		client, err := fliptClient(c.address, c.token)
		if err != nil {
			return err
		}
		return c.export(cmd.Context(), out, client)
	}

	// Otherwise, go direct to the DB using Flipt configuration file.
	logger, cfg, err := buildConfig()
	if err != nil {
		return err
	}

	defer func() {
		_ = logger.Sync()
	}()

	server, cleanup, err := fliptServer(logger, cfg)
	if err != nil {
		return err
	}

	defer cleanup()

	return c.export(cmd.Context(), out, server)
}

func (c *exportCommand) export(ctx context.Context, dst io.Writer, lister ext.Lister) error {
	return ext.NewExporter(lister, c.namespaces, c.allNamespaces).Export(ctx, dst)
}
