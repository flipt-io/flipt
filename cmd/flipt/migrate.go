package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/storage/analytics"
	"go.uber.org/zap"
)

func runMigrations(cfg *config.Config, logger *zap.Logger) error {
	var (
		migrator *analytics.Migrator
		err      error
	)

	migrator, err = analytics.NewMigrator(*cfg, logger)
	if err != nil {
		return err
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

			if err := runMigrations(cfg, logger); err != nil {
				return err
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&providedConfigFile, "config", "", "path to config file")
	return cmd
}
