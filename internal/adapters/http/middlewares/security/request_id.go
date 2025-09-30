package security

import (
	"api-gateway/pkg/logger"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

func RequestID(logger logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Request().Header.Get("X-Request-ID")
			logger.Info(fmt.Sprintf("Value of header X-Request-ID: %s", requestID))
			if requestID == "" {
				logger.Info("RequestID is empty")
				return c.String(http.StatusBadRequest, "X-Request-ID is required")
			}
			logger.Info(fmt.Sprintf("setting header X-Request-ID: %s", requestID))
			c.Request().Header.Set("X-Request-ID", requestID)

			c.Response().Header().Set("X-Request-ID", requestID)

			c.Set("request_id", requestID)
			return next(c)
		}
	}
}
