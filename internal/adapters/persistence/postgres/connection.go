// internal/adapters/persistence/connection.go
package persistence

import (
	"context"
	"fmt"
	"time"

	"api-gateway/internal/config"
	"api-gateway/pkg/logger"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type GormDB struct {
	db     *gorm.DB
	logger logger.Logger
}

func NewGormConnection(cfg *config.Config, log logger.Logger) (*GormDB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port, cfg.Database.Username, cfg.Database.Password, cfg.Database.Database, cfg.Database.SSLMode)
	// Configure GORM with your zap logger
	gormLogLevel := StringToGormLogLevel(cfg.LogLevel)

	customLogger := NewGormZapLoggerWithConfig(log, GormLoggerConfig{
		LogLevel:                  gormLogLevel,
		IgnoreRecordNotFoundError: true,
		SlowThreshold:             200 * time.Millisecond,
	})

	gormConfig := &gorm.Config{
		Logger: customLogger,
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
	}

	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres with GORM: %w", err)
	}

	// Get underlying sql.DB to configure connection pool
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	// Configure connection pool
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.MaxLifetime)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	log.Info("GORM PostgreSQL connection established",
		"host", cfg.Database.Host,
		"port", cfg.Database.Port,
		"database", cfg.Database.Database,
		"max_open_conns", cfg.Database.MaxOpenConns)

	return &GormDB{
		db:     db,
		logger: log.With("component", "gorm"),
	}, nil
}

func (g *GormDB) DB() *gorm.DB {
	return g.db
}

func (g *GormDB) Close() error {
	g.logger.Info("Closing GORM PostgreSQL connection")
	sqlDB, err := g.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// Health check implementation
func (g *GormDB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	sqlDB, err := g.db.DB()
	if err != nil {
		g.logger.Error("Failed to get underlying sql.DB for health check", "error", err)
		return fmt.Errorf("gorm health check failed: %w", err)
	}

	if err := sqlDB.PingContext(ctx); err != nil {
		g.logger.Error("GORM PostgreSQL health check failed", "error", err)
		return fmt.Errorf("gorm postgres health check failed: %w", err)
	}

	return nil
}

// AutoMigrate runs database migrations
func (g *GormDB) AutoMigrate(models ...interface{}) error {
	g.logger.Info("Running database migrations")

	if err := g.db.AutoMigrate(models...); err != nil {
		g.logger.Error("Database migration failed", "error", err)
		return fmt.Errorf("database migration failed: %w", err)
	}

	g.logger.Info("Database migrations completed successfully")
	return nil
}

// Transaction helper for GORM
func (g *GormDB) WithTransaction(ctx context.Context, fn func(tx *gorm.DB) error) error {
	return g.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
