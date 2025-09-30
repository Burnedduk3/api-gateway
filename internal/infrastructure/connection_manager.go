package infrastructure

import (
	"api-gateway/internal/adapters/persistence/repositories"
	"api-gateway/internal/application/ports"
	"context"
	"fmt"
	"time"

	"api-gateway/internal/config"
	"api-gateway/pkg/logger"

	"github.com/redis/go-redis/v9"
)

type DatabaseConnections struct {
	logger logger.Logger
	redis  ports.ApiKeyRepository
}

func NewDatabaseConnections(cfg *config.Config, logger logger.Logger) (*DatabaseConnections, error) {
	log := logger.With("component", "database_connections")
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%s", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.Database,
		MaxRetries:   cfg.Redis.MaxRetries,
		PoolSize:     cfg.Redis.PoolSize,
		DialTimeout:  cfg.Redis.DialTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		WriteTimeout: cfg.Redis.WriteTimeout,
	})
	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	logger.Info("Redis connection established",
		"host", cfg.Redis.Host,
		"port", cfg.Redis.Port)

	redisRepo := repositories.NewRedisApiKeyRepository(client, log)

	log.Info("All database connections established successfully")
	return &DatabaseConnections{
		logger: log,
		redis:  redisRepo,
	}, nil
}

func (d *DatabaseConnections) HealthCheck(ctx context.Context) map[string]error {
	checks := make(map[string]error)
	checks["redis"] = d.redis.HealthCheck(ctx)
	return checks
}

func (d *DatabaseConnections) GetApiKeyRepo() ports.ApiKeyRepository {
	return d.redis
}
