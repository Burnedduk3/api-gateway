package handlers

import (
	"api-gateway/internal/application/dto"
	"api-gateway/pkg/logger"
	"context"
	"net/http"
	"time"
)

type ProxyClient struct {
	log        logger.Logger
	httpClient *http.Client
}

func NewProxyClient(log logger.Logger, timeout time.Duration) *ProxyClient {
	return &ProxyClient{
		log: log,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (p *ProxyClient) Forward(ctx context.Context, req *dto.ProxyRequest) (*dto.ProxyResponse, error) {
	// Create HTTP request
	// Forward to backend
	// Read response
	// Return proxy response
	return nil, nil
}
