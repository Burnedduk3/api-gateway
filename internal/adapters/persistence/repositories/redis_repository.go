package repositories

import (
	"api-gateway/internal/application/ports"
	"api-gateway/pkg/logger"
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisApiKeyRepository struct {
	client *redis.Client
	log    logger.Logger
}

func NewRedisApiKeyRepository(client *redis.Client, log logger.Logger) ports.ApiKeyRepository {
	return &RedisApiKeyRepository{
		client: client,
		log:    log,
	}
}

// HealthCheck Health check implementation
func (r *RedisApiKeyRepository) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	err := r.client.Ping(ctx).Err()
	if err != nil {
		r.log.Error("Redis health check failed", "error", err)
		return fmt.Errorf("redis health check failed: %w", err)
	}

	r.client.Set(ctx, "key-123", true, 0)
	r.client.Set(ctx, "key-1234", false, 0)

	return nil
}

// IsValidKey checks if an API key exists and is valid in Redis
func (r *RedisApiKeyRepository) IsValidKey(ctx context.Context, key string) (bool, error) {
	// Check if key exists in Redis
	exists, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		r.log.Error("Failed to check API key in Redis", "error", err)
		return false, err
	}

	if exists == 0 {
		return false, nil
	}

	isValidKey := r.client.Get(ctx, key).Val()

	return isValidKey == "1", nil
}

// GetKeyMetadata retrieves metadata for an API key
func (r *RedisApiKeyRepository) GetKeyMetadata(ctx context.Context, key string) (map[string]interface{}, error) {
	redisKey := fmt.Sprintf("apikey:%s", key)

	// Get all fields from hash
	data, err := r.client.HGetAll(ctx, redisKey).Result()
	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("API key not found")
	}

	// Convert to map[string]interface{}
	metadata := make(map[string]interface{})
	for k, v := range data {
		metadata[k] = v
	}

	return metadata, nil
}

// StoreKey stores an API key in Redis with metadata
func (r *RedisApiKeyRepository) StoreKey(ctx context.Context, key string, metadata map[string]interface{}) error {
	redisKey := fmt.Sprintf("apikey:%s", key)

	// Convert metadata to map[string]interface{} for HMSET
	data := make(map[string]interface{})
	for k, v := range metadata {
		data[k] = v
	}

	// Store as hash
	if err := r.client.HSet(ctx, redisKey, data).Err(); err != nil {
		return err
	}

	r.log.Info("API key stored", "key", key)
	return nil
}

// RevokeKey removes an API key from Redis
func (r *RedisApiKeyRepository) RevokeKey(ctx context.Context, key string) error {
	redisKey := fmt.Sprintf("apikey:%s", key)

	if err := r.client.Del(ctx, redisKey).Err(); err != nil {
		return err
	}

	r.log.Info("API key revoked", "key", key)
	return nil
}

// Close closes the Redis connection
func (r *RedisApiKeyRepository) Close() error {
	return r.client.Close()
}
