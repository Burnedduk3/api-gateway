package infrastructure

import (
	"context"
	"fmt"

	gormConn "api-gateway/internal/adapters/persistence/postgres"
	"api-gateway/internal/config"
	"api-gateway/pkg/logger"

	"gorm.io/gorm"
)

type DatabaseConnections struct {
	conn   *gormConn.GormDB
	logger logger.Logger
}

func NewDatabaseConnections(cfg *config.Config, logger logger.Logger) (*DatabaseConnections, error) {
	log := logger.With("component", "database_connections")

	// PostgreSQL connection
	log.Info("Connecting to PostgreSQL...")
	pg, err := gormConn.NewGormConnection(cfg, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gormConn: %w", err)
	}

	log.Info("All database connections established successfully")

	return &DatabaseConnections{
		conn:   pg,
		logger: log,
	}, nil
}

func (d *DatabaseConnections) Close() error {
	d.logger.Info("Closing all database connections...")

	var errs []error

	if err := d.conn.Close(); err != nil {
		errs = append(errs, fmt.Errorf("gormConn close error: %w", err))
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}

	d.logger.Info("All database connections closed successfully")
	return nil
}

func (d *DatabaseConnections) HealthCheck(ctx context.Context) map[string]error {
	checks := make(map[string]error)

	checks["postgress"] = d.conn.HealthCheck(ctx)

	return checks
}

func (d *DatabaseConnections) GetGormDB() *gorm.DB {
	return d.conn.DB()
}
