package usecases

import (
	"context"
	"product-service/internal/application/dto"
	"product-service/internal/domain/entities"
	domainErrors "product-service/internal/domain/errors"
	"product-service/pkg/logger"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockProductRepository implements the ProductRepository interface for testing
type MockProductRepository struct {
	mock.Mock
}

func (m *MockProductRepository) Create(ctx context.Context, product *entities.Product) (*entities.Product, error) {
	args := m.Called(ctx, product)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Product), args.Error(1)
}

func (m *MockProductRepository) GetByID(ctx context.Context, id uint) (*entities.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Product), args.Error(1)
}

func (m *MockProductRepository) GetBySKU(ctx context.Context, sku string) (*entities.Product, error) {
	args := m.Called(ctx, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Product), args.Error(1)
}

func (m *MockProductRepository) ExistsBySKU(ctx context.Context, sku string) (bool, error) {
	args := m.Called(ctx, sku)
	return args.Bool(0), args.Error(1)
}

func setupTestUseCases() (ProductUseCases, *MockProductRepository) {
	mockRepo := new(MockProductRepository)
	log := logger.New("test")
	useCases := NewProductUseCases(mockRepo, log)
	return useCases, mockRepo
}

// CreateProduct Tests
func TestProductUseCases_CreateProduct_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	request := &dto.CreateProductRequestDTO{
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "IPH15-128GB",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
	}

	// Mock repository calls
	mockRepo.On("ExistsBySKU", ctx, "IPH15-128GB").Return(false, nil)

	// Expected product to be created
	expectedCreatedProduct := &entities.Product{
		ID:          1,
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "IPH15-128GB",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
		Status:      entities.ProductStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("Create", ctx, mock.MatchedBy(func(product *entities.Product) bool {
		return product.Name == "iPhone 15" &&
			product.SKU == "IPH15-128GB" &&
			product.Price == 999.99 &&
			product.Category == "Electronics" &&
			product.Brand == "Apple" &&
			product.Stock == 100 &&
			product.Status == entities.ProductStatusActive
	})).Return(expectedCreatedProduct, nil)

	// When
	result, err := useCases.CreateProduct(ctx, request)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, "iPhone 15", result.Name)
	assert.Equal(t, "IPH15-128GB", result.SKU)
	assert.Equal(t, 999.99, result.Price)
	assert.Equal(t, "Electronics", result.Category)
	assert.Equal(t, "Apple", result.Brand)
	assert.Equal(t, 100, result.Stock)
	assert.Equal(t, entities.ProductStatusActive, result.Status)

	mockRepo.AssertExpectations(t)
}

func TestProductUseCases_CreateProduct_SKUAlreadyExists(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	request := &dto.CreateProductRequestDTO{
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "EXISTING-SKU",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
	}

	// Mock repository to return true for existing SKU
	mockRepo.On("ExistsBySKU", ctx, "EXISTING-SKU").Return(true, nil)

	// When
	result, err := useCases.CreateProduct(ctx, request)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domainErrors.ErrProductAlreadyExists, err)

	mockRepo.AssertExpectations(t)
}

func TestProductUseCases_CreateProduct_InvalidSKU(t *testing.T) {
	// Given
	useCases, _ := setupTestUseCases()
	ctx := context.Background()

	request := &dto.CreateProductRequestDTO{
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "AB", // Invalid SKU - too short
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
	}

	// When
	result, err := useCases.CreateProduct(ctx, request)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domainErrors.ErrInvalidProductSKU, err)
}

func TestProductUseCases_CreateProduct_RepositoryExistsError(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	request := &dto.CreateProductRequestDTO{
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "IPH15-128GB",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
	}

	// Mock repository to return error when checking if SKU exists
	mockRepo.On("ExistsBySKU", ctx, "IPH15-128GB").Return(false, assert.AnError)

	// When
	result, err := useCases.CreateProduct(ctx, request)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domainErrors.ErrFailedToCheckProductExistance, err)

	mockRepo.AssertExpectations(t)
}

func TestProductUseCases_CreateProduct_RepositoryCreateError(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	request := &dto.CreateProductRequestDTO{
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "IPH15-128GB",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
	}

	// Mock successful SKU check but failed create
	mockRepo.On("ExistsBySKU", ctx, "IPH15-128GB").Return(false, nil)
	mockRepo.On("Create", ctx, mock.Anything).Return(nil, assert.AnError)

	// When
	result, err := useCases.CreateProduct(ctx, request)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domainErrors.ErrFailedToCreateProduct, err)

	mockRepo.AssertExpectations(t)
}

