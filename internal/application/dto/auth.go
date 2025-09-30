package dto

import "api-gateway/internal/domain/entities"

type AuthRequest struct {
	Headers map[string][]string
	Policy  *entities.AuthPolicy
}

type AuthResponse struct {
	Authenticated bool
	UserID        string
	ErrorMessage  string
}
