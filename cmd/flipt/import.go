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
	"go.uber.org/zap"
)

type importCommand struct {
	dropBeforeImport bool
	importStdin      bool
	address          string
	token            string
}

func newImportCommand() *cobra.Command {
	importCmd := &importCommand{}

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import flags/segments/rules from file",
		RunE:  importCmd.run,
	}

	cmd.Flags().BoolVar(
		&importCmd.dropBeforeImport,
		"drop",
		false,
		"drop database before import",
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
		"address of remote Flipt instance to import into (defaults to direct DB import if not supplied)",
	)

	cmd.Flags().StringVarP(
		&importCmd.token,
		"token", "t",
		"",
		"client token used to authenticate access to remote Flipt instance when importing.",
	)

	return cmd
}

func (c *importCommand) run(cmd *cobra.Command, args []string) error {
	var (
		in     io.Reader = os.Stdin
		logger           = zap.Must(zap.NewDevelopment())
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

		logger.Debug("importing", zap.String("source_path", f))

		fi, err := os.Open(f)
		if err != nil {
			return fmt.Errorf("opening import file: %w", err)
		}

		defer fi.Close()

		in = fi
	}

	// Use client when remote address is configured.
	if c.address != "" {
		return ext.NewImporter(
			fliptClient(logger, c.address, c.token),
		).Import(cmd.Context(), in)
	}

	logger, cfg := buildConfig()

	// drop tables if specified
	if c.dropBeforeImport {
		logger.Debug("dropping tables")

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
	).Import(cmd.Context(), in)
}
