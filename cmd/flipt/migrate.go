package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/storage/sql"
	"go.uber.org/zap"
)

var (
	analytics bool
)

func runMigrations(cfg *config.Config, logger *zap.Logger, analytics bool) error {
	migrator, err := sql.NewMigrator(*cfg, logger, analytics)
	if err != nil {
		return fmt.Errorf("initializing migrator %w", err)
	}

	defer migrator.Close()

	if err := migrator.Up(true); err != nil {
		return fmt.Errorf("running migrator %w", err)
	}

	return nil
}

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

			// Run the OLTP and OLAP database migrations sequentially because of
			// potential danger in DB migrations in general.
			if err := runMigrations(cfg, logger, false); err != nil {
				return err
			}

			if analytics {
				if err := runMigrations(cfg, logger, true); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&providedConfigFile, "config", "", "path to config file")
	cmd.Flags().BoolVar(&analytics, "analytics", false, "migrate analytics database")
	return cmd
}
