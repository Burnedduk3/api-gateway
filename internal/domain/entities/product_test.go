package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewProduct(t *testing.T) {
	tests := []struct {
		name          string
		productName   string
		description   string
		sku           string
		category      string
		brand         string
		price         float64
		stock         int
		expectError   bool
		errorContains string
	}{
		{
			name:        "valid product creation",
			productName: "iPhone 15",
			description: "Latest Apple smartphone",
			sku:         "IPH15-128GB",
			category:    "Electronics",
			brand:       "Apple",
			price:       999.99,
			stock:       100,
			expectError: false,
		},
		{
			name:          "empty product name",
			productName:   "",
			description:   "Description",
			sku:           "SKU123",
			category:      "Category",
			brand:         "Brand",
			price:         100.0,
			stock:         10,
			expectError:   true,
			errorContains: "product name is required",
		},
		{
			name:          "product name too short",
			productName:   "A",
			description:   "Description",
			sku:           "SKU123",
			category:      "Category",
			brand:         "Brand",
			price:         100.0,
			stock:         10,
			expectError:   true,
			errorContains: "product name must be at least 2 characters",
		},
		{
			name:          "empty SKU",
			productName:   "Product Name",
			description:   "Description",
			sku:           "",
			category:      "Category",
			brand:         "Brand",
			price:         100.0,
			stock:         10,
			expectError:   true,
			errorContains: "SKU is required",
		},
		{
			name:          "SKU too short",
			productName:   "Product Name",
			description:   "Description",
			sku:           "AB",
			category:      "Category",
			brand:         "Brand",
			price:         100.0,
			stock:         10,
			expectError:   true,
			errorContains: "SKU must be at least 3 characters",
		},
		{
			name:          "negative price",
			productName:   "Product Name",
			description:   "Description",
			sku:           "SKU123",
			category:      "Category",
			brand:         "Brand",
			price:         -10.0,
			stock:         10,
			expectError:   true,
			errorContains: "price cannot be negative",
		},
		{
			name:          "negative stock",
			productName:   "Product Name",
			description:   "Description",
			sku:           "SKU123",
			category:      "Category",
			brand:         "Brand",
			price:         100.0,
			stock:         -5,
			expectError:   true,
			errorContains: "stock cannot be negative",
		},
		{
			name:          "empty category",
			productName:   "Product Name",
			description:   "Description",
			sku:           "SKU123",
			category:      "",
			brand:         "Brand",
			price:         100.0,
			stock:         10,
			expectError:   true,
			errorContains: "category is required",
		},
		{
			name:        "empty brand should be valid",
			productName: "Product Name",
			description: "Description",
			sku:         "SKU123",
			category:    "Category",
			brand:       "",
			price:       100.0,
			stock:       10,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product, err := NewProduct(tt.productName, tt.description, tt.sku, tt.category, tt.brand, tt.price, tt.stock)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, product)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, product)
				assert.Equal(t, tt.productName, product.Name)
				assert.Equal(t, tt.description, product.Description)
				assert.Equal(t, tt.sku, product.SKU)
				assert.Equal(t, tt.category, product.Category)
				assert.Equal(t, tt.brand, product.Brand)
				assert.Equal(t, tt.price, product.Price)
				assert.Equal(t, tt.stock, product.Stock)
				assert.Equal(t, ProductStatusActive, product.Status)
				assert.WithinDuration(t, time.Now(), product.CreatedAt, time.Second)
				assert.WithinDuration(t, time.Now(), product.UpdatedAt, time.Second)
			}
		})
	}
}

func TestProduct_IsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   ProductStatus
		expected bool
	}{
		{
			name:     "active product",
			status:   ProductStatusActive,
			expected: true,
		},
		{
			name:     "inactive product",
			status:   ProductStatusInactive,
			expected: false,
		},
		{
			name:     "discontinued product",
			status:   ProductStatusDiscontinued,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := &Product{Status: tt.status}
			assert.Equal(t, tt.expected, product.IsActive())
		})
	}
}

