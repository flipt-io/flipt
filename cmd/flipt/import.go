package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/ext"
	"go.flipt.io/flipt/internal/storage"
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

	cmd.Flags().StringVar(
		&importCmd.address,
		"import-to-address",
		"",
		"address of remote Flipt instance to import intp (defaults to direct DB import if not supplied)",
	)

	cmd.Flags().StringVar(
		&importCmd.token,
		"import-to-token",
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
			storage.DefaultNamespace,
		).Import(cmd.Context(), in)
	}

	// Otherwise, go direct to the DB using Flipt configuration file.
	server, err := fliptServer(c.dropBeforeImport)
	if err != nil {
		return err
	}

	return ext.NewImporter(server, storage.DefaultNamespace).Import(cmd.Context(), in)
}
