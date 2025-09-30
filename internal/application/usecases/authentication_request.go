package usecases

import (
	"api-gateway/internal/application/dto"
	"api-gateway/internal/application/ports"
	"api-gateway/internal/domain/entities"
	"api-gateway/pkg/logger"
	"context"
)

// AuthenticationUseCases defines the interface for product business operations
type AuthenticationUseCases interface {
	Execute(ctx context.Context, req *dto.AuthRequest) (*dto.AuthResponse, error)
}

// authenticationUseCasesImpl implements AuthenticationUseCases interface
type authenticationUseCasesImpl struct {
	logger        logger.Logger
	authValidator ports.AuthValidator
}

// NewAuthenticateRequestUseCase creates a new instance of authentication use cases
func NewAuthenticateRequestUseCase(authValidator ports.AuthValidator, log logger.Logger) AuthenticationUseCases {
	return &authenticationUseCasesImpl{
		authValidator: authValidator,
		logger:        log.With("component", "authentication_usecases"),
	}
}

func (a authenticationUseCasesImpl) Execute(ctx context.Context, req *dto.AuthRequest) (*dto.AuthResponse, error) {
	a.logger.Info("start authentication request")
	defer a.logger.Info("end authentication request")
	authRespone := dto.AuthResponse{}
	if req.Policy == nil || req.Policy.Type == entities.AuthTypeNone {
		return &dto.AuthResponse{
			Authenticated: true,
			ErrorMessage:  "",
			UserID:        "",
		}, nil
	}
	if req.Policy.Type == entities.AuthTypeAPIKey {
		token, err := a.authValidator.ExtractToken(ctx, req.Headers, req.Policy.Type)
		if err != nil {
			return nil, err
		}
		err = a.authValidator.Validate(ctx, token, req.Policy)
		if err != nil {
			return nil, err
		}
		authRespone = dto.AuthResponse{
			Authenticated: true,
			UserID:        "",
			ErrorMessage:  "",
		}
		return &authRespone, nil
	}
	return nil, nil
}
