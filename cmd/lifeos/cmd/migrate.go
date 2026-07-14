package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/valentinezhov/lifeos/internal/platform/config"
	platformmigrate "github.com/valentinezhov/lifeos/internal/platform/migrate"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database migrations",
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations",
	Run: func(_ *cobra.Command, _ []string) {
		exitOnError(runMigrateUp())
	},
}

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback the last migration",
	Run: func(_ *cobra.Command, _ []string) {
		exitOnError(runMigrateDown())
	},
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show migration status",
	Run: func(_ *cobra.Command, _ []string) {
		exitOnError(runMigrateStatus())
	},
}

func init() {
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateStatusCmd)
}

func runMigrateUp() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	return platformmigrate.Up(context.Background(), cfg.DatabaseURL, cfg.MigrationsDir)
}

func runMigrateDown() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	return platformmigrate.Down(context.Background(), cfg.DatabaseURL, cfg.MigrationsDir)
}

func runMigrateStatus() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	return platformmigrate.Status(context.Background(), cfg.DatabaseURL, cfg.MigrationsDir)
}
