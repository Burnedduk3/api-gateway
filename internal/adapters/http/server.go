package http

import (
	"api-gateway/internal/adapters/auth"
	"api-gateway/internal/adapters/http/handlers"
	"api-gateway/internal/adapters/http/middlewares/logging"
	"api-gateway/internal/adapters/http/middlewares/security"
	"api-gateway/internal/adapters/persistence/repositories"
	"api-gateway/internal/application/usecases"
	"api-gateway/internal/config"
	"api-gateway/internal/domain/entities"
	"api-gateway/internal/infrastructure"
	"api-gateway/pkg/logger"
	"context"
	"fmt"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

type Server struct {
	echo        *echo.Echo
	config      *config.Config
	logger      logger.Logger
	connections *infrastructure.DatabaseConnections
}

func NewServer(cfg *config.Config, log logger.Logger, connections *infrastructure.DatabaseConnections) (*Server, error) {
	e := echo.New()

	// Configure Echo
	e.HideBanner = true
	e.HidePort = true

	server := &Server{
		echo:        e,
		config:      cfg,
		logger:      log,
		connections: connections,
	}

	// Setup middleware
	server.setupMiddleware()

	// Setup routes
	server.setupRoutes(cfg)

	return server, nil
}

func (s *Server) setupMiddleware() {
	// Replace Echo's logger with our custom Zap logger
	s.echo.Use(logging.ZapLogger(s.logger.With("component", "http")))

	// Security headers
	s.echo.Use(middleware.SecureWithConfig(middleware.SecureConfig{
		XSSProtection:         "1; mode=block",
		ContentTypeNosniff:    "nosniff",
		XFrameOptions:         "DENY",
		HSTSMaxAge:            31536000,
		ContentSecurityPolicy: "default-src 'self'",
	}))

	// CORS middleware
	s.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: s.config.Server.CORS.AllowOrigins,
		AllowMethods: s.config.Server.CORS.AllowMethods,
		AllowHeaders: s.config.Server.CORS.AllowHeaders,
	}))

	// Request timeout middleware
	s.echo.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: s.config.Server.ReadTimeout,
	}))
}

func (s *Server) setupRoutes(cfg *config.Config) {
	// Health check handlers with database connections
	routes := s.parseRoutes(cfg)
	ctx := context.Background()
	healthHandler := handlers.NewHealthHandler(s.logger, s.connections)
	proxyClientRepo := handlers.NewProxyClient(60*time.Second, s.logger)
	memoryRouteRepo := repositories.NewMemoryRouteRepo(s.logger)
	for _, route := range routes {
		err := memoryRouteRepo.Save(ctx, &route)
		if err != nil {
			s.logger.Fatal("failed to save route", zap.Error(err))
			return
		}
	}
	authValidator := auth.NewAuthValidator(s.logger, s.connections.GetApiKeyRepo())
	authUseCase := usecases.NewAuthenticateRequestUseCase(authValidator, s.logger)
	routeUseCase := usecases.NewRouteRequestUseCase(cfg.Server.PathPrefix, proxyClientRepo, memoryRouteRepo, s.logger)
	gatewayHandler := handlers.NewGatewayHandler(s.logger, routeUseCase, authUseCase)
	// API routes
	api := s.echo.Group(cfg.Server.PathPrefix)
	health := api.Group("/health")

	// Health endpoints
	health.GET("/", healthHandler.Health)
	health.GET("/ready", healthHandler.Ready)
	health.GET("/live", healthHandler.Live)

	// Metrics endpoint
	api.GET("/metrics", healthHandler.Metrics)

	api.Any("/*", gatewayHandler.HandleRequest, security.RequestID(s.logger.With("component", "security")))

	s.logRegisteredRoutes()
}

func (s *Server) logRegisteredRoutes() {
	s.logger.Info("HTTP routes registered:")
	for _, route := range s.echo.Routes() {
		s.logger.Info("Route registered",
			"method", route.Method,
			"path", route.Path,
			"name", route.Name)
	}
}

func (s *Server) Start() error {
	address := fmt.Sprintf("%s:%s", s.config.Server.Host, s.config.Server.Port)
	s.logger.Info("Starting Product Service HTTP server", "address", address)

	return s.echo.Start(address)
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down Product Service HTTP server...")
	return s.echo.Shutdown(ctx)
}

func (s *Server) parseRoutes(cfg *config.Config) []entities.Route {
	var routes []entities.Route
	for _, backend := range cfg.Backends {
		entityBackend := entities.Backend{
			Id:         backend.ID,
			Host:       backend.Host,
			PathPrefix: backend.PathPrefix,
			Timeout:    30 * time.Second,
		}
		for _, route := range backend.Routes {
			routes = append(routes, entities.Route{
				ID:       route.ID,
				Method:   route.Method,
				Path:     route.Path,
				PathType: entities.PathType(route.PathType),
				Enabled:  route.Enabled,
				Backend:  &entityBackend,
				AuthPolicy: &entities.AuthPolicy{
					Enabled: route.AuthPolicy.Enabled,
					Type:    route.AuthPolicy.Type,
				}})
		}
	}
	return routes
}
