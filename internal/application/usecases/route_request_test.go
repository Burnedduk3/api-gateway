package usecases_test

import (
	"api-gateway/pkg/logger"
	"context"
	"errors"
	"net/http"
	"testing"

	"api-gateway/internal/application/dto"
	"api-gateway/internal/application/usecases"
	"api-gateway/internal/domain/entities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRouteRepository is a mock for the RouteRepository port
type MockRouteRepository struct {
	mock.Mock
}

func (m *MockRouteRepository) FindByPathAndMethod(ctx context.Context, path, method string) (*entities.Route, error) {
	args := m.Called(ctx, path, method)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Route), args.Error(1)
}

func (m *MockRouteRepository) GetAll(ctx context.Context) ([]*entities.Route, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entities.Route), args.Error(1)
}

func (m *MockRouteRepository) Save(ctx context.Context, route *entities.Route) error {
	args := m.Called(ctx, route)
	return args.Error(0)
}

// MockProxyClient is a mock for the ProxyClient port
type MockProxyClient struct {
	mock.Mock
}

func (m *MockProxyClient) Forward(ctx context.Context, req *dto.ProxyRequest) (*dto.ProxyResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProxyResponse), args.Error(1)
}

func TestRouteRequestUseCase_Execute_Success(t *testing.T) {
	mockRepo := new(MockRouteRepository)
	mockProxy := new(MockProxyClient)
	log := logger.New("test")

	route := &entities.Route{
		ID:      "route-1",
		Path:    "/api/users",
		Method:  "GET",
		Enabled: true,
		Backend: &entities.Backend{
			Host:    "http://service:8080",
			Healthy: true,
		},
	}

	request := &dto.GatewayRequest{
		Path:    "/api/users",
		Method:  "GET",
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    []byte{},
	}

	expectedProxyReq := &dto.ProxyRequest{
		URL:     "http://service:8080/api/users",
		Method:  "GET",
		Headers: request.Headers,
		Body:    request.Body,
	}

	proxyResponse := &dto.ProxyResponse{
		StatusCode: http.StatusOK,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       []byte(`{"message":"success"}`),
	}

	mockRepo.On("FindByPathAndMethod", mock.Anything, "/api/users", "GET").Return(route, nil)
	mockProxy.On("Forward", mock.Anything, expectedProxyReq).Return(proxyResponse, nil)

	useCase := usecases.NewRouteRequestUseCase(mockProxy, mockRepo, log)

	response, err := useCase.Execute(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, proxyResponse.Body, response.Body)

	mockRepo.AssertExpectations(t)
	mockProxy.AssertExpectations(t)
}

func TestRouteRequestUseCase_Execute_RouteNotFound(t *testing.T) {
	mockRepo := new(MockRouteRepository)
	mockProxy := new(MockProxyClient)
	log := logger.New("test")

	request := &dto.GatewayRequest{
		Path:   "/api/unknown",
		Method: "GET",
	}

	mockRepo.On("FindByPathAndMethod", mock.Anything, "/api/unknown", "GET").
		Return(nil, errors.New("route not found"))

	useCase := usecases.NewRouteRequestUseCase(mockProxy, mockRepo, log)

	response, err := useCase.Execute(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "route not found")

	mockRepo.AssertExpectations(t)
	mockProxy.AssertNotCalled(t, "Forward")
}

func TestRouteRequestUseCase_Execute_DisabledRoute(t *testing.T) {
	mockRepo := new(MockRouteRepository)
	mockProxy := new(MockProxyClient)
	log := logger.New("test")

	route := &entities.Route{
		ID:      "route-1",
		Path:    "/api/users",
		Method:  "GET",
		Enabled: false,
		Backend: &entities.Backend{
			Host: "http://service:8080",
		},
	}

	request := &dto.GatewayRequest{
		Path:   "/api/users",
		Method: "GET",
	}

	mockRepo.On("FindByPathAndMethod", mock.Anything, "/api/users", "GET").Return(route, nil)

	useCase := usecases.NewRouteRequestUseCase(mockProxy, mockRepo, log)

	response, err := useCase.Execute(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "disabled")

	mockProxy.AssertNotCalled(t, "Forward")
}

func TestRouteRequestUseCase_Execute_UnhealthyBackend(t *testing.T) {
	mockRepo := new(MockRouteRepository)
	mockProxy := new(MockProxyClient)
	log := logger.New("test")

	route := &entities.Route{
		ID:      "route-1",
		Path:    "/api/users",
		Method:  "GET",
		Enabled: true,
		Backend: &entities.Backend{
			Host:    "http://service:8080",
			Healthy: false,
		},
	}

	request := &dto.GatewayRequest{
		Path:   "/api/users",
		Method: "GET",
	}

	mockRepo.On("FindByPathAndMethod", mock.Anything, "/api/users", "GET").Return(route, nil)

	useCase := usecases.NewRouteRequestUseCase(mockProxy, mockRepo, log)

	response, err := useCase.Execute(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "unhealthy")

	mockProxy.AssertNotCalled(t, "Forward")
}

func TestRouteRequestUseCase_Execute_ProxyError(t *testing.T) {
	mockRepo := new(MockRouteRepository)
	mockProxy := new(MockProxyClient)
	log := logger.New("test")

	route := &entities.Route{
		ID:      "route-1",
		Path:    "/api/users",
		Method:  "GET",
		Enabled: true,
		Backend: &entities.Backend{
			Host:    "http://service:8080",
			Healthy: true,
		},
	}

	request := &dto.GatewayRequest{
		Path:   "/api/users",
		Method: "GET",
	}

	mockRepo.On("FindByPathAndMethod", mock.Anything, "/api/users", "GET").Return(route, nil)
	mockProxy.On("Forward", mock.Anything, mock.Anything).
		Return(nil, errors.New("connection timeout"))

	useCase := usecases.NewRouteRequestUseCase(mockProxy, mockRepo, log)

	response, err := useCase.Execute(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "connection timeout")

	mockRepo.AssertExpectations(t)
	mockProxy.AssertExpectations(t)
}
