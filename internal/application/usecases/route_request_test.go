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

func (m *MockRouteRepository) GetAll(ctx context.Context) ([]entities.Route, error) {
	args := m.Called(ctx)
	return args.Get(0).([]entities.Route), args.Error(1)
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

	request := &dto.GatewayRequest{
		Path:    "/api/users",
		Method:  "GET",
		Headers: map[string][]string{"Content-Type": {"application/json"}},
		Body:    []byte{},
		Host:    "http://service:8080",
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

	mockProxy.On("Forward", mock.Anything, expectedProxyReq).Return(proxyResponse, nil)

	useCase := usecases.NewRouteRequestUseCase(mockProxy, mockRepo, log)

	response, err := useCase.Execute(context.Background(), request)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, http.StatusOK, response.StatusCode)
	assert.Equal(t, proxyResponse.Body, response.Body)

	mockProxy.AssertExpectations(t)
}

func TestRouteRequestUseCase_Execute_ProxyError(t *testing.T) {
	mockRepo := new(MockRouteRepository)
	mockProxy := new(MockProxyClient)
	log := logger.New("test")

	request := &dto.GatewayRequest{
		Path:   "/api/users",
		Method: "GET",
		Host:   "http://service:8080",
	}

	mockProxy.On("Forward", mock.Anything, mock.Anything).
		Return(nil, errors.New("connection timeout"))

	useCase := usecases.NewRouteRequestUseCase(mockProxy, mockRepo, log)

	response, err := useCase.Execute(context.Background(), request)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.Contains(t, err.Error(), "connection timeout")

	mockProxy.AssertExpectations(t)
}
