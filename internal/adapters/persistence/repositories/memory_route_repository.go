package repositories

import (
	"api-gateway/internal/domain/entities"
	"api-gateway/pkg/logger"
	"context"
	"errors"
	"sync"
)

type MemoryRouteRepo struct {
	backends map[string][]entities.Route
	mu       sync.RWMutex
	log      logger.Logger
}

func NewMemoryRouteRepo(log logger.Logger) *MemoryRouteRepo {
	return &MemoryRouteRepo{
		backends: make(map[string][]entities.Route),
		log:      log,
	}
}

func (repo *MemoryRouteRepo) GetAll(ctx context.Context) ([]entities.Route, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	allRoutes := make([]entities.Route, 0, len(repo.backends))
	for _, routes := range repo.backends {
		allRoutes = append(allRoutes, routes...)
	}

	return allRoutes, nil
}

func (repo *MemoryRouteRepo) Save(ctx context.Context, route *entities.Route) error {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if err := route.Validate(); err != nil {
		return err
	}

	if route.ID == "" {
		return errors.New("route ID is required")
	}

	repo.backends[route.Backend.Id] = append(repo.backends[route.Backend.Id], *route)

	repo.log.Debug("Route saved", "id", route.ID, "path", route.Path, "method", route.Method)

	return nil
}

func (repo *MemoryRouteRepo) FindByPathAndMethod(ctx context.Context, path, method string) (*entities.Route, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()
	// Loop through all stored routes
	for _, routes := range repo.backends {
		for _, route := range routes {
			if !route.IsEnabled() {
				continue
			}

			// Use the route's Match method to check if it matches
			if route.Match(path, method) {
				return &route, nil
			}
		}

	}

	return nil, errors.New("route not found")
}
