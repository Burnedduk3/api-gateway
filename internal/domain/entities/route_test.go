package entities_test

import (
	"testing"

	"api-gateway/internal/domain/entities"

	"github.com/stretchr/testify/assert"
)

func TestRoute_Match(t *testing.T) {
	tests := []struct {
		name           string
		route          *entities.Route
		incomingPath   string
		incomingMethod string
		shouldMatch    bool
	}{
		{
			name: "exact path match",
			route: &entities.Route{
				Path:   "user/api/users",
				Method: "GET",
				Backend: &entities.Backend{
					Host: "http://service:8080",
					Id:   "user",
				},
			},
			incomingPath:   "api/users",
			incomingMethod: "GET",
			shouldMatch:    true,
		},
		{
			name: "method mismatch",
			route: &entities.Route{
				Path:   "user/api/users",
				Method: "POST",
				Backend: &entities.Backend{
					Host: "http://service:8080",
					Id:   "user",
				},
			},
			incomingPath:   "user/api/users",
			incomingMethod: "GET",
			shouldMatch:    false,
		},
		{
			name: "wildcard method match",
			route: &entities.Route{
				Path:   "user/api/health",
				Method: "*",
				Backend: &entities.Backend{
					Host: "http://service:8080",
					Id:   "user",
				},
			},
			incomingPath:   "user/api/health",
			incomingMethod: "GET",
			shouldMatch:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.route.Match(tt.incomingPath, tt.incomingMethod)
			assert.Equal(t, tt.shouldMatch, result)
		})
	}
}

func TestRoute_IsEnabled(t *testing.T) {
	tests := []struct {
		name     string
		route    *entities.Route
		expected bool
	}{
		{
			name: "enabled route",
			route: &entities.Route{
				Enabled: true,
			},
			expected: true,
		},
		{
			name: "disabled route",
			route: &entities.Route{
				Enabled: false,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.route.IsEnabled())
		})
	}
}

func TestRoute_GetBackend(t *testing.T) {
	backend := &entities.Backend{
		Host: "http://service-a:8080",
	}

	route := &entities.Route{
		Backend: backend,
	}

	assert.Equal(t, backend, route.GetBackend())
}

func TestRoute_Validate(t *testing.T) {
	tests := []struct {
		name    string
		route   *entities.Route
		wantErr bool
	}{
		{
			name: "valid route",
			route: &entities.Route{
				ID:      "route-1",
				Path:    "/api/users",
				Method:  "GET",
				Backend: &entities.Backend{Host: "http://service:8080"},
				Enabled: true,
			},
			wantErr: false,
		},
		{
			name: "missing path",
			route: &entities.Route{
				ID:      "route-1",
				Method:  "GET",
				Backend: &entities.Backend{Host: "http://service:8080"},
			},
			wantErr: true,
		},
		{
			name: "missing method",
			route: &entities.Route{
				ID:      "route-1",
				Path:    "/api/users",
				Backend: &entities.Backend{Host: "http://service:8080"},
			},
			wantErr: true,
		},
		{
			name: "missing backend",
			route: &entities.Route{
				ID:     "route-1",
				Path:   "/api/users",
				Method: "GET",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.route.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
