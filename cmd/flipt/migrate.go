package main

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/config"
	"go.flipt.io/flipt/internal/storage/sql"
	"go.uber.org/zap"
)

const (
	// Default is the default database
	defaultConfig = "default"
	// Analytics is the analytics database
	analytics = "analytics"
)

func runMigrations(cfg *config.Config, logger *zap.Logger, database string) error {
	var err error
	if database == analytics {
		logger.Info("analytics migrations running")
		logger.Debug("migrating analytics database")

		migrator, err := sql.NewAnalyticsMigrator(*cfg, logger)
		if err != nil {
			return fmt.Errorf("creating migrator: %w", err)
		}

		if err := migrator.Up(false); err != nil {
			return fmt.Errorf("running migrator %w", err)
		}

		if _, err := migrator.Close(); err != nil {
			return fmt.Errorf("closing migrator: %w", err)
		}

		return nil
	}

	logger.Info("migrations running")
	logger.Debug("migrating config/default database")

	migrator, err := sql.NewMigrator(*cfg, logger)
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}

	if err := migrator.Up(false); err != nil {
		return fmt.Errorf("running migrator %w", err)
	}

	return nil
}

type migrateCommand struct {
	root     *rootCommand
	database string
}

func newMigrateCommand(root *rootCommand) *cobra.Command {
	migrateCmd := &migrateCommand{
		root:     root,
		database: defaultConfig,
	}

	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run pending database migrations",
		RunE:  migrateCmd.run,
	}

	cmd.Flags().StringVar(&migrateCmd.root.configFile, "config", migrateCmd.root.configFile, "path to config file")
	cmd.Flags().StringVar(&migrateCmd.database, "database", defaultConfig, "string to denote which database type to migrate")

	return cmd
}

func (c *migrateCommand) run(cmd *cobra.Command, _ []string) error {
	ctx := cmd.Context()

	logger, cfg, err := c.root.buildConfig(ctx)
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

	if c.database == analytics {
		if err := runMigrations(cfg, logger, analytics); err != nil {
			return err
		}
	}

	return nil
}
