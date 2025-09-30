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
	// Log the forwarding attempt
	p.log.Debug("Forwarding request to backend",
		"method", req.Method,
		"url", req.URL,
		"body_size", len(req.Body),
	)

	// 1. Create HTTP request with context
	httpReq, err := http.NewRequestWithContext(
		ctx,
		req.Method,
		req.URL,
		bytes.NewReader(req.Body),
	)
	if err != nil {
		p.log.Error("Failed to create HTTP request",
			"error", err,
			"url", req.URL,
		)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// 2. Copy headers from proxy request to HTTP request
	for key, values := range req.Headers {
		for _, value := range values {
			httpReq.Header.Add(key, value)
		}
	}

	// 3. Set Content-Length if body exists
	if len(req.Body) > 0 {
		httpReq.ContentLength = int64(len(req.Body))
	}

	// 4. Make the HTTP call to backend
	startTime := time.Now()
	resp, err := p.httpClient.Do(httpReq)
	duration := time.Since(startTime)

	if err != nil {
		p.log.Error("Failed to forward request to backend",
			"error", err,
			"url", req.URL,
			"method", req.Method,
			"duration_ms", duration.Milliseconds(),
		)
		return nil, fmt.Errorf("failed to forward request: %w", err)
	}
	defer resp.Body.Close()

	// 5. Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		p.log.Error("Failed to read response body",
			"error", err,
			"url", req.URL,
			"status_code", resp.StatusCode,
		)
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Log successful forwarding
	p.log.Info("Request forwarded successfully",
		"method", req.Method,
		"url", req.URL,
		"status_code", resp.StatusCode,
		"response_size", len(body),
		"duration_ms", duration.Milliseconds(),
	)

	// 6. Build and return proxy response
	proxyResp := &dto.ProxyResponse{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}

	return proxyResp, nil
}
