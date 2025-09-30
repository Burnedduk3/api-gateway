package handlers

import (
	"api-gateway/internal/application/dto"
	"api-gateway/pkg/logger"
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

type ProxyClient struct {
	httpClient *http.Client
	log        logger.Logger
}

func NewProxyClient(timeout time.Duration, log logger.Logger) *ProxyClient {
	log.Info("Initializing proxy client",
		"timeout_seconds", timeout.Seconds(),
	)

	return &ProxyClient{
		httpClient: &http.Client{
			Timeout: timeout,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		log: log,
	}
}

func (p *ProxyClient) Forward(ctx context.Context, req *dto.ProxyRequest) (*dto.ProxyResponse, error) {
	startTime := time.Now()

	// Extract request ID from context if available
	requestID := "unknown"
	if id := ctx.Value("request_id"); id != nil {
		if strID, ok := id.(string); ok {
			requestID = strID
		}
	}

	// Log the forwarding attempt
	p.log.Info("Starting request forward to backend",
		"request_id", requestID,
		"method", req.Method,
		"url", req.URL,
		"body_size", len(req.Body),
		"has_body", len(req.Body) > 0,
	)

	p.log.Debug("Request details",
		"request_id", requestID,
		"headers_count", len(req.Headers),
	)

	// 1. Create HTTP request with context
	p.log.Debug("Creating HTTP request with context",
		"request_id", requestID,
		"method", req.Method,
		"url", req.URL,
	)

	httpReq, err := http.NewRequestWithContext(
		ctx,
		req.Method,
		req.URL,
		bytes.NewReader(req.Body),
	)
	if err != nil {
		p.log.Error("Failed to create HTTP request",
			"request_id", requestID,
			"error", err,
			"url", req.URL,
			"duration_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	p.log.Debug("HTTP request created successfully",
		"request_id", requestID,
	)

	// 2. Copy headers from proxy request to HTTP request
	p.log.Debug("Copying headers to HTTP request",
		"request_id", requestID,
		"headers_count", len(req.Headers),
	)

	headersCopied := 0
	for key, values := range req.Headers {
		for _, value := range values {
			httpReq.Header.Add(key, value)
			headersCopied++
		}
	}

	p.log.Debug("Headers copied successfully",
		"request_id", requestID,
		"headers_copied", headersCopied,
	)

	// 3. Set Content-Length if body exists
	if len(req.Body) > 0 {
		httpReq.ContentLength = int64(len(req.Body))
		p.log.Debug("Content-Length header set",
			"request_id", requestID,
			"content_length", httpReq.ContentLength,
		)
	}

	// 4. Make the HTTP call to backend
	p.log.Info("Sending HTTP request to backend",
		"request_id", requestID,
		"method", req.Method,
		"url", req.URL,
	)

	callStart := time.Now()
	resp, err := p.httpClient.Do(httpReq)
	duration := time.Since(callStart)

	if err != nil {
		p.log.Error("Failed to forward request to backend",
			"request_id", requestID,
			"error", err,
			"url", req.URL,
			"method", req.Method,
			"call_duration_ms", duration.Milliseconds(),
			"total_duration_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, fmt.Errorf("failed to forward request: %w", err)
	}
	defer resp.Body.Close()

	p.log.Info("Received response from backend",
		"request_id", requestID,
		"status_code", resp.StatusCode,
		"status", resp.Status,
		"call_duration_ms", duration.Milliseconds(),
	)

	p.log.Debug("Response headers received",
		"request_id", requestID,
		"headers_count", len(resp.Header),
		"content_type", resp.Header.Get("Content-Type"),
		"content_length", resp.Header.Get("Content-Length"),
	)

	// Warn on error status codes
	if resp.StatusCode >= 400 && resp.StatusCode < 500 {
		p.log.Warn("Backend returned client error",
			"request_id", requestID,
			"status_code", resp.StatusCode,
			"url", req.URL,
		)
	} else if resp.StatusCode >= 500 {
		p.log.Warn("Backend returned server error",
			"request_id", requestID,
			"status_code", resp.StatusCode,
			"url", req.URL,
		)
	}

	// Warn on slow responses
	if duration > 2*time.Second {
		p.log.Warn("Slow backend response detected",
			"request_id", requestID,
			"call_duration_ms", duration.Milliseconds(),
			"url", req.URL,
		)
	}

	// 5. Read response body
	p.log.Debug("Reading response body",
		"request_id", requestID,
		"status_code", resp.StatusCode,
	)

	readStart := time.Now()
	body, err := io.ReadAll(resp.Body)
	readDuration := time.Since(readStart)

	if err != nil {
		p.log.Error("Failed to read response body",
			"request_id", requestID,
			"error", err,
			"url", req.URL,
			"status_code", resp.StatusCode,
			"read_duration_ms", readDuration.Milliseconds(),
			"total_duration_ms", time.Since(startTime).Milliseconds(),
		)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	p.log.Debug("Response body read successfully",
		"request_id", requestID,
		"body_size", len(body),
		"read_duration_ms", readDuration.Milliseconds(),
	)

	// Warn on large response bodies
	if len(body) > 10*1024*1024 {
		p.log.Warn("Large response body detected",
			"request_id", requestID,
			"body_size_mb", len(body)/(1024*1024),
			"url", req.URL,
		)
	}

	totalDuration := time.Since(startTime)

	// Log successful forwarding
	p.log.Info("Request forwarded successfully",
		"request_id", requestID,
		"method", req.Method,
		"url", req.URL,
		"status_code", resp.StatusCode,
		"request_size", len(req.Body),
		"response_size", len(body),
		"backend_call_duration_ms", duration.Milliseconds(),
		"body_read_duration_ms", readDuration.Milliseconds(),
		"total_duration_ms", totalDuration.Milliseconds(),
	)

	// 6. Build and return proxy response
	p.log.Debug("Building proxy response",
		"request_id", requestID,
		"status_code", resp.StatusCode,
		"body_size", len(body),
	)

	proxyResp := &dto.ProxyResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}

	return proxyResp, nil
}
