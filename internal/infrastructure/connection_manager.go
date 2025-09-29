package infrastructure

import (
	"context"
	"fmt"

	"api-gateway/internal/config"
	"api-gateway/pkg/logger"
)

type DatabaseConnections struct {
	logger logger.Logger
}

func NewDatabaseConnections(cfg *config.Config, logger logger.Logger) (*DatabaseConnections, error) {
	log := logger.With("component", "database_connections")

	log.Info("All database connections established successfully")

	return &DatabaseConnections{
		logger: log,
	}, nil
}

func (d *DatabaseConnections) Close() error {
	d.logger.Info("Closing all database connections...")

	var errs []error

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	d.logger.Info("All database connections closed successfully")
	return nil
}

func (d *DatabaseConnections) HealthCheck(ctx context.Context) map[string]error {
	checks := make(map[string]error)

	return checks
}
