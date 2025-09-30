package ports

import (
	"api-gateway/internal/domain/entities"
	"context"
)

type RouteRepository interface {
	FindByPathAndMethod(ctx context.Context, path, method string) (*entities.Route, error)
	GetAll(ctx context.Context) ([]entities.Route, error)
	Save(ctx context.Context, route *entities.Route) error
}
