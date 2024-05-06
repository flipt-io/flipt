package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/storage/sql"
	"go.uber.org/zap"
)

const (
	defaultConfig = "default"
	analytics     = "analytics"
)

var database string

func runMigrations(cfg *config.Config, logger *zap.Logger, database string) error {
	var (
		migrator *sql.Migrator
		err      error
	)

	if database == analytics {
		migrator, err = sql.NewAnalyticsMigrator(*cfg, logger)
		if err != nil {
			return err
		}
	} else {
		migrator, err = sql.NewMigrator(*cfg, logger)
		if err != nil {
			return err
		}
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
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := cmd.Context()
			logger, cfg, err := buildConfig(ctx)
			if err != nil {
				return err
			}

			defer func() {
				_ = logger.Sync()
			}()

			// Run the OLTP and OLAP database migrations sequentially because of
			// potential danger in DB migrations in general.
			if err := runMigrations(cfg, logger, defaultConfig); err != nil {
				return err
			}

			if database == analytics {
				if err := runMigrations(cfg, logger, analytics); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&providedConfigFile, "config", "", "path to config file")
	cmd.Flags().StringVar(&database, "database", "default", "string to denote which database type to migrate")
	return cmd
}
