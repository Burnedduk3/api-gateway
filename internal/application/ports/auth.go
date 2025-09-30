package ports

import (
	"api-gateway/internal/domain/entities"
	"context"
)

type AuthValidator interface {
	Validate(ctx context.Context, token string, policy *entities.AuthPolicy) error
	ExtractToken(ctx context.Context, headers map[string][]string, authType string) (string, error)
}
