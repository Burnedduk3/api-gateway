/*
Copyright © 2025 Juan David Cabrera Duran juandavid.juandis@gmail.com
*/
package cmd

import (
	"api-gateway/internal/adapters/http"
	"api-gateway/internal/config"
	"api-gateway/internal/infrastructure"
	"api-gateway/pkg/logger"
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"
)

// serverCmd represents the server command
var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the HTTP server",
	Long:  "Start the user service HTTP server with Echo framework",
	RunE:  runServer,
}

func init() {
	rootCmd.AddCommand(serverCmd)

	// Add server-specific flags
	serverCmd.Flags().StringVarP(&port, "port", "p", "", "server port")
}

func runServer(cmd *cobra.Command, args []string) error {
	// Initialize logging
	log := logger.New(env)

	log.Info("Starting Identity Service...")

	// Load configuration
	cfg, err := config.Load(configFile, env)
	if err != nil {
		log.Fatal("Failed to load configuration", "error", err)
		return err
	}

	// Override port if provided via flag
	if cmd.Flags().Changed("port") {
		cfg.Server.Port = port
		log.Info("Port overridden by command line flag", "port", port)
	}

	log.Info("Configuration loaded",
		"env", cfg.Environment,
		"port", cfg.Server.Port,
		"log_level", cfg.Logging.Level)

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

	// Create HTTP server with database connections
	server, err := http.NewServer(cfg, log, connections) // Updated
	if err != nil {
		log.Fatal("Failed to create server", "error", err)
		return err
	}

	// Start server in goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Fatal("Server failed to start", "error", err)
		}
	}()

	log.Info("Server started successfully", "port", cfg.Server.Port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown", "error", err)
		return err
	}

	log.Info("Server exited")
	return nil
}
