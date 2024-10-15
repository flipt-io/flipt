package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/internal/storage/sql"
	"go.flipt.io/flipt/rpc/flipt"
)

type importCommand struct {
	dropBeforeImport bool
	skipExisting     bool
	importStdin      bool
	address          string
	token            string
}

func newImportCommand() *cobra.Command {
	importCmd := &importCommand{}

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import Flipt data from file/stdin",
		RunE:  importCmd.run,
	}

	cmd.Flags().BoolVar(
		&importCmd.dropBeforeImport,
		"drop",
		false,
		"drop database before import",
	)
	cmd.Flags().BoolVar(
		&importCmd.skipExisting,
		"skip-existing",
		false,
		"only import new data",
	)

	cmd.Flags().BoolVar(
		&importCmd.importStdin,
		"stdin",
		false,
		"import from STDIN",
	)

	cmd.Flags().StringVarP(
		&importCmd.address,
		"address", "a",
		"",
		"address of Flipt instance (defaults to direct DB import if not supplied).",
	)

	cmd.Flags().StringVarP(
		&importCmd.token,
		"token", "t",
		"",
		"client token used to authenticate access to Flipt instance.",
	)

	cmd.Flags().StringVar(&providedConfigFile, "config", "", "path to config file")
	return cmd
}

func (c *importCommand) run(cmd *cobra.Command, args []string) error {
	var (
		ctx           = cmd.Context()
		in  io.Reader = os.Stdin
		enc           = ext.EncodingYML
	)

	if !c.importStdin {
		if len(args) < 1 {
			return errors.New("import filename required")
		}

		importFilename := args[0]
		if importFilename == "" {
			return errors.New("import filename required")
		}

		f := filepath.Clean(importFilename)

		fi, err := os.Open(f)
		if err != nil {
			return fmt.Errorf("opening import file: %w", err)
		}

		defer fi.Close()

		in = fi

		if extn := filepath.Ext(importFilename); len(extn) > 0 {
			// strip off leading .
			enc = ext.Encoding(extn[1:])
		}
	}

	// Use client when remote address is configured.
	if c.address != "" {
		client, err := fliptClient(c.address, c.token)
		if err != nil {
			return err
		}
		if c.dropBeforeImport {
			err = client.DeleteAllNamespaces(ctx, &flipt.DeleteAllNamespacesRequest{})
			if err != nil {
				return fmt.Errorf("deleting all namespaces: %w", err)
			}
		}
		return ext.NewImporter(client).Import(ctx, enc, in, c.skipExisting)
	}

	logger, cfg, err := buildConfig(ctx)
	if err != nil {
		return err
	}

	defer func() {
		_ = logger.Sync()
	}()

	// drop tables if specified
	if c.dropBeforeImport {

		migrator, err := sql.NewMigrator(*cfg, logger)
		if err != nil {
			return err
		}

		if err := migrator.Drop(); err != nil {
			return fmt.Errorf("attempting to drop: %w", err)
		}

		if _, err := migrator.Close(); err != nil {
			return fmt.Errorf("closing migrator: %w", err)
		}
	}

	migrator, err := sql.NewMigrator(*cfg, logger)
	if err != nil {
		return err
	}

	if err := migrator.Up(forceMigrate); err != nil {
		return err
	}

	if _, err := migrator.Close(); err != nil {
		return fmt.Errorf("closing migrator: %w", err)
	}

	// Otherwise, go direct to the DB using Flipt configuration file.
	server, cleanup, err := fliptServer(logger, cfg)
	if err != nil {
		return err
	}

	defer cleanup()

	return ext.NewImporter(
		server,
	).Import(ctx, enc, in, c.skipExisting)
}
