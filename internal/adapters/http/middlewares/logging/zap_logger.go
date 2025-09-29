package logging

import (
	"api-gateway/pkg/logger"
	"time"

	"github.com/labstack/echo/v4"
)

func ZapLogger(logger logger.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			// Process request
			err := next(c)
			if err != nil {
				c.Error(err)
			}

			// Calculate latency
			latency := time.Since(start)

			// Get request details
			req := c.Request()
			res := c.Response()

			// Determine log level based on status code
			status := res.Status

			// Create structured log fields
			fields := []interface{}{
				"time", start.Format(time.RFC3339),
				"id", res.Header().Get(echo.HeaderXRequestID),
				"remote_ip", c.RealIP(),
				"host", req.Host,
				"method", req.Method,
				"uri", req.RequestURI,
				"user_agent", req.UserAgent(),
				"status", status,
				"latency", latency.Nanoseconds(),
				"latency_human", latency.String(),
				"bytes_in", req.ContentLength,
				"bytes_out", res.Size,
			}

			// Add error if present
			if err != nil {
				fields = append(fields, "error", err.Error())
			}

			// Log based on status code
			switch {
			case status >= 500:
				logger.Error("HTTP request completed", fields...)
			case status >= 400:
				logger.Warn("HTTP request completed", fields...)
			case status >= 300:
				logger.Info("HTTP request completed", fields...)
			default:
				logger.Debug("HTTP request completed", fields...)
			}

			return nil
		}
	}
}
