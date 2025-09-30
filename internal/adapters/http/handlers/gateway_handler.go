package handlers

import (
	"api-gateway/internal/application/dto"
	"api-gateway/internal/application/usecases"
	domainErrors "api-gateway/internal/domain/errors"
	"api-gateway/pkg/logger"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

type GatewayHandler struct {
	log          logger.Logger
	routeUseCase usecases.RouteRequestUseCases
	authUseCase  usecases.AuthenticationUseCases
}

func NewGatewayHandler(log logger.Logger, routeUseCase usecases.RouteRequestUseCases, authUseCase usecases.AuthenticationUseCases) *GatewayHandler {
	log.Info("Initializing gateway handler")

	return &GatewayHandler{
		log:          log,
		routeUseCase: routeUseCase,
		authUseCase:  authUseCase,
	}
}

func (h *GatewayHandler) HandleRequest(c echo.Context) error {
	startTime := time.Now()
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.log.Info("Incoming request received",
		"request_id", requestID,
		"method", c.Request().Method,
		"path", c.Request().URL.Path,
		"query", c.Request().URL.RawQuery,
		"remote_addr", c.RealIP(),
		"user_agent", c.Request().UserAgent(),
	)

	h.log.Debug("Request headers",
		"request_id", requestID,
		"headers", fmt.Sprintf("%v", c.Request().Header),
	)

	ctx := context.Background()

	h.log.Debug("Reading request body",
		"request_id", requestID,
	)

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Failed to read request body",
		})
	}

	h.log.Debug("Request body read",
		"request_id", requestID,
		"body_size", len(body),
		"body_payload", string(body),
	)

	gatewayRequestDto := dto.GatewayRequest{
		Path:        c.Request().URL.Path,
		Method:      c.Request().Method,
		Headers:     c.Request().Header,
		Body:        body,
		QueryParams: c.Request().URL.Query(),
	}

	h.log.Info("Gateway request created",
		"request_id", requestID,
		"path", gatewayRequestDto.Path,
		"method", gatewayRequestDto.Method,
		"has_query_params", len(gatewayRequestDto.QueryParams) > 0,
	)

	h.log.Debug("Looking up route",
		"request_id", requestID,
		"path", gatewayRequestDto.Path,
		"method", gatewayRequestDto.Method,
	)

	route, err := h.routeUseCase.GetRoute(ctx, &gatewayRequestDto)
	if err != nil {
		h.log.Warn("Route lookup failed",
			"request_id", requestID,
			"path", gatewayRequestDto.Path,
			"method", gatewayRequestDto.Method,
			"error", err,
			"duration_ms", time.Since(startTime).Milliseconds(),
		)

		if err.Error() == "route not found" {
			h.log.Info("Returning 404 Not Found",
				"request_id", requestID,
				"path", gatewayRequestDto.Path,
			)
			return c.JSON(http.StatusNotFound, domainErrors.NewValidationError("NOT_FOUND", err.Error()))
		}
		return err
	}

	h.log.Info("Route found",
		"request_id", requestID,
		"route_id", route.ID,
		"backend_host", route.Backend.Host,
		"backend_path_prefix", route.Backend.PathPrefix,
	)

	gatewayRequestDto.Host = route.Backend.Host + route.Backend.PathPrefix
	gatewayRequestDto.Path = route.Path

	if c.QueryString() != "" {
		gatewayRequestDto.Path += "?" + c.QueryString()
	}

	h.log.Debug("Gateway request updated with backend info",
		"request_id", requestID,
		"backend_host", gatewayRequestDto.Host,
		"backend_path", gatewayRequestDto.Path,
	)

	authRequest := dto.AuthRequest{
		Headers: c.Request().Header,
		Policy:  route.AuthPolicy,
	}

	h.log.Info("Starting authentication",
		"request_id", requestID,
		"auth_policy_type", route.AuthPolicy.Type,
		"auth_policy_enabled", route.AuthPolicy.Enabled,
	)

	authResponse, err := h.authUseCase.Execute(ctx, &authRequest)
	if err != nil {
		h.log.Warn("Authentication failed",
			"request_id", requestID,
			"error", fmt.Sprintf("%v", err),
			"duration_ms", time.Since(startTime).Milliseconds(),
		)

		h.log.Info("Returning 401 Unauthorized - Invalid token",
			"request_id", requestID,
		)

		return c.JSON(http.StatusUnauthorized, domainErrors.NewValidationError("API_TOKEN_NOT_VALID", err.Error()))
	}

	h.log.Info("Authentication completed",
		"request_id", requestID,
		"authenticated", authResponse.Authenticated,
		"user_id", authResponse.UserID,
	)

	h.log.Debug("Auth response details",
		"request_id", requestID,
		"auth_response", fmt.Sprintf("%v", authResponse),
	)

	if authResponse.Authenticated {
		h.log.Info("User authenticated, executing route",
			"request_id", requestID,
			"backend_url", gatewayRequestDto.Host+gatewayRequestDto.Path,
		)

		gatewayResponse, err := h.routeUseCase.Execute(ctx, &gatewayRequestDto)
		if err != nil {
			h.log.Error("Route execution failed",
				"request_id", requestID,
				"error", fmt.Sprintf("%v", err),
				"backend_host", gatewayRequestDto.Host,
				"backend_path", gatewayRequestDto.Path,
				"duration_ms", time.Since(startTime).Milliseconds(),
			)
			return err
		}

		h.log.Info("Route executed successfully",
			"request_id", requestID,
			"status_code", gatewayResponse.StatusCode,
			"response_size", len(gatewayResponse.Body),
		)

		h.log.Debug("Gateway response details",
			"request_id", requestID,
			"gateway_response", fmt.Sprintf("%v", gatewayResponse),
		)

		h.log.Debug("Copying response headers",
			"request_id", requestID,
			"headers_count", len(gatewayResponse.Headers),
		)

		for key, values := range gatewayResponse.Headers {
			for _, value := range values {
				c.Response().Header().Add(key, value)
			}
		}

		h.log.Info("Sending response to client",
			"request_id", requestID,
			"status_code", gatewayResponse.StatusCode,
			"content_type", c.Response().Header().Get("Content-Type"),
			"response_size", len(gatewayResponse.Body),
			"total_duration_ms", time.Since(startTime).Milliseconds(),
		)

		return c.Blob(
			gatewayResponse.StatusCode,
			c.Response().Header().Get("Content-Type"),
			gatewayResponse.Body,
		)
	}

	h.log.Warn("User not authenticated",
		"request_id", requestID,
		"authenticated", authResponse.Authenticated,
		"duration_ms", time.Since(startTime).Milliseconds(),
	)

	h.log.Info("Returning 401 Unauthorized - No authenticated user",
		"request_id", requestID,
	)

	return c.JSON(http.StatusUnauthorized, domainErrors.NewValidationError("NOT_UNAUTHENTICATED", "No authenticated user"))
}
