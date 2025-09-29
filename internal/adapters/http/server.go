package http

import (
	"api-gateway/internal/adapters/http/handlers"
	"api-gateway/internal/adapters/http/middlewares/logging"
	"api-gateway/internal/config"
	"api-gateway/internal/infrastructure"
	"api-gateway/pkg/logger"
	"context"
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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
	server.setupRoutes()

	return server, nil
}

func (s *Server) setupMiddleware() {
	// Request ID middleware
	s.echo.Use(middleware.RequestID())

	// Replace Echo's logger with our custom Zap logger
	s.echo.Use(logging.ZapLogger(s.logger.With("component", "http")))

	// Recovery middleware
	s.echo.Use(middleware.Recover())

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

func (s *Server) setupRoutes() {
	// Health check handlers with database connections
	healthHandler := handlers.NewHealthHandler(s.logger, s.connections)

	// Product repository and use cases setup

	// API v1 routes
	v1 := s.echo.Group("/api/v1")

	// Health endpoints
	v1.GET("/health", healthHandler.Health)
	v1.GET("/health/ready", healthHandler.Ready)
	v1.GET("/health/live", healthHandler.Live)

	// Metrics endpoint
	v1.GET("/metrics", healthHandler.Metrics)

	//// Product endpoints
	//products := v1.Group("/products")
	//{
	//	// Core CRUD operations
	//	products.POST("", productHandler.CreateProduct)    // Create product
	//	products.GET("", productHandler.ListProducts)      // List products with pagination
	//	products.GET("/:id", productHandler.GetProduct)    // Get product by ID
	//	products.PUT("/:id", productHandler.UpdateProduct) // Update product
	//
	//	// SKU-based operations
	//	products.GET("/sku/:sku", productHandler.GetProductBySKU) // Get product by SKU
	//
	//	// Stock management
	//	products.PATCH("/:id/stock", productHandler.UpdateProductStock) // Update stock only
	//
	//	// Price management
	//	products.PATCH("/:id/price", productHandler.UpdateProductPrice) // Update price only
	//
	//	// Status management
	//	products.PATCH("/:id/activate", productHandler.ActivateProduct)       // Activate product
	//	products.PATCH("/:id/deactivate", productHandler.DeactivateProduct)   // Deactivate product
	//	products.PATCH("/:id/discontinue", productHandler.DiscontinueProduct) // Discontinue product
	//}

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
