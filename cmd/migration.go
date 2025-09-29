/*
Copyright © 2025 Juan David Cabrera Duran juandavid.juandis@gmail.com
*/
package cmd

import (
	"fmt"
	"product-service/internal/adapters/persistence/product_repository"

	"product-service/internal/config"
	"product-service/internal/infrastructure"
	"product-service/pkg/logger"

	"github.com/spf13/cobra"
)

var (
	dryRun bool
)

// migrationCmd represents the migration command
var migrationCmd = &cobra.Command{
	Use:   "migration",
	Short: "Run database migrations",
	Long: `Run database migrations to create and update tables.

This command will:
- Create missing tables
- Add new columns to existing tables  
- Update column types if needed
- Create indexes

Examples:
  # Run migrations
  user-service migration

  # Preview what would be executed
  user-service migration --dry-run`,
	RunE: runMigration,
}

func init() {
	rootCmd.AddCommand(migrationCmd)
}

func runMigration(cmd *cobra.Command, args []string) error {
	// Initialize logging
	log := logger.New(env)

	log.Info("Starting database migration...")

	// Load configuration
	cfg, err := config.Load(configFile, env)
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
		return err
	}

	log.Info("Configuration loaded",
		"env", cfg.Environment,
		"database", cfg.Database.Database)

	// Initialize database connections
	log.Info("Initializing database connections...")
	connections, err := infrastructure.NewDatabaseConnections(cfg, log)
	if err != nil {
		log.Fatal("Failed to initialize database connections", "error", err)
		return err
	}

	// Ensure connections are closed on exit
	defer func() {
		if err := connections.Close(); err != nil {
			log.Error("Failed to close database connections", "error", err)
		}
	}()

	log.Info("Database connection established successfully")

	if err := runDatabaseMigrations(connections, log); err != nil {
		log.Error("Migration failed", "error", err)
		return err
	}

	log.Info("Database migration completed successfully")
	return nil
}

func runDatabaseMigrations(connections *infrastructure.DatabaseConnections, log logger.Logger) error {
	// Get the GORM database instance
	db := connections.GetGormDB()

	models := getAllModels()

	log.Info("Running AutoMigrate", "models_count", len(models))

	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("failed to run AutoMigrate: %w", err)
	}

	log.Info("All migrations completed successfully")
	return nil
}

// getAllModels returns all database models that need migration
func getAllModels() []interface{} {
	return []interface{}{
		&product_repository.ProductModel{},
	}
}
