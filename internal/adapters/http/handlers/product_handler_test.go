package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"product-service/internal/application/dto"
	"product-service/internal/domain/entities"
	domainErrors "product-service/internal/domain/errors"
	"product-service/pkg/logger"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockProductUseCases implements the ProductUseCases interface for testing
type MockProductUseCases struct {
	mock.Mock
}

func (m *MockProductUseCases) CreateProduct(ctx context.Context, request *dto.CreateProductRequestDTO) (*dto.ProductResponseDTO, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResponseDTO), args.Error(1)
}

func (m *MockProductUseCases) GetProductByID(ctx context.Context, id uint) (*dto.ProductResponseDTO, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResponseDTO), args.Error(1)
}

func (m *MockProductUseCases) GetProductBySKU(ctx context.Context, sku string) (*dto.ProductResponseDTO, error) {
	args := m.Called(ctx, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResponseDTO), args.Error(1)
}

func (m *MockProductUseCases) UpdateProduct(ctx context.Context, id uint, request *dto.UpdateProductRequestDTO) (*dto.ProductResponseDTO, error) {
	args := m.Called(ctx, id, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResponseDTO), args.Error(1)
}

func (m *MockProductUseCases) UpdateProductStock(ctx context.Context, id uint, stock int) (*dto.ProductResponseDTO, error) {
	args := m.Called(ctx, id, stock)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResponseDTO), args.Error(1)
}

func (m *MockProductUseCases) UpdateProductPrice(ctx context.Context, id uint, price float64) (*dto.ProductResponseDTO, error) {
	args := m.Called(ctx, id, price)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResponseDTO), args.Error(1)
}

func (m *MockProductUseCases) ActivateProduct(ctx context.Context, id uint) (*dto.ProductResponseDTO, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResponseDTO), args.Error(1)
}

func (m *MockProductUseCases) DeactivateProduct(ctx context.Context, id uint) (*dto.ProductResponseDTO, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResponseDTO), args.Error(1)
}

func (m *MockProductUseCases) DiscontinueProduct(ctx context.Context, id uint) (*dto.ProductResponseDTO, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductResponseDTO), args.Error(1)
}

func (m *MockProductUseCases) ListProducts(ctx context.Context, page, pageSize int) (*dto.ProductListResponseDTO, error) {
	args := m.Called(ctx, page, pageSize)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProductListResponseDTO), args.Error(1)
}

func setupTestHandler() (*ProductHandler, *MockProductUseCases) {
	mockUseCases := new(MockProductUseCases)
	log := logger.New("test")
	handler := NewProductHandler(mockUseCases, log)
	return handler, mockUseCases
}

func TestProductHandler_CreateProduct_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestHandler()

	requestBody := dto.CreateProductRequestDTO{
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "IPH15-128GB",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
	}

	expectedResponse := &dto.ProductResponseDTO{
		ID:          1,
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "IPH15-128GB",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
		Status:      entities.ProductStatusActive,
		IsActive:    true,
		IsInStock:   true,
		IsAvailable: true,
	}

	mockUseCases.On("CreateProduct", mock.Anything, &requestBody).Return(expectedResponse, nil)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Execute
	err := handler.CreateProduct(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response dto.ProductResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedResponse.ID, response.ID)
	assert.Equal(t, expectedResponse.Name, response.Name)
	assert.Equal(t, expectedResponse.SKU, response.SKU)

	mockUseCases.AssertExpectations(t)
}

func TestProductHandler_CreateProduct_ValidationError(t *testing.T) {
	// Setup
	handler, _ := setupTestHandler()

	requestBody := dto.CreateProductRequestDTO{
		Name:        "", // Required field missing
		Description: "Description",
		SKU:         "AB",  // Too short
		Price:       -10.0, // Invalid negative price
		Category:    "",    // Required field missing
		Brand:       "Brand",
		Stock:       -5, // Invalid negative stock
	}

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Execute
	err := handler.CreateProduct(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "VALIDATION_ERROR", response.Error)
	assert.NotNil(t, response.Details)
}

