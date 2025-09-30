package ports

import "context"

type ApiKeyRepository interface {
	// IsValidKey checks if an API key is valid
	IsValidKey(ctx context.Context, key string) (bool, error)
	HealthCheck(ctx context.Context) error
	// GetKeyMetadata returns metadata about the key (user, permissions, etc.)
	GetKeyMetadata(ctx context.Context, key string) (map[string]interface{}, error)

	StoreKey(ctx context.Context, key string, metadata map[string]interface{}) error
	RevokeKey(ctx context.Context, key string) error
}