// GetProductByID Tests
func TestProductUseCases_GetProductByID_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	expectedProduct := &entities.Product{
		ID:          1,
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "IPH15-128GB",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
		Status:      entities.ProductStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(expectedProduct, nil)

	// When
	result, err := useCases.GetProductByID(ctx, 1)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, "iPhone 15", result.Name)
	assert.Equal(t, "IPH15-128GB", result.SKU)

	mockRepo.AssertExpectations(t)
}

func TestProductUseCases_GetProductByID_NotFound(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	mockRepo.On("GetByID", ctx, uint(999)).Return(nil, domainErrors.ErrProductNotFound)

	// When
	result, err := useCases.GetProductByID(ctx, 999)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domainErrors.ErrProductNotFound, err)

	mockRepo.AssertExpectations(t)
}

// GetProductBySKU Tests
func TestProductUseCases_GetProductBySKU_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	expectedProduct := &entities.Product{
		ID:          1,
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "IPH15-128GB",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Status:      entities.ProductStatusActive,
	}

	mockRepo.On("GetBySKU", ctx, "IPH15-128GB").Return(expectedProduct, nil)

	// When
	result, err := useCases.GetProductBySKU(ctx, "IPH15-128GB")

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "IPH15-128GB", result.SKU)
	assert.Equal(t, "iPhone 15", result.Name)

	mockRepo.AssertExpectations(t)
}

func TestProductUseCases_GetProductBySKU_NotFound(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	mockRepo.On("GetBySKU", ctx, "NOTFOUND-SKU").Return(nil, domainErrors.ErrProductNotFound)

	// When
	result, err := useCases.GetProductBySKU(ctx, "NOTFOUND-SKU")

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domainErrors.ErrProductNotFound, err)

	mockRepo.AssertExpectations(t)
}

// UpdateProduct Tests
func TestProductUseCases_UpdateProduct_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	existingProduct := &entities.Product{
		ID:          1,
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "IPH15-128GB",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
		Status:      entities.ProductStatusActive,
	}

	newPrice := 899.99
	newStock := 150
	request := &dto.UpdateProductRequestDTO{
		Name:        "iPhone 15 Pro",
		Description: "Updated description",
		Price:       &newPrice,
		Stock:       &newStock,
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingProduct, nil)

	// When
	result, err := useCases.UpdateProduct(ctx, 1, request)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, "iPhone 15 Pro", result.Name)
	assert.Equal(t, "Updated description", result.Description)
	assert.Equal(t, 899.99, result.Price)
	assert.Equal(t, 150, result.Stock)

	mockRepo.AssertExpectations(t)
}

func TestProductUseCases_UpdateProduct_ProductNotFound(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	request := &dto.UpdateProductRequestDTO{
		Name: "Updated Name",
	}

	mockRepo.On("GetByID", ctx, uint(999)).Return(nil, domainErrors.ErrProductNotFound)

	// When
	result, err := useCases.UpdateProduct(ctx, 999, request)

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domainErrors.ErrProductNotFound, err)

	mockRepo.AssertExpectations(t)
}

func TestProductUseCases_UpdateProduct_PartialUpdate(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	existingProduct := &entities.Product{
		ID:          1,
		Name:        "iPhone 15",
		Description: "Original description",
		SKU:         "IPH15-128GB",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
		Status:      entities.ProductStatusActive,
	}

	request := &dto.UpdateProductRequestDTO{
		Name: "iPhone 15 Pro",
		// Only name updated - other fields should remain the same
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingProduct, nil)

	// When
	result, err := useCases.UpdateProduct(ctx, 1, request)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, "iPhone 15 Pro", result.Name)
	assert.Equal(t, "Original description", result.Description) // Should remain unchanged
	assert.Equal(t, 999.99, result.Price)                       // Should remain unchanged
	assert.Equal(t, 100, result.Stock)                          // Should remain unchanged

	mockRepo.AssertExpectations(t)
}

// UpdateProductStock Tests
func TestProductUseCases_UpdateProductStock_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	existingProduct := &entities.Product{
		ID:        1,
		Name:      "iPhone 15",
		SKU:       "IPH15-128GB",
		Stock:     100,
		Status:    entities.ProductStatusActive,
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingProduct, nil)

	// When
	result, err := useCases.UpdateProductStock(ctx, 1, 150)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, 150, result.Stock)

	mockRepo.AssertExpectations(t)
}

