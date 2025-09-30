package handlers_test

import (
	"api-gateway/pkg/logger"
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"api-gateway/internal/adapters/http/handlers"
	"api-gateway/internal/application/dto"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRouteRequestUseCase mocks the route request use case
type MockRouteRequestUseCase struct {
	mock.Mock
}

func (m *MockRouteRequestUseCase) Execute(ctx context.Context, req *dto.GatewayRequest) (*dto.GatewayResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.GatewayResponse), args.Error(1)
}

func (m *MockRouteRequestUseCase) Execute(ctx context.Context, req *dto.GatewayRequest) (*dto.GatewayResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.GatewayResponse), args.Error(1)
}

// MockAuthUseCase mocks the authentication use case
type MockAuthUseCase struct {
	mock.Mock
}

func (m *MockAuthUseCase) Execute(ctx context.Context, req *dto.AuthRequest) (*dto.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.AuthResponse), args.Error(1)
}

func TestGatewayHandler_HandleRequest_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	log := logger.New("test")
	c := e.NewContext(req, rec)

	mockRouteUseCase := new(MockRouteRequestUseCase)
	mockAuthUseCase := new(MockAuthUseCase)

	// Mock auth response (no auth required)
	authResponse := &dto.AuthResponse{
		Authenticated: true,
	}
	mockAuthUseCase.On("Execute", mock.Anything, mock.Anything).Return(authResponse, nil)

	// Mock route response
	gatewayResponse := &dto.GatewayResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       []byte(`{"message":"success"}`),
	}
	mockRouteUseCase.On("Execute", mock.Anything, mock.Anything).Return(gatewayResponse, nil)

	handler := handlers.NewGatewayHandler(log, mockRouteUseCase, mockAuthUseCase)

	err := handler.HandleRequest(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "success")

	mockRouteUseCase.AssertExpectations(t)
	mockAuthUseCase.AssertExpectations(t)
}

func TestGatewayHandler_HandleRequest_AuthFailure(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()
	log := logger.New("test")
	c := e.NewContext(req, rec)

	mockRouteUseCase := new(MockRouteRequestUseCase)
	mockAuthUseCase := new(MockAuthUseCase)

	// Mock auth failure
	mockAuthUseCase.On("Execute", mock.Anything, mock.Anything).
		Return(nil, errors.New("unauthorized"))

	handler := handlers.NewGatewayHandler(log, mockRouteUseCase, mockAuthUseCase)

	err := handler.HandleRequest(c)

	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusUnauthorized, httpErr.Code)

	mockAuthUseCase.AssertExpectations(t)
	mockRouteUseCase.AssertNotCalled(t, "Execute")
}

func TestGatewayHandler_HandleRequest_RouteNotFound(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/unknown", nil)
	rec := httptest.NewRecorder()
	log := logger.New("test")
	c := e.NewContext(req, rec)

	mockRouteUseCase := new(MockRouteRequestUseCase)
	mockAuthUseCase := new(MockAuthUseCase)

	// Mock auth success
	authResponse := &dto.AuthResponse{
		Authenticated: true,
	}
	mockAuthUseCase.On("Execute", mock.Anything, mock.Anything).Return(authResponse, nil)

	// Mock route not found
	mockRouteUseCase.On("Execute", mock.Anything, mock.Anything).
		Return(nil, errors.New("route not found"))

	handler := handlers.NewGatewayHandler(log, mockRouteUseCase, mockAuthUseCase)

	err := handler.HandleRequest(c)

	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusNotFound, httpErr.Code)

	mockRouteUseCase.AssertExpectations(t)
}

func TestGatewayHandler_HandleRequest_WithRequestBody(t *testing.T) {
	e := echo.New()
	body := []byte(`{"name":"John"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	log := logger.New("test")
	c := e.NewContext(req, rec)

	mockRouteUseCase := new(MockRouteRequestUseCase)
	mockAuthUseCase := new(MockAuthUseCase)

	authResponse := &dto.AuthResponse{Authenticated: true}
	mockAuthUseCase.On("Execute", mock.Anything, mock.Anything).Return(authResponse, nil)

	gatewayResponse := &dto.GatewayResponse{
		StatusCode: http.StatusCreated,
		Body:       []byte(`{"id":1,"name":"John"}`),
	}
	mockRouteUseCase.On("Execute", mock.Anything, mock.MatchedBy(func(req *dto.GatewayRequest) bool {
		return bytes.Equal(req.Body, body)
	})).Return(gatewayResponse, nil)

	handler := handlers.NewGatewayHandler(log, mockRouteUseCase, mockAuthUseCase)

	err := handler.HandleRequest(c)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	mockRouteUseCase.AssertExpectations(t)
}

func TestGatewayHandler_HandleRequest_InternalServerError(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/users", nil)
	rec := httptest.NewRecorder()
	log := logger.New("test")
	c := e.NewContext(req, rec)

	mockRouteUseCase := new(MockRouteRequestUseCase)
	mockAuthUseCase := new(MockAuthUseCase)

	authResponse := &dto.AuthResponse{Authenticated: true}
	mockAuthUseCase.On("Execute", mock.Anything, mock.Anything).Return(authResponse, nil)

	// Mock unexpected error
	mockRouteUseCase.On("Execute", mock.Anything, mock.Anything).
		Return(nil, errors.New("database connection failed"))

	handler := handlers.NewGatewayHandler(log, mockRouteUseCase, mockAuthUseCase)

	err := handler.HandleRequest(c)

	assert.Error(t, err)
	httpErr, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusInternalServerError, httpErr.Code)
}
