package usecases

import (
	"api-gateway/internal/application/dto"
	"api-gateway/internal/application/ports"
	"api-gateway/internal/domain/entities"
	"api-gateway/pkg/logger"
	"context"
	"time"
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
	log.Info("Initializing authentication use case")

	return &authenticationUseCasesImpl{
		authValidator: authValidator,
		logger:        log.With("component", "authentication_usecases"),
	}
}

func (a authenticationUseCasesImpl) Execute(ctx context.Context, req *dto.AuthRequest) (*dto.AuthResponse, error) {
	startTime := time.Now()

	a.logger.Info("Starting authentication request",
		"has_policy", req.Policy != nil,
		"headers_count", len(req.Headers),
	)

	defer func() {
		a.logger.Info("Authentication request completed",
			"duration_ms", time.Since(startTime).Milliseconds(),
		)
	}()

	authRespone := dto.AuthResponse{}

	// Check for no authentication policy
	if req.Policy == nil || req.Policy.Type == entities.AuthTypeNone {
		a.logger.Info("No authentication required",
			"policy_type", "none",
			"authenticated", true,
		)

		return &dto.AuthResponse{
			Authenticated: true,
			ErrorMessage:  "",
			UserID:        "",
		}, nil
	}

	a.logger.Debug("Authentication policy detected",
		"policy_type", req.Policy.Type,
		"policy_enabled", req.Policy.Enabled,
	)

	// Handle API Key authentication
	if req.Policy.Type == entities.AuthTypeAPIKey {
		a.logger.Info("Processing API Key authentication",
			"policy_type", req.Policy.Type,
		)

		a.logger.Debug("Extracting token from headers",
			"policy_type", req.Policy.Type,
			"headers_count", len(req.Headers),
		)

		token, err := a.authValidator.ExtractToken(ctx, req.Headers, req.Policy.Type)
		if err != nil {
			a.logger.Warn("Token extraction failed",
				"error", err,
				"policy_type", req.Policy.Type,
				"duration_ms", time.Since(startTime).Milliseconds(),
			)
			return nil, err
		}

		a.logger.Debug("Token extracted successfully",
			"token_length", len(token),
			"policy_type", req.Policy.Type,
		)

		a.logger.Info("Validating token",
			"policy_type", req.Policy.Type,
		)

		err = a.authValidator.Validate(ctx, token, req.Policy)
		if err != nil {
			a.logger.Warn("Token validation failed",
				"error", err,
				"policy_type", req.Policy.Type,
				"duration_ms", time.Since(startTime).Milliseconds(),
			)
			return nil, err
		}

		a.logger.Info("Token validated successfully",
			"policy_type", req.Policy.Type,
			"authenticated", true,
		)

		authRespone = dto.AuthResponse{
			Authenticated: true,
			UserID:        "",
			ErrorMessage:  "",
		}

		a.logger.Info("API Key authentication successful",
			"authenticated", authRespone.Authenticated,
			"duration_ms", time.Since(startTime).Milliseconds(),
		)

		return &authRespone, nil
	}

	a.logger.Warn("Unsupported authentication policy type",
		"policy_type", req.Policy.Type,
		"duration_ms", time.Since(startTime).Milliseconds(),
	)

	return nil, nil
}
