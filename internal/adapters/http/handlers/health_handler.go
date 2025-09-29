package handlers

import (
	"api-gateway/internal/infrastructure"
	"context"
	"net/http"
	"runtime"
	"time"

	"api-gateway/pkg/logger"

	"github.com/labstack/echo/v4"
)

type HealthHandler struct {
	logger      logger.Logger
	startTime   time.Time
	connections *infrastructure.DatabaseConnections
}

func NewHealthHandler(logger logger.Logger, connections *infrastructure.DatabaseConnections) *HealthHandler {
	return &HealthHandler{
		logger:      logger.With("component", "health_handler"),
		startTime:   time.Now(),
		connections: connections,
	}
}

type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Uptime    string                 `json:"uptime"`
	Checks    map[string]interface{} `json:"checks,omitempty"`
}

type MetricsResponse struct {
	Service   string    `json:"service"`
	Version   string    `json:"version"`
	Timestamp time.Time `json:"timestamp"`
	Uptime    string    `json:"uptime"`
	Runtime   struct {
		GoVersion   string `json:"go_version"`
		Goroutines  int    `json:"goroutines"`
		MemoryAlloc uint64 `json:"memory_alloc"`
		MemoryTotal uint64 `json:"memory_total"`
		MemorySys   uint64 `json:"memory_sys"`
		GCCount     uint32 `json:"gc_count"`
	} `json:"runtime"`
}

// Health returns basic service health status
func (h *HealthHandler) Health(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Debug("Health check requested",
		"request_id", requestID,
		"remote_ip", c.RealIP(),
		"user_agent", c.Request().UserAgent())

	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now(),
		Service:   "api-gateway",
		Version:   "1.0.0",
		Uptime:    time.Since(h.startTime).String(),
	}

	return c.JSON(http.StatusOK, response)
}

// Ready checks if the service is ready to accept requests
// This is where you'd add database connectivity checks, etc.
func (h *HealthHandler) Ready(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Info("Readiness check requested",
		"request_id", requestID,
		"remote_ip", c.RealIP())

	// Create context with timeout for health checks
	ctx, cancel := context.WithTimeout(c.Request().Context(), 5*time.Second)
	defer cancel()

	// Perform actual health checks
	checks := h.connections.HealthCheck(ctx)

	// Convert to response format and check if all are healthy
	responseChecks := make(map[string]interface{})

	allHealthy := true
	for component, err := range checks {
		if err != nil {
			allHealthy = false
			responseChecks[component] = map[string]interface{}{
				"status":  "unhealthy",
				"message": err.Error(),
			}
			h.logger.Warn("Component unhealthy during readiness check",
				"component", component,
				"error", err.Error(),
				"request_id", requestID)
		} else {
			responseChecks[component] = map[string]interface{}{
				"status":  "healthy",
				"message": "Connection successful",
			}
		}
	}

	status := "ready"
	httpStatus := http.StatusOK
	if !allHealthy {
		status = "not_ready"
		httpStatus = http.StatusServiceUnavailable
	}

	response := HealthResponse{
		Status:    status,
		Timestamp: time.Now(),
		Service:   "api-gateway",
		Version:   "1.0.0",
		Uptime:    time.Since(h.startTime).String(),
		Checks:    responseChecks,
	}

	h.logger.Info("Readiness check completed",
		"status", status,
		"checks_count", len(responseChecks),
		"healthy", allHealthy,
		"request_id", requestID)

	return c.JSON(httpStatus, response)
}

// Live checks if the service is alive (minimal check)
func (h *HealthHandler) Live(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Debug("Liveness check requested",
		"request_id", requestID,
		"remote_ip", c.RealIP())

	response := HealthResponse{
		Status:    "alive",
		Timestamp: time.Now(),
		Service:   "api-gateway",
		Version:   "1.0.0",
		Uptime:    time.Since(h.startTime).String(),
	}

	return c.JSON(http.StatusOK, response)
}

// Metrics returns service metrics and runtime information
func (h *HealthHandler) Metrics(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Debug("Metrics requested",
		"request_id", requestID,
		"remote_ip", c.RealIP())

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	response := MetricsResponse{
		Service:   "api-gateway",
		Version:   "1.0.0",
		Timestamp: time.Now(),
		Uptime:    time.Since(h.startTime).String(),
	}

	response.Runtime.GoVersion = runtime.Version()
	response.Runtime.Goroutines = runtime.NumGoroutine()
	response.Runtime.MemoryAlloc = m.Alloc
	response.Runtime.MemoryTotal = m.TotalAlloc
	response.Runtime.MemorySys = m.Sys
	response.Runtime.GCCount = m.NumGC

	h.logger.Info("Metrics collected",
		"goroutines", response.Runtime.Goroutines,
		"memory_alloc_mb", response.Runtime.MemoryAlloc/1024/1024,
		"gc_count", response.Runtime.GCCount,
		"request_id", requestID)

	return c.JSON(http.StatusOK, response)
}
