package entities

import (
	domainErrors "api-gateway/internal/domain/errors"
	"strings"
)

type PathType string

const (
	PathTypeExact  PathType = "exact"
	PathTypePrefix PathType = "prefix"
	PathTypeRegEx  PathType = "regex"
)

type Route struct {
	ID       string   `json:"id"`
	Method   string   `json:"method"`
	Path     string   `json:"path"`
	PathType PathType `json:"pathType,omitempty"`
	Enabled  bool

	Backend    *Backend
	AuthPolicy *AuthPolicy `json:"authPolicy,omitempty"`
}

func NewRoute(method, path, pathType string, enabled bool, backend *Backend, authPolicy *AuthPolicy) *Route {
	return &Route{
		Method:     method,
		Path:       path,
		PathType:   PathType(pathType),
		Enabled:    enabled,
		Backend:    backend,
		AuthPolicy: authPolicy,
	}
}

// Match checks if the route matches the incoming request
// backendID is prepended to the route path before matching
func (r *Route) Match(incomingPath, incomingMethod, backendID string) bool {
	// Check method match first (including wildcard)
	if r.Method != "*" && r.Method != incomingMethod {
		return false
	}

	// Prepend backend ID to route path
	fullPath := "/" + backendID + r.Path

	switch r.PathType {
	case "exact":
		return fullPath == incomingPath

	case "prefix":
		// First try parameterized matching if path contains ":"
		if strings.Contains(fullPath, ":") {
			return r.matchParameterizedPath(incomingPath, fullPath)
		}
		// Otherwise do simple prefix matching
		return strings.HasPrefix(incomingPath, fullPath)

	default:
		// Default to exact match
		return fullPath == incomingPath
	}
}

// matchParameterizedPath handles paths with parameters like /users/:id
func (r *Route) matchParameterizedPath(incomingPath, fullPath string) bool {
	routeParts := strings.Split(fullPath, "/")
	pathParts := strings.Split(incomingPath, "/")

	// Must have same number of segments
	if len(routeParts) != len(pathParts) {
		return false
	}

	for i := 0; i < len(routeParts); i++ {
		routePart := routeParts[i]
		pathPart := pathParts[i]

		// If it's a parameter (starts with :), it matches anything
		if strings.HasPrefix(routePart, ":") {
			continue
		}

		// Otherwise must match exactly
		if routePart != pathPart {
			return false
		}
	}

	return true
}

func (r *Route) IsEnabled() bool {
	return r.Enabled
}

func (r *Route) GetBackend() *Backend {
	return r.Backend
}

func (r *Route) Validate() error {
	if r.Path == "" {
		return domainErrors.ErrRouteMissingPath
	}

	if r.Method == "" {
		return domainErrors.ErrRouteMissingMethod
	}

	if r.Backend == nil {
		return domainErrors.ErrRouteMissingBackend
	}
	return nil
}
