package usecases

import (
	"api-gateway/internal/application/dto"
	"api-gateway/internal/application/ports"
	"api-gateway/internal/domain/entities"
	"api-gateway/pkg/logger"
	"context"
	"strings"
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
	return &routeRequestUseCaseImpl{
		serverPathPrefix: serverPathPrefix,
		routeRepo:        routeRepo,
		proxyClient:      proxyClient,
		logger:           log.With("component", "routeRequest_usecases"),
	}
}

func (r routeRequestUseCaseImpl) GetRoute(ctx context.Context, req *dto.GatewayRequest) (*entities.Route, error) {
	cleanPath := strings.TrimPrefix(req.Path, r.serverPathPrefix)
	route, err := r.routeRepo.FindByPathAndMethod(ctx, cleanPath, req.Method)
	if err != nil {
		return nil, err
	}
	return route, err
}

func (r routeRequestUseCaseImpl) Execute(ctx context.Context, req *dto.GatewayRequest) (*dto.GatewayResponse, error) {
	proxyRequest := dto.ProxyRequest{
		Method:  req.Method,
		Headers: req.Headers,
		Body:    req.Body,
		URL:     req.Host + req.Path,
	}
	res, err := r.proxyClient.Forward(ctx, &proxyRequest)
	if err != nil {
		r.logger.Error(err.Error())
		return nil, err
	}
	gatewayResponse := dto.GatewayResponse{
		StatusCode: res.StatusCode,
		Headers:    res.Headers,
		Body:       res.Body,
	}
	return &gatewayResponse, nil
}
