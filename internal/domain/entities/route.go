package entities

import (
	domainErrors "api-gateway/internal/domain/errors"
	"regexp"
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

func (r *Route) Match(incomingPath string, incomingMethod string) bool {
	if r.Method != incomingMethod && r.Method != "*" {
		return false
	}
	switch r.PathType {
	case PathTypeExact:
		return incomingPath == (r.Backend.PathPrefix + r.Backend.Id + "/" + r.Path)
	case PathTypePrefix:
		return strings.HasPrefix(incomingPath, r.Path)
	case PathTypeRegEx:
		compiledRegex, err := regexp.Compile(r.Path)
		if err != nil {
			return false
		}
		return compiledRegex.MatchString(incomingPath)
	default:
		return r.Path == incomingPath
	}
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
