package usecases_test

import (
	"api-gateway/pkg/logger"
	"context"
	"errors"
	"testing"

	"api-gateway/internal/application/dto"
	"api-gateway/internal/application/usecases"
	"api-gateway/internal/domain/entities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthValidator is a mock for the AuthValidator port
type MockAuthValidator struct {
	mock.Mock
}

func (m *MockAuthValidator) Validate(ctx context.Context, token string, policy *entities.AuthPolicy) error {
	args := m.Called(ctx, token, policy)
	return args.Error(0)
}

func (m *MockAuthValidator) ExtractToken(ctx context.Context, headers map[string][]string, authType string) (string, error) {
	args := m.Called(headers, authType)
	return args.String(0), args.Error(1)
}

func TestAuthenticateRequestUseCase_Execute_Success(t *testing.T) {
	mockValidator := new(MockAuthValidator)
	log := logger.New("test")

	authPolicy := &entities.AuthPolicy{
		Type:    entities.AuthTypeAPIKey,
		Enabled: true,
	}

	request := &dto.AuthRequest{
		Headers: map[string][]string{
			"X-API-Key": {"key123"},
		},
		Policy: authPolicy,
	}

	mockValidator.On("ExtractToken", request.Headers, entities.AuthTypeAPIKey).
		Return("key123", nil)
	mockValidator.On("Validate", mock.Anything, "key123", authPolicy).
		Return(nil)

	useCase := usecases.NewAuthenticateRequestUseCase(mockValidator, log)

	result, err := useCase.Execute(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Authenticated)
	assert.Empty(t, result.ErrorMessage)

	mockValidator.AssertExpectations(t)
}

func TestAuthenticateRequestUseCase_Execute_NoAuthRequired(t *testing.T) {
	mockValidator := new(MockAuthValidator)
	log := logger.New("test")

	authPolicy := &entities.AuthPolicy{
		Type:    entities.AuthTypeNone,
		Enabled: false,
	}

	request := &dto.AuthRequest{
		Headers: map[string][]string{},
		Policy:  authPolicy,
	}

	useCase := usecases.NewAuthenticateRequestUseCase(mockValidator, log)

	result, err := useCase.Execute(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Authenticated)

	mockValidator.AssertNotCalled(t, "ExtractToken")
	mockValidator.AssertNotCalled(t, "Validate")
}

func TestAuthenticateRequestUseCase_Execute_MissingToken(t *testing.T) {
	mockValidator := new(MockAuthValidator)
	log := logger.New("test")

	authPolicy := &entities.AuthPolicy{
		Type:    entities.AuthTypeAPIKey,
		Enabled: true,
	}

	request := &dto.AuthRequest{
		Headers: map[string][]string{},
		Policy:  authPolicy,
	}

	mockValidator.On("ExtractToken", request.Headers, entities.AuthTypeAPIKey).
		Return("", errors.New("missing auth token"))

	useCase := usecases.NewAuthenticateRequestUseCase(mockValidator, log)

	result, err := useCase.Execute(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "missing auth token")

	mockValidator.AssertExpectations(t)
	mockValidator.AssertNotCalled(t, "Validate")
}

func TestAuthenticateRequestUseCase_Execute_NilPolicy(t *testing.T) {
	mockValidator := new(MockAuthValidator)
	log := logger.New("test")

	request := &dto.AuthRequest{
		Headers: map[string][]string{},
		Policy:  nil,
	}

	useCase := usecases.NewAuthenticateRequestUseCase(mockValidator, log)

	result, err := useCase.Execute(context.Background(), request)

	// Nil policy means no auth required
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Authenticated)
}
