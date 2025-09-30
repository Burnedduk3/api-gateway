package auth

import (
	"api-gateway/internal/application/ports"
	"api-gateway/internal/domain/entities"
	"api-gateway/pkg/logger"
	"context"
	"errors"
)

type ValidatorRepository struct {
	apiKeyRepo ports.ApiKeyRepository
	logger     logger.Logger
}

func NewAuthValidator(log logger.Logger, apiKeyRepo ports.ApiKeyRepository) ports.AuthValidator {
	return &ValidatorRepository{
		logger:     log,
		apiKeyRepo: apiKeyRepo,
	}
}

func (v ValidatorRepository) Validate(ctx context.Context, token string, policy *entities.AuthPolicy) error {
	isValidToken, err := v.apiKeyRepo.IsValidKey(ctx, token)
	if err != nil {
		return err
	}
	if !isValidToken {
		v.logger.Error("invalid api key", "token", token)
		return errors.New("invalid api key")
	}
	return nil
}

func (v ValidatorRepository) ExtractToken(ctx context.Context, headers map[string][]string, authType string) (string, error) {
	apiToken := ""
	if authType == entities.AuthTypeAPIKey {
		apiTokenHeader := headers["X-Api-Key"]
		if len(apiTokenHeader) == 0 {
			return "", errors.New("missing api key")
		}
		apiToken = apiTokenHeader[0]
		isValid, err := v.apiKeyRepo.IsValidKey(ctx, apiToken)
		if err != nil {
			return "", err
		}
		if !isValid {
			return "", errors.New("invalid api key")
		}
		return apiToken, nil
	}
	return "apiToken", errors.New("invalid auth type")
}
