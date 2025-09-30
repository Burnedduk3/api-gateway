package entities_test

import (
	"testing"
	"time"

	"api-gateway/internal/domain/entities"

	"github.com/stretchr/testify/assert"
)

func TestBackend_GetURL(t *testing.T) {
	tests := []struct {
		name        string
		backend     *entities.Backend
		requestPath string
		expectedURL string
	}{
		{
			name: "with path prefix addition",
			backend: &entities.Backend{
				Host:       "http://service-c:8080",
				PathPrefix: "/v2",
			},
			requestPath: "/users",
			expectedURL: "http://service-c:8080/v2/users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.backend.GetURL(tt.requestPath)
			assert.Equal(t, tt.expectedURL, result)
		})
	}
}

func TestBackend_IsHealthy(t *testing.T) {
	tests := []struct {
		name     string
		backend  *entities.Backend
		expected bool
	}{
		{
			name: "healthy backend",
			backend: &entities.Backend{
				Host:            "http://service:8080",
				Healthy:         true,
				LastHealthCheck: time.Now(),
			},
			expected: true,
		},
		{
			name: "unhealthy backend",
			backend: &entities.Backend{
				Host:            "http://service:8080",
				Healthy:         false,
				LastHealthCheck: time.Now(),
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.backend.IsHealthy())
		})
	}
}

func TestBackend_Validate(t *testing.T) {
	tests := []struct {
		name    string
		backend *entities.Backend
		wantErr bool
	}{
		{
			name: "valid backend",
			backend: &entities.Backend{
				Host:    "http://service:8080",
				Timeout: 30 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "missing host",
			backend: &entities.Backend{
				Timeout: 30 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid host format",
			backend: &entities.Backend{
				Host: "not-a-valid-url",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.backend.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBackend_UpdateHealth(t *testing.T) {
	backend := &entities.Backend{
		Host:    "http://service:8080",
		Healthy: true,
	}

	backend.UpdateHealth(false)

	assert.False(t, backend.Healthy)
	assert.False(t, backend.LastHealthCheck.IsZero())
}
