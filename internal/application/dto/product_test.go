package dto

import (
	"encoding/json"
	"testing"
	"time"

	"product-service/internal/domain/entities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateProductRequestDTO_ToEntity(t *testing.T) {
	tests := []struct {
		name          string
		dto           CreateProductRequestDTO
		expectError   bool
		errorContains string
	}{
		{
			name: "valid conversion",
			dto: CreateProductRequestDTO{
				Name:        "iPhone 15",
				Description: "Latest Apple smartphone",
				SKU:         "IPH15-128GB",
				Price:       999.99,
				Category:    "Electronics",
				Brand:       "Apple",
				Stock:       100,
			},
			expectError: false,
		},
		{
			name: "empty name",
			dto: CreateProductRequestDTO{
				Name:        "",
				Description: "Description",
				SKU:         "SKU123",
				Price:       100.0,
				Category:    "Electronics",
				Brand:       "Brand",
				Stock:       10,
			},
			expectError:   true,
			errorContains: "product name is required",
		},
		{
			name: "invalid SKU",
			dto: CreateProductRequestDTO{
				Name:        "Product Name",
				Description: "Description",
				SKU:         "AB", // Too short
				Price:       100.0,
				Category:    "Electronics",
				Brand:       "Brand",
				Stock:       10,
			},
			expectError:   true,
			errorContains: "SKU must be at least 3 characters",
		},
		{
			name: "negative price",
			dto: CreateProductRequestDTO{
				Name:        "Product Name",
				Description: "Description",
				SKU:         "SKU123",
				Price:       -10.0,
				Category:    "Electronics",
				Brand:       "Brand",
				Stock:       10,
			},
			expectError:   true,
			errorContains: "price cannot be negative",
		},
		{
			name: "negative stock",
			dto: CreateProductRequestDTO{
				Name:        "Product Name",
				Description: "Description",
				SKU:         "SKU123",
				Price:       100.0,
				Category:    "Electronics",
				Brand:       "Brand",
				Stock:       -5,
			},
			expectError:   true,
			errorContains: "stock cannot be negative",
		},
		{
			name: "empty category",
			dto: CreateProductRequestDTO{
				Name:        "Product Name",
				Description: "Description",
				SKU:         "SKU123",
				Price:       100.0,
				Category:    "",
				Brand:       "Brand",
				Stock:       10,
			},
			expectError:   true,
			errorContains: "category is required",
		},
		{
			name: "empty brand should be valid",
			dto: CreateProductRequestDTO{
				Name:        "Product Name",
				Description: "Description",
				SKU:         "SKU123",
				Price:       100.0,
				Category:    "Electronics",
				Brand:       "",
				Stock:       10,
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entity, err := tt.dto.ToEntity()

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, entity)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, entity)
				assert.Equal(t, tt.dto.Name, entity.Name)
				assert.Equal(t, tt.dto.Description, entity.Description)
				assert.Equal(t, tt.dto.SKU, entity.SKU)
				assert.Equal(t, tt.dto.Price, entity.Price)
				assert.Equal(t, tt.dto.Category, entity.Category)
				assert.Equal(t, tt.dto.Brand, entity.Brand)
				assert.Equal(t, tt.dto.Stock, entity.Stock)
				assert.Equal(t, entities.ProductStatusActive, entity.Status)
			}
		})
	}
}