func TestProduct_IsInStock(t *testing.T) {
	tests := []struct {
		name     string
		stock    int
		expected bool
	}{
		{
			name:     "product in stock",
			stock:    10,
			expected: true,
		},
		{
			name:     "product out of stock",
			stock:    0,
			expected: false,
		},
		{
			name:     "product with negative stock",
			stock:    -1,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := &Product{Stock: tt.stock}
			assert.Equal(t, tt.expected, product.IsInStock())
		})
	}
}

func TestProduct_IsAvailable(t *testing.T) {
	tests := []struct {
		name     string
		status   ProductStatus
		stock    int
		expected bool
	}{
		{
			name:     "active product in stock",
			status:   ProductStatusActive,
			stock:    10,
			expected: true,
		},
		{
			name:     "active product out of stock",
			status:   ProductStatusActive,
			stock:    0,
			expected: false,
		},
		{
			name:     "inactive product in stock",
			status:   ProductStatusInactive,
			stock:    10,
			expected: false,
		},
		{
			name:     "discontinued product in stock",
			status:   ProductStatusDiscontinued,
			stock:    10,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := &Product{
				Status: tt.status,
				Stock:  tt.stock,
			}
			assert.Equal(t, tt.expected, product.IsAvailable())
		})
	}
}

func TestProduct_Activate(t *testing.T) {
	product := &Product{
		Status:    ProductStatusInactive,
		UpdatedAt: time.Now().Add(-time.Hour),
	}
	oldUpdatedAt := product.UpdatedAt

	product.Activate()

	assert.Equal(t, ProductStatusActive, product.Status)
	assert.True(t, product.UpdatedAt.After(oldUpdatedAt))
}

func TestProduct_Deactivate(t *testing.T) {
	product := &Product{
		Status:    ProductStatusActive,
		UpdatedAt: time.Now().Add(-time.Hour),
	}
	oldUpdatedAt := product.UpdatedAt

	product.Deactivate()

	assert.Equal(t, ProductStatusInactive, product.Status)
	assert.True(t, product.UpdatedAt.After(oldUpdatedAt))
}

func TestProduct_Discontinue(t *testing.T) {
	product := &Product{
		Status:    ProductStatusActive,
		UpdatedAt: time.Now().Add(-time.Hour),
	}
	oldUpdatedAt := product.UpdatedAt

	product.Discontinue()

	assert.Equal(t, ProductStatusDiscontinued, product.Status)
	assert.True(t, product.UpdatedAt.After(oldUpdatedAt))
}

func TestProduct_UpdateStock(t *testing.T) {
	tests := []struct {
		name          string
		initialStock  int
		newStock      int
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid stock update",
			initialStock: 10,
			newStock:     20,
			expectError:  false,
		},
		{
			name:         "update to zero stock",
			initialStock: 10,
			newStock:     0,
			expectError:  false,
		},
		{
			name:          "negative stock",
			initialStock:  10,
			newStock:      -5,
			expectError:   true,
			errorContains: "stock quantity cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := &Product{
				Stock:     tt.initialStock,
				UpdatedAt: time.Now().Add(-time.Hour),
			}
			oldUpdatedAt := product.UpdatedAt

			err := product.UpdateStock(tt.newStock)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Equal(t, tt.initialStock, product.Stock)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.newStock, product.Stock)
				assert.True(t, product.UpdatedAt.After(oldUpdatedAt))
			}
		})
	}
}

