package handlers

import (
	"api-gateway/internal/application/dto"
	"api-gateway/internal/application/usecases"
	domainErrors "api-gateway/internal/domain/errors"
	"api-gateway/pkg/logger"
	"context"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type GatewayHandler struct {
	log          logger.Logger
	routeUseCase usecases.RouteRequestUseCases
	authUseCase  usecases.AuthenticationUseCases
}

func NewGatewayHandler(log logger.Logger, routeUseCase usecases.RouteRequestUseCases, authUseCase usecases.AuthenticationUseCases) *GatewayHandler {
	return &GatewayHandler{
		log:          log,
		routeUseCase: routeUseCase,
		authUseCase:  authUseCase,
	}
}

func (h *GatewayHandler) HandleRequest(c echo.Context) error {
	h.log.Info(fmt.Sprintf("Incoming Headers: %v", c.Request().Header))
	ctx := context.Background()
	var body []byte
	c.Request().Body.Read(body)
	gatewayRequestDto := dto.GatewayRequest{
		Path:        c.Request().URL.Path,
		Method:      c.Request().Method,
		Headers:     c.Request().Header,
		Body:        body,
		QueryParams: c.Request().URL.Query(),
	}
	route, err := h.routeUseCase.GetRoute(ctx, &gatewayRequestDto)
	if err != nil {
		if err.Error() == "route not found" {
			return c.JSON(http.StatusNotFound, domainErrors.NewValidationError("NOT_FOUND", err.Error()))
		}
		return err
	}
	gatewayRequestDto.Host = route.Backend.Host
	gatewayRequestDto.Path = route.Backend.PathPrefix + route.Path
	authRequest := dto.AuthRequest{
		Headers: c.Request().Header,
		Policy:  route.AuthPolicy,
	}
	authResponse, err := h.authUseCase.Execute(ctx, &authRequest)
	if err != nil {
		h.log.Error(fmt.Sprintf("Error executing auth: %v", err))
		return err
	}
	h.log.Info(fmt.Sprintf("Auth Response: %v", authResponse))
	if authResponse.Authenticated {
		gatewayResponse, err := h.routeUseCase.Execute(ctx, &gatewayRequestDto)
		if err != nil {
			h.log.Error(fmt.Sprintf("Error executing route: %v", err))
			return err
		}
		h.log.Debug(fmt.Sprintf("Gateway Response: %v", gatewayResponse))

		for key, values := range gatewayResponse.Headers {
			for _, value := range values {
				c.Response().Header().Add(key, value)
			}
		}
		
		return c.Blob(
			gatewayResponse.StatusCode,
			c.Response().Header().Get("Content-Type"),
			gatewayResponse.Body,
		)
	}
	return c.JSON(http.StatusUnauthorized, domainErrors.NewValidationError("NOT_UNAUTHENTICATED", "No authenticated user"))
}