func TestProductHandler_CreateProduct_ProductAlreadyExists(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestHandler()

	requestBody := dto.CreateProductRequestDTO{
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "EXISTING-SKU",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
	}

	mockUseCases.On("CreateProduct", mock.Anything, &requestBody).Return(nil, domainErrors.ErrProductAlreadyExists)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/api/v1/products", bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Execute
	err := handler.CreateProduct(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, rec.Code)

	var response ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "PRODUCT_ALREADY_EXISTS", response.Error)
	mockUseCases.AssertExpectations(t)
}

func TestProductHandler_GetProduct_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestHandler()

	expectedResponse := &dto.ProductResponseDTO{
		ID:          1,
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "IPH15-128GB",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
		Status:      entities.ProductStatusActive,
		IsActive:    true,
		IsInStock:   true,
		IsAvailable: true,
	}

	mockUseCases.On("GetProductByID", mock.Anything, uint(1)).Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/1", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Execute
	err := handler.GetProduct(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.ProductResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedResponse.ID, response.ID)
	assert.Equal(t, expectedResponse.Name, response.Name)
	assert.Equal(t, expectedResponse.SKU, response.SKU)

	mockUseCases.AssertExpectations(t)
}

func TestProductHandler_GetProduct_NotFound(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestHandler()

	mockUseCases.On("GetProductByID", mock.Anything, uint(999)).Return(nil, domainErrors.ErrProductNotFound)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/999", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("999")

	// Execute
	err := handler.GetProduct(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "PRODUCT_NOT_FOUND", response.Error)
	mockUseCases.AssertExpectations(t)
}

func TestProductHandler_GetProduct_InvalidID(t *testing.T) {
	// Setup
	handler, _ := setupTestHandler()

	// Create request with invalid ID
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/invalid", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("invalid")

	// Execute
	err := handler.GetProduct(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response ErrorResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "INVALID_ID", response.Error)
}

func TestProductHandler_GetProductBySKU_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestHandler()

	expectedResponse := &dto.ProductResponseDTO{
		ID:          1,
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "IPH15-128GB",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Status:      entities.ProductStatusActive,
		IsActive:    true,
		IsInStock:   true,
		IsAvailable: true,
	}

	mockUseCases.On("GetProductBySKU", mock.Anything, "IPH15-128GB").Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products/sku/IPH15-128GB", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("sku")
	c.SetParamValues("IPH15-128GB")

	// Execute
	err := handler.GetProductBySKU(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.ProductResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedResponse.SKU, response.SKU)
	assert.Equal(t, expectedResponse.Name, response.Name)

	mockUseCases.AssertExpectations(t)
}

func TestProductHandler_UpdateProduct_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestHandler()

	newPrice := 899.99
	requestBody := dto.UpdateProductRequestDTO{
		Name:        "iPhone 15 Pro",
		Description: "Updated description",
		Price:       &newPrice,
	}

	expectedResponse := &dto.ProductResponseDTO{
		ID:          1,
		Name:        "iPhone 15 Pro",
		Description: "Updated description",
		SKU:         "IPH15-128GB",
		Price:       899.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
		Status:      entities.ProductStatusActive,
		IsActive:    true,
		IsInStock:   true,
		IsAvailable: true,
	}

	mockUseCases.On("UpdateProduct", mock.Anything, uint(1), &requestBody).Return(expectedResponse, nil)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPut, "/api/v1/products/1", bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Execute
	err := handler.UpdateProduct(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.ProductResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, expectedResponse.Name, response.Name)
	assert.Equal(t, expectedResponse.Price, response.Price)

	mockUseCases.AssertExpectations(t)
}

func TestProductHandler_UpdateProductStock_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestHandler()

	requestBody := dto.StockUpdateRequestDTO{
		Stock: 150,
	}

	expectedResponse := &dto.ProductResponseDTO{
		ID:          1,
		Name:        "iPhone 15",
		SKU:         "IPH15-128GB",
		Stock:       150,
		Status:      entities.ProductStatusActive,
		IsActive:    true,
		IsInStock:   true,
		IsAvailable: true,
	}

	mockUseCases.On("UpdateProductStock", mock.Anything, uint(1), 150).Return(expectedResponse, nil)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/products/1/stock", bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Execute
	err := handler.UpdateProductStock(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.ProductResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 150, response.Stock)

	mockUseCases.AssertExpectations(t)
}

