package main

import (
	"github.com/spf13/cobra"
	"go.flipt.io/flipt/internal/storage/sql"
	"go.uber.org/zap"
)

func newMigrateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "migrate",
		Short: "Run pending database migrations",
		Run: func(cmd *cobra.Command, _ []string) {
			logger, cfg := buildConfig()
			defer func() {
				_ = logger.Sync()
			}()

			migrator, err := sql.NewMigrator(*cfg, logger)
			if err != nil {
				logger.Fatal("initializing migrator", zap.Error(err))
			}

			defer migrator.Close()

			if err := migrator.Up(true); err != nil {
				logger.Fatal("running migrator", zap.Error(err))
			}
		},
	}
}
