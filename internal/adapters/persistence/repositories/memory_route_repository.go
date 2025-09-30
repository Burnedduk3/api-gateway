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

	if route.Backend == nil {
		return errors.New("route backend is required")
	}

	backendID := route.Backend.Id
	repo.backends[backendID] = append(repo.backends[backendID], *route)

	repo.log.Debug("Route saved",
		"id", route.ID,
		"path", route.Path,
		"method", route.Method,
		"backend_id", backendID,
	)

	return nil
}

func (repo *MemoryRouteRepo) FindByPathAndMethod(ctx context.Context, path, method string) (*entities.Route, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	repo.log.Debug("Looking for route",
		"path", path,
		"method", method,
		"backends_count", len(repo.backends),
	)

	// Iterate over backend IDs and their routes
	for backendID, routes := range repo.backends {
		repo.log.Debug("Checking backend",
			"backend_id", backendID,
			"routes_count", len(routes),
		)

		for i := range routes {
			route := &routes[i]

			repo.log.Debug("Checking route",
				"route_id", route.ID,
				"route_path", route.Path,
				"route_method", route.Method,
				"route_enabled", route.Enabled,
			)

			// Pass backend ID to match
			if route.Match(path, method, backendID) && route.IsEnabled() {
				repo.log.Info("Route matched",
					"route_id", route.ID,
					"route_path", route.Path,
					"backend_id", backendID,
				)
				return route, nil
			}
		}
	}

	repo.log.Warn("No route found",
		"path", path,
		"method", method,
	)

	return nil, errors.New("route not found")
}