func TestProduct_ReduceStock(t *testing.T) {
	tests := []struct {
		name          string
		initialStock  int
		reduction     int
		expectedStock int
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid stock reduction",
			initialStock:  10,
			reduction:     3,
			expectedStock: 7,
			expectError:   false,
		},
		{
			name:          "reduce to zero",
			initialStock:  5,
			reduction:     5,
			expectedStock: 0,
			expectError:   false,
		},
		{
			name:          "insufficient stock",
			initialStock:  5,
			reduction:     10,
			expectError:   true,
			errorContains: "insufficient stock",
		},
		{
			name:          "zero reduction",
			initialStock:  10,
			reduction:     0,
			expectError:   true,
			errorContains: "reduction quantity must be positive",
		},
		{
			name:          "negative reduction",
			initialStock:  10,
			reduction:     -5,
			expectError:   true,
			errorContains: "reduction quantity must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := &Product{
				Stock:     tt.initialStock,
				UpdatedAt: time.Now().Add(-time.Hour),
			}
			oldUpdatedAt := product.UpdatedAt

			err := product.ReduceStock(tt.reduction)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Equal(t, tt.initialStock, product.Stock)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStock, product.Stock)
				assert.True(t, product.UpdatedAt.After(oldUpdatedAt))
			}
		})
	}
}

func TestProduct_AddStock(t *testing.T) {
	tests := []struct {
		name          string
		initialStock  int
		addition      int
		expectedStock int
		expectError   bool
		errorContains string
	}{
		{
			name:          "valid stock addition",
			initialStock:  10,
			addition:      5,
			expectedStock: 15,
			expectError:   false,
		},
		{
			name:          "add to zero stock",
			initialStock:  0,
			addition:      10,
			expectedStock: 10,
			expectError:   false,
		},
		{
			name:          "zero addition",
			initialStock:  10,
			addition:      0,
			expectError:   true,
			errorContains: "addition quantity must be positive",
		},
		{
			name:          "negative addition",
			initialStock:  10,
			addition:      -5,
			expectError:   true,
			errorContains: "addition quantity must be positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := &Product{
				Stock:     tt.initialStock,
				UpdatedAt: time.Now().Add(-time.Hour),
			}
			oldUpdatedAt := product.UpdatedAt

			err := product.AddStock(tt.addition)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Equal(t, tt.initialStock, product.Stock)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedStock, product.Stock)
				assert.True(t, product.UpdatedAt.After(oldUpdatedAt))
			}
		})
	}
}

func TestProduct_UpdatePrice(t *testing.T) {
	tests := []struct {
		name          string
		initialPrice  float64
		newPrice      float64
		expectError   bool
		errorContains string
	}{
		{
			name:         "valid price update",
			initialPrice: 100.0,
			newPrice:     150.0,
			expectError:  false,
		},
		{
			name:         "update to zero price",
			initialPrice: 100.0,
			newPrice:     0.0,
			expectError:  false,
		},
		{
			name:          "negative price",
			initialPrice:  100.0,
			newPrice:      -50.0,
			expectError:   true,
			errorContains: "price cannot be negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			product := &Product{
				Price:     tt.initialPrice,
				UpdatedAt: time.Now().Add(-time.Hour),
			}
			oldUpdatedAt := product.UpdatedAt

			err := product.UpdatePrice(tt.newPrice)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Equal(t, tt.initialPrice, product.Price)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.newPrice, product.Price)
				assert.True(t, product.UpdatedAt.After(oldUpdatedAt))
			}
		})
	}
}

func TestValidateProductName(t *testing.T) {
	tests := []struct {
		name        string
		productName string
		expectError bool
	}{
		{"valid name", "iPhone 15", false},
		{"minimum length", "AB", false},
		{"empty name", "", true},
		{"name too short", "A", true},
		{"name with spaces", " Product Name ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateProductName(tt.productName)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
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

func TestValidatePrice(t *testing.T) {
	tests := []struct {
		name        string
		price       float64
		expectError bool
	}{
		{"valid price", 99.99, false},
		{"zero price", 0.0, false},
		{"maximum price", 999999.99, false},
		{"negative price", -10.0, true},
		{"price too high", 1000000.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePrice(tt.price)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateStock(t *testing.T) {
	tests := []struct {
		name        string
		stock       int
		expectError bool
	}{
		{"valid stock", 100, false},
		{"zero stock", 0, false},
		{"negative stock", -1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateStock(tt.stock)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