func TestProductToResponseDTO(t *testing.T) {
	// Given
	now := time.Now()
	product := &entities.Product{
		ID:          1,
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "IPH15-128GB",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
		Status:      entities.ProductStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// When
	dto := ProductToResponseDTO(product)

	// Then
	assert.NotNil(t, dto)
	assert.Equal(t, product.ID, dto.ID)
	assert.Equal(t, product.Name, dto.Name)
	assert.Equal(t, product.Description, dto.Description)
	assert.Equal(t, product.SKU, dto.SKU)
	assert.Equal(t, product.Price, dto.Price)
	assert.Equal(t, product.Category, dto.Category)
	assert.Equal(t, product.Brand, dto.Brand)
	assert.Equal(t, product.Stock, dto.Stock)
	assert.Equal(t, product.Status, dto.Status)
	assert.True(t, dto.IsActive)    // Product is active
	assert.True(t, dto.IsInStock)   // Product has stock > 0
	assert.True(t, dto.IsAvailable) // Product is active and in stock
	assert.Equal(t, product.CreatedAt, dto.CreatedAt)
	assert.Equal(t, product.UpdatedAt, dto.UpdatedAt)
}

func TestProductToResponseDTO_InactiveProduct(t *testing.T) {
	// Given
	product := &entities.Product{
		ID:     1,
		Name:   "Discontinued Product",
		SKU:    "DISC-001",
		Price:  100.0,
		Stock:  10,
		Status: entities.ProductStatusDiscontinued,
	}

	// When
	dto := ProductToResponseDTO(product)

	// Then
	assert.NotNil(t, dto)
	assert.False(t, dto.IsActive)    // Product is discontinued
	assert.True(t, dto.IsInStock)    // Product has stock > 0
	assert.False(t, dto.IsAvailable) // Product is not available (inactive)
}

func TestProductToResponseDTO_OutOfStock(t *testing.T) {
	// Given
	product := &entities.Product{
		ID:     1,
		Name:   "Out of Stock Product",
		SKU:    "OOS-001",
		Price:  100.0,
		Stock:  0, // No stock
		Status: entities.ProductStatusActive,
	}

	// When
	dto := ProductToResponseDTO(product)

	// Then
	assert.NotNil(t, dto)
	assert.True(t, dto.IsActive)     // Product is active
	assert.False(t, dto.IsInStock)   // Product has no stock
	assert.False(t, dto.IsAvailable) // Product is not available (out of stock)
}

func TestProductsToResponseDTOs(t *testing.T) {
	// Given
	now := time.Now()
	products := []*entities.Product{
		{
			ID:          1,
			Name:        "iPhone 15",
			Description: "Latest Apple smartphone",
			SKU:         "IPH15-128GB",
			Price:       999.99,
			Category:    "Electronics",
			Brand:       "Apple",
			Stock:       100,
			Status:      entities.ProductStatusActive,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          2,
			Name:        "Samsung Galaxy S24",
			Description: "Latest Samsung smartphone",
			SKU:         "SGS24-128GB",
			Price:       899.99,
			Category:    "Electronics",
			Brand:       "Samsung",
			Stock:       50,
			Status:      entities.ProductStatusActive,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	// When
	dtos := ProductsToResponseDTOs(products)

	// Then
	assert.Len(t, dtos, 2)

	assert.Equal(t, products[0].ID, dtos[0].ID)
	assert.Equal(t, products[0].Name, dtos[0].Name)
	assert.Equal(t, products[0].SKU, dtos[0].SKU)
	assert.True(t, dtos[0].IsAvailable)

	assert.Equal(t, products[1].ID, dtos[1].ID)
	assert.Equal(t, products[1].Name, dtos[1].Name)
	assert.Equal(t, products[1].SKU, dtos[1].SKU)
	assert.True(t, dtos[1].IsAvailable)
}

func TestCreateProductRequestDTO_JSONSerialization(t *testing.T) {
	// Given
	dto := CreateProductRequestDTO{
		Name:        "iPhone 15",
		Description: "Latest Apple smartphone",
		SKU:         "IPH15-128GB",
		Price:       999.99,
		Category:    "Electronics",
		Brand:       "Apple",
		Stock:       100,
	}

	// When - Serialize to JSON
	jsonData, err := json.Marshal(dto)
	require.NoError(t, err)

	// Then - Deserialize back
	var decodedDTO CreateProductRequestDTO
	err = json.Unmarshal(jsonData, &decodedDTO)
	require.NoError(t, err)

	assert.Equal(t, dto.Name, decodedDTO.Name)
	assert.Equal(t, dto.Description, decodedDTO.Description)
	assert.Equal(t, dto.SKU, decodedDTO.SKU)
	assert.Equal(t, dto.Price, decodedDTO.Price)
	assert.Equal(t, dto.Category, decodedDTO.Category)
	assert.Equal(t, dto.Brand, decodedDTO.Brand)
	assert.Equal(t, dto.Stock, decodedDTO.Stock)
}

func TestProductResponseDTO_JSONSerialization(t *testing.T) {
	// Given
	now := time.Now()
	dto := ProductResponseDTO{
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
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	// When - Serialize to JSON
	jsonData, err := json.Marshal(dto)
	require.NoError(t, err)

	// Then - Deserialize back
	var decodedDTO ProductResponseDTO
	err = json.Unmarshal(jsonData, &decodedDTO)
	require.NoError(t, err)

	assert.Equal(t, dto.ID, decodedDTO.ID)
	assert.Equal(t, dto.Name, decodedDTO.Name)
	assert.Equal(t, dto.Description, decodedDTO.Description)
	assert.Equal(t, dto.SKU, decodedDTO.SKU)
	assert.Equal(t, dto.Price, decodedDTO.Price)
	assert.Equal(t, dto.Category, decodedDTO.Category)
	assert.Equal(t, dto.Brand, decodedDTO.Brand)
	assert.Equal(t, dto.Stock, decodedDTO.Stock)
	assert.Equal(t, dto.Status, decodedDTO.Status)
	assert.Equal(t, dto.IsActive, decodedDTO.IsActive)
	assert.Equal(t, dto.IsInStock, decodedDTO.IsInStock)
	assert.Equal(t, dto.IsAvailable, decodedDTO.IsAvailable)
}

func TestUpdateProductRequestDTO_PartialUpdate(t *testing.T) {
	// Given - Only some fields provided
	newPrice := 899.99
	dto := UpdateProductRequestDTO{
		Name:  "iPhone 15 Pro",
		Price: &newPrice,
		// Other fields are omitted
	}

	// When - Serialize to JSON
	jsonData, err := json.Marshal(dto)
	require.NoError(t, err)

	// Then - Verify structure
	var decoded map[string]interface{}
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "iPhone 15 Pro", decoded["name"])
	assert.Equal(t, 899.99, decoded["price"])
	assert.Equal(t, "", decoded["description"]) // Empty string, not nil
	assert.Equal(t, "", decoded["category"])    // Empty string, not nil
	assert.Equal(t, "", decoded["brand"])       // Empty string, not nil
	assert.Nil(t, decoded["stock"])             // Nil for pointer field not set
}

func TestUpdateProductRequestDTO_WithPointers(t *testing.T) {
	// Given
	newPrice := 799.99
	newStock := 200
	dto := UpdateProductRequestDTO{
		Name:  "Updated Product",
		Price: &newPrice,
		Stock: &newStock,
	}

	// When - Serialize to JSON
	jsonData, err := json.Marshal(dto)
	require.NoError(t, err)

	// Then - Deserialize back
	var decodedDTO UpdateProductRequestDTO
	err = json.Unmarshal(jsonData, &decodedDTO)
	require.NoError(t, err)

	assert.Equal(t, dto.Name, decodedDTO.Name)
	require.NotNil(t, decodedDTO.Price)
	assert.Equal(t, *dto.Price, *decodedDTO.Price)
	require.NotNil(t, decodedDTO.Stock)
	assert.Equal(t, *dto.Stock, *decodedDTO.Stock)
}

func TestProductListResponseDTO_Structure(t *testing.T) {
	// Given
	products := []*ProductResponseDTO{
		{ID: 1, Name: "iPhone 15", SKU: "IPH15-128GB", IsAvailable: true},
		{ID: 2, Name: "Samsung Galaxy S24", SKU: "SGS24-128GB", IsAvailable: true},
	}

	dto := ProductListResponseDTO{
		Products: products,
		Total:    10,
		Page:     1,
		PageSize: 2,
	}

	// When - Serialize to JSON
	jsonData, err := json.Marshal(dto)
	require.NoError(t, err)

	// Then - Deserialize and verify structure
	var decoded ProductListResponseDTO
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Len(t, decoded.Products, 2)
	assert.Equal(t, 10, decoded.Total)
	assert.Equal(t, 1, decoded.Page)
	assert.Equal(t, 2, decoded.PageSize)
}

func TestStockUpdateRequestDTO_Validation(t *testing.T) {
	tests := []struct {
		name  string
		stock int
		valid bool
	}{
		{"valid positive stock", 100, true},
		{"valid zero stock", 0, true},
		{"invalid negative stock", -10, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := StockUpdateRequestDTO{Stock: tt.stock}

			// Serialize to JSON to test structure
			jsonData, err := json.Marshal(dto)
			require.NoError(t, err)

			var decoded StockUpdateRequestDTO
			err = json.Unmarshal(jsonData, &decoded)
			require.NoError(t, err)

			assert.Equal(t, tt.stock, decoded.Stock)
		})
	}
}

func TestPriceUpdateRequestDTO_Validation(t *testing.T) {
	tests := []struct {
		name  string
		price float64
		valid bool
	}{
		{"valid positive price", 99.99, true},
		{"valid zero price", 0.0, true},
		{"invalid negative price", -10.0, false},
		{"valid maximum price", 999999.99, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := PriceUpdateRequestDTO{Price: tt.price}

			// Serialize to JSON to test structure
			jsonData, err := json.Marshal(dto)
			require.NoError(t, err)

			var decoded PriceUpdateRequestDTO
			err = json.Unmarshal(jsonData, &decoded)
			require.NoError(t, err)

			assert.Equal(t, tt.price, decoded.Price)
		})
	}
}

func TestProductSearchRequestDTO_Structure(t *testing.T) {
	// Given
	minPrice := 100.0
	maxPrice := 1000.0
	inStock := true
	status := entities.ProductStatusActive

	dto := ProductSearchRequestDTO{
		Query:    "smartphone",
		Category: "Electronics",
		Brand:    "Apple",
		MinPrice: &minPrice,
		MaxPrice: &maxPrice,
		InStock:  &inStock,
		Status:   &status,
		Page:     0,
		PageSize: 10,
	}

	// When - Serialize to JSON
	jsonData, err := json.Marshal(dto)
	require.NoError(t, err)

	// Then - Deserialize and verify structure
	var decoded ProductSearchRequestDTO
	err = json.Unmarshal(jsonData, &decoded)
	require.NoError(t, err)

	assert.Equal(t, dto.Query, decoded.Query)
	assert.Equal(t, dto.Category, decoded.Category)
	assert.Equal(t, dto.Brand, decoded.Brand)
	require.NotNil(t, decoded.MinPrice)
	assert.Equal(t, *dto.MinPrice, *decoded.MinPrice)
	require.NotNil(t, decoded.MaxPrice)
	assert.Equal(t, *dto.MaxPrice, *decoded.MaxPrice)
	require.NotNil(t, decoded.InStock)
	assert.Equal(t, *dto.InStock, *decoded.InStock)
	require.NotNil(t, decoded.Status)
	assert.Equal(t, *dto.Status, *decoded.Status)
	assert.Equal(t, dto.Page, decoded.Page)
	assert.Equal(t, dto.PageSize, decoded.PageSize)
}
