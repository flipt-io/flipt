package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/storage/sql"
)

func newMigrateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run pending database migrations",
		RunE: func(_ *cobra.Command, _ []string) error {
			logger, cfg, err := buildConfig()
			if err != nil {
				return err
			}

			defer func() {
				_ = logger.Sync()
			}()

			migrator, err := sql.NewMigrator(*cfg, logger)
			if err != nil {
				return fmt.Errorf("initializing migrator %w", err)
			}

			defer migrator.Close()

			if err := migrator.Up(true); err != nil {
				return fmt.Errorf("running migrator %w", err)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&providedConfigFile, "config", "", "path to config file")
	return cmd
}
