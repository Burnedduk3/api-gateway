package usecases

import (
	"api-gateway/internal/application/dto"
	"api-gateway/internal/application/ports"
	"api-gateway/internal/domain/entities"
	"api-gateway/pkg/logger"
	"context"
	"strings"
	"time"
)

// RouteRequestUseCases defines the interface for route operations
type RouteRequestUseCases interface {
	Execute(ctx context.Context, req *dto.GatewayRequest) (*dto.GatewayResponse, error)
	GetRoute(ctx context.Context, req *dto.GatewayRequest) (*entities.Route, error)
}

// RouteRequestUseCase implements RouteUseCases interface
type routeRequestUseCaseImpl struct {
	serverPathPrefix string
	logger           logger.Logger
	routeRepo        ports.RouteRepository
	proxyClient      ports.ProxyClient
}

// NewRouteRequestUseCase creates a new instance of route request use case
func NewRouteRequestUseCase(serverPathPrefix string, proxyClient ports.ProxyClient, routeRepo ports.RouteRepository, log logger.Logger) RouteRequestUseCases {
	log.Info("Initializing route request use case",
		"server_path_prefix", serverPathPrefix,
	)

	return &routeRequestUseCaseImpl{
		serverPathPrefix: serverPathPrefix,
		routeRepo:        routeRepo,
		proxyClient:      proxyClient,
		logger:           log.With("component", "routeRequest_usecases"),
	}
}

func (r routeRequestUseCaseImpl) GetRoute(ctx context.Context, req *dto.GatewayRequest) (*entities.Route, error) {
	startTime := time.Now()

	r.logger.Debug("Starting route lookup",
		"path", req.Path,
		"method", req.Method,
		"server_path_prefix", r.serverPathPrefix,
	)

	cleanPath := strings.TrimPrefix(req.Path, r.serverPathPrefix)

	r.logger.Debug("Path cleaned for route lookup",
		"original_path", req.Path,
		"clean_path", cleanPath,
		"method", req.Method,
	)

	route, err := r.routeRepo.FindByPathAndMethod(ctx, cleanPath, req.Method)
	duration := time.Since(startTime)

	if err != nil {
		r.logger.Warn("Route lookup failed",
			"clean_path", cleanPath,
			"method", req.Method,
			"error", err,
			"duration_ms", duration.Milliseconds(),
		)
		return nil, err
	}

	r.logger.Info("Route found successfully",
		"route_id", route.ID,
		"route_path", route.Path,
		"method", route.Method,
		"backend_host", route.Backend.Host,
		"backend_path_prefix", route.Backend.PathPrefix,
		"enabled", route.Enabled,
		"duration_ms", duration.Milliseconds(),
	)

	actualPath := strings.TrimPrefix(cleanPath, "/"+route.Backend.Id)
	route.Path = actualPath
	return route, err
}

func (r routeRequestUseCaseImpl) Execute(ctx context.Context, req *dto.GatewayRequest) (*dto.GatewayResponse, error) {
	startTime := time.Now()

	r.logger.Info("Starting route execution",
		"method", req.Method,
		"host", req.Host,
		"path", req.Path,
		"body_size", len(req.Body),
	)

	r.logger.Debug("Building proxy request",
		"target_url", req.Host+req.Path,
		"method", req.Method,
		"headers_count", len(req.Headers),
		"has_body", len(req.Body) > 0,
	)

	proxyRequest := dto.ProxyRequest{
		Method:  req.Method,
		Headers: req.Headers,
		Body:    req.Body,
		URL:     req.Host + req.Path,
	}

	r.logger.Info("Forwarding request to backend via proxy",
		"url", proxyRequest.URL,
		"method", proxyRequest.Method,
	)

	proxyStart := time.Now()
	res, err := r.proxyClient.Forward(ctx, &proxyRequest)
	proxyDuration := time.Since(proxyStart)

	if err != nil {
		r.logger.Error("Proxy forward failed",
			"error", err.Error(),
			"url", proxyRequest.URL,
			"method", proxyRequest.Method,
			"proxy_duration_ms", proxyDuration.Milliseconds(),
			"total_duration_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, err
	}

	r.logger.Info("Proxy response received",
		"status_code", res.StatusCode,
		"response_size", len(res.Body),
		"proxy_duration_ms", proxyDuration.Milliseconds(),
	)

	r.logger.Debug("Building gateway response",
		"status_code", res.StatusCode,
		"headers_count", len(res.Headers),
		"body_size", len(res.Body),
	)

	gatewayResponse := dto.GatewayResponse{
		StatusCode: res.StatusCode,
		Headers:    res.Headers,
		Body:       res.Body,
	}

	totalDuration := time.Since(startTime)

	r.logger.Info("Route execution completed",
		"status_code", gatewayResponse.StatusCode,
		"request_size", len(req.Body),
		"response_size", len(gatewayResponse.Body),
		"proxy_duration_ms", proxyDuration.Milliseconds(),
		"total_duration_ms", totalDuration.Milliseconds(),
	)

	return &gatewayResponse, nil
}