func TestProductUseCases_UpdateProductStock_InvalidStock(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	existingProduct := &entities.Product{
		ID:     1,
		Name:   "iPhone 15",
		SKU:    "IPH15-128GB",
		Stock:  100,
		Status: entities.ProductStatusActive,
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingProduct, nil)

	// When
	result, err := useCases.UpdateProductStock(ctx, 1, -10) // Invalid negative stock

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "stock quantity cannot be negative")

	mockRepo.AssertExpectations(t)
}

// UpdateProductPrice Tests
func TestProductUseCases_UpdateProductPrice_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	existingProduct := &entities.Product{
		ID:        1,
		Name:      "iPhone 15",
		SKU:       "IPH15-128GB",
		Price:     999.99,
		Status:    entities.ProductStatusActive,
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingProduct, nil)

	// When
	result, err := useCases.UpdateProductPrice(ctx, 1, 899.99)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, 899.99, result.Price)

	mockRepo.AssertExpectations(t)
}

func TestProductUseCases_UpdateProductPrice_InvalidPrice(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	existingProduct := &entities.Product{
		ID:     1,
		Name:   "iPhone 15",
		SKU:    "IPH15-128GB",
		Price:  999.99,
		Status: entities.ProductStatusActive,
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingProduct, nil)

	// When
	result, err := useCases.UpdateProductPrice(ctx, 1, -100.0) // Invalid negative price

	// Then
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "price cannot be negative")

	mockRepo.AssertExpectations(t)
}

// ActivateProduct Tests
func TestProductUseCases_ActivateProduct_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	existingProduct := &entities.Product{
		ID:        1,
		Name:      "iPhone 15",
		SKU:       "IPH15-128GB",
		Status:    entities.ProductStatusInactive,
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingProduct, nil)

	// When
	result, err := useCases.ActivateProduct(ctx, 1)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, entities.ProductStatusActive, result.Status)

	mockRepo.AssertExpectations(t)
}

// DeactivateProduct Tests
func TestProductUseCases_DeactivateProduct_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	existingProduct := &entities.Product{
		ID:        1,
		Name:      "iPhone 15",
		SKU:       "IPH15-128GB",
		Status:    entities.ProductStatusActive,
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingProduct, nil)

	// When
	result, err := useCases.DeactivateProduct(ctx, 1)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, entities.ProductStatusInactive, result.Status)

	mockRepo.AssertExpectations(t)
}

// DiscontinueProduct Tests
func TestProductUseCases_DiscontinueProduct_Success(t *testing.T) {
	// Given
	useCases, mockRepo := setupTestUseCases()
	ctx := context.Background()

	existingProduct := &entities.Product{
		ID:        1,
		Name:      "iPhone 15",
		SKU:       "IPH15-128GB",
		Status:    entities.ProductStatusActive,
		UpdatedAt: time.Now().Add(-time.Hour),
	}

	mockRepo.On("GetByID", ctx, uint(1)).Return(existingProduct, nil)

	// When
	result, err := useCases.DiscontinueProduct(ctx, 1)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, uint(1), result.ID)
	assert.Equal(t, entities.ProductStatusDiscontinued, result.Status)

	mockRepo.AssertExpectations(t)
}

// ListProducts Tests
func TestProductUseCases_ListProducts_Success(t *testing.T) {
	// Given
	useCases, _ := setupTestUseCases()
	ctx := context.Background()

	// When
	result, err := useCases.ListProducts(ctx, 0, 10)

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.Total) // Empty list since we don't have List method in repository yet
	assert.Equal(t, 0, result.Page)
	assert.Equal(t, 10, result.PageSize)
}

func TestProductUseCases_ListProducts_InvalidPagination(t *testing.T) {
	// Given
	useCases, _ := setupTestUseCases()
	ctx := context.Background()

	// When - Pass invalid pagination parameters
	result, err := useCases.ListProducts(ctx, -1, 150) // Invalid page and page_size

	// Then
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 0, result.Page)      // Should default to 0
	assert.Equal(t, 10, result.PageSize) // Should default to 10
}

func TestValidateSKU(t *testing.T) {
	tests := []struct {
		name        string
		sku         string
		expectError bool
	}{
		{"valid SKU", "IPH15-128", false},
		{"minimum length", "ABC", false},
		{"empty SKU", "", true},
		{"SKU too short", "AB", true},
		{"SKU with spaces", " ABC123 ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSKU(tt.sku)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
