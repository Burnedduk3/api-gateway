package ports

import (
	"api-gateway/internal/domain/entities"
	"context"
)

// ProductRepository defines the contract for product persistence
type ProductRepository interface {
	// Create a new product
	Create(ctx context.Context, product *entities.Product) (*entities.Product, error)

	// GetByID retrieves a product by its ID
	GetByID(ctx context.Context, id uint) (*entities.Product, error)

	// GetBySKU retrieves a product by its SKU (unique identifier)
	GetBySKU(ctx context.Context, sku string) (*entities.Product, error)

	// ExistsBySKU checks if a product with the given SKU exists
	ExistsBySKU(ctx context.Context, sku string) (bool, error)
}
