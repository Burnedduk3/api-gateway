package ports

import (
	"api-gateway/internal/application/dto"
	"context"
)

type ProxyClient interface {
	Forward(ctx context.Context, req *dto.ProxyRequest) (*dto.ProxyResponse, error)
}
