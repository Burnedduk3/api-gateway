package entities

import (
	domainErrors "api-gateway/internal/domain/errors"
	"net/url"
	"strings"
	"time"
)

type Backend struct {
	Id              string
	Host            string
	StripPrefix     string
	PathPrefix      string
	LastHealthCheck time.Time
	Timeout         time.Duration
	Healthy         bool
}

func (b *Backend) GetURL(requestPath string) string {
	backendUrl := b.Host + b.PathPrefix
	backendUrl += strings.TrimPrefix(requestPath, b.StripPrefix)
	return backendUrl
}

func (b *Backend) IsHealthy() bool {
	return b.Healthy
}

func (b *Backend) Validate() error {
	if b.Host == "" {
		return domainErrors.ErrBackendMissingHost
	}

	_, err := url.ParseRequestURI(b.Host)

	if err != nil {
		return domainErrors.ErrBackendInvalidHost
	}
	return nil
}

func (b *Backend) UpdateHealth(healthy bool) {
	b.Healthy = healthy
	b.LastHealthCheck = time.Now()
}