func TestProductHandler_UpdateProductPrice_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestHandler()

	requestBody := dto.PriceUpdateRequestDTO{
		Price: 799.99,
	}

	expectedResponse := &dto.ProductResponseDTO{
		ID:          1,
		Name:        "iPhone 15",
		SKU:         "IPH15-128GB",
		Price:       799.99,
		Status:      entities.ProductStatusActive,
		IsActive:    true,
		IsInStock:   true,
		IsAvailable: true,
	}

	mockUseCases.On("UpdateProductPrice", mock.Anything, uint(1), 799.99).Return(expectedResponse, nil)

	// Create request
	jsonBody, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/products/1/price", bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)

	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Execute
	err := handler.UpdateProductPrice(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.ProductResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 799.99, response.Price)

	mockUseCases.AssertExpectations(t)
}

func TestProductHandler_ActivateProduct_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestHandler()

	expectedResponse := &dto.ProductResponseDTO{
		ID:          1,
		Name:        "iPhone 15",
		SKU:         "IPH15-128GB",
		Status:      entities.ProductStatusActive,
		IsActive:    true,
		IsInStock:   true,
		IsAvailable: true,
	}

	mockUseCases.On("ActivateProduct", mock.Anything, uint(1)).Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/products/1/activate", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Execute
	err := handler.ActivateProduct(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.ProductResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, entities.ProductStatusActive, response.Status)
	assert.True(t, response.IsActive)

	mockUseCases.AssertExpectations(t)
}

func TestProductHandler_DiscontinueProduct_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestHandler()

	expectedResponse := &dto.ProductResponseDTO{
		ID:          1,
		Name:        "iPhone 15",
		SKU:         "IPH15-128GB",
		Status:      entities.ProductStatusDiscontinued,
		IsActive:    false,
		IsInStock:   true,
		IsAvailable: false,
	}

	mockUseCases.On("DiscontinueProduct", mock.Anything, uint(1)).Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/products/1/discontinue", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("1")

	// Execute
	err := handler.DiscontinueProduct(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.ProductResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, entities.ProductStatusDiscontinued, response.Status)
	assert.False(t, response.IsActive)
	assert.False(t, response.IsAvailable)

	mockUseCases.AssertExpectations(t)
}

func TestProductHandler_ListProducts_Success(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestHandler()

	expectedProducts := []*dto.ProductResponseDTO{
		{
			ID:          1,
			Name:        "iPhone 15",
			SKU:         "IPH15-128GB",
			Price:       999.99,
			Status:      entities.ProductStatusActive,
			IsActive:    true,
			IsInStock:   true,
			IsAvailable: true,
		},
		{
			ID:          2,
			Name:        "Samsung Galaxy S24",
			SKU:         "SGS24-128GB",
			Price:       899.99,
			Status:      entities.ProductStatusActive,
			IsActive:    true,
			IsInStock:   true,
			IsAvailable: true,
		},
	}

	expectedResponse := &dto.ProductListResponseDTO{
		Products: expectedProducts,
		Total:    2,
		Page:     0,
		PageSize: 10,
	}

	mockUseCases.On("ListProducts", mock.Anything, 0, 10).Return(expectedResponse, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Execute
	err := handler.ListProducts(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.ProductListResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Len(t, response.Products, 2)
	assert.Equal(t, 2, response.Total)
	assert.Equal(t, 0, response.Page)

	mockUseCases.AssertExpectations(t)
}

func TestProductHandler_ListProducts_WithPagination(t *testing.T) {
	// Setup
	handler, mockUseCases := setupTestHandler()

	expectedResponse := &dto.ProductListResponseDTO{
		Products: []*dto.ProductResponseDTO{},
		Total:    0,
		Page:     2,
		PageSize: 5,
	}

	mockUseCases.On("ListProducts", mock.Anything, 2, 5).Return(expectedResponse, nil)

	// Create request with pagination parameters
	req := httptest.NewRequest(http.MethodGet, "/api/v1/products?page=2&page_size=5", nil)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	// Execute
	err := handler.ListProducts(c)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var response dto.ProductListResponseDTO
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, response.Page)
	assert.Equal(t, 5, response.PageSize)

	mockUseCases.AssertExpectations(t)
}
