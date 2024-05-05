package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/rpc/flipt"
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
		Short: "Export Flipt data to file/stdout",
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
		"address of Flipt instance (defaults to direct DB export if not supplied).",
	)

	cmd.Flags().StringVarP(
		&export.token,
		"token", "t",
		"",
		"client token used to authenticate access to Flipt instance.",
	)

	cmd.Flags().StringVarP(
		&export.namespaces,
		"namespace", "n",
		flipt.DefaultNamespace,
		"source namespace for exported resources.",
	)

	cmd.Flags().StringVar(
		&export.namespaces,
		"namespaces",
		flipt.DefaultNamespace,
		"comma-delimited list of namespaces to export from. (mutually exclusive with --all-namespaces)",
	)

	cmd.Flags().BoolVar(
		&export.allNamespaces,
		"all-namespaces",
		false,
		"export all namespaces. (mutually exclusive with --namespaces)",
	)

	cmd.Flags().StringVar(&providedConfigFile, "config", "", "path to config file")

	cmd.MarkFlagsMutuallyExclusive("all-namespaces", "namespaces", "namespace")

	// We can ignore the error here since "namespace" will be a flag that exists.
	_ = cmd.Flags().MarkDeprecated("namespace", "please use namespaces instead")

	return cmd
}

func (c *exportCommand) run(cmd *cobra.Command, _ []string) error {
	var (
		ctx = cmd.Context()
		// default to stdout
		out io.Writer = os.Stdout
		enc           = ext.EncodingYML
	)

	// export to file
	if c.filename != "" {
		fi, err := os.Create(c.filename)
		if err != nil {
			return fmt.Errorf("creating output file: %w", err)
		}

		defer fi.Close()

		fmt.Fprintf(fi, "# exported by Flipt (%s) on %s\n\n", version, time.Now().UTC().Format(time.RFC3339))

		out = fi

		if extn := filepath.Ext(c.filename); len(extn) > 0 {
			// strip off the leading .
			enc = ext.Encoding(extn[1:])
		}
	}

	// Use client when remote address is configured.
	if c.address != "" {
		client, err := fliptClient(c.address, c.token)
		if err != nil {
			return err
		}
		return c.export(ctx, enc, out, client)
	}

	// Otherwise, go direct to the DB using Flipt configuration file.
	logger, cfg, err := buildConfig(ctx)
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

	return c.export(ctx, enc, out, server)
}

func (c *exportCommand) export(ctx context.Context, enc ext.Encoding, dst io.Writer, lister ext.Lister) error {
	return ext.NewExporter(lister, c.namespaces, c.allNamespaces).Export(ctx, enc, dst)
}
