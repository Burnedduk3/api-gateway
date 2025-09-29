package ports

import (
	"context"
)

type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}
