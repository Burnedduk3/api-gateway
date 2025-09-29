package dto

import (
	"api-gateway/internal/domain/entities"
	"time"
)

// CreateProductRequestDTO for product creation
type CreateProductRequestDTO struct {
	Name        string  `json:"name" validate:"required,min=2,max=255"`
	Description string  `json:"description" validate:"omitempty,max=1000"`
	SKU         string  `json:"sku" validate:"required,min=3,max=50"`
	Price       float64 `json:"price" validate:"required,min=0,max=999999.99"`
	Category    string  `json:"category" validate:"required,min=2,max=100"`
	Brand       string  `json:"brand" validate:"omitempty,max=100"`
	Stock       int     `json:"stock" validate:"min=0"`
}

// UpdateProductRequestDTO for product updates
type UpdateProductRequestDTO struct {
	Name        string   `json:"name" validate:"omitempty,min=2,max=255"`
	Description string   `json:"description" validate:"omitempty,max=1000"`
	Category    string   `json:"category" validate:"omitempty,min=2,max=100"`
	Brand       string   `json:"brand" validate:"omitempty,max=100"`
	Price       *float64 `json:"price" validate:"omitempty,min=0,max=999999.99"`
	Stock       *int     `json:"stock" validate:"omitempty,min=0"`
}

// ProductResponseDTO for product responses
type ProductResponseDTO struct {
	ID          uint                   `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	SKU         string                 `json:"sku"`
	Price       float64                `json:"price"`
	Category    string                 `json:"category"`
	Brand       string                 `json:"brand"`
	Stock       int                    `json:"stock"`
	Status      entities.ProductStatus `json:"status"`
	IsActive    bool                   `json:"is_active"`
	IsInStock   bool                   `json:"is_in_stock"`
	IsAvailable bool                   `json:"is_available"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ProductListResponseDTO for paginated product lists
type ProductListResponseDTO struct {
	Products []*ProductResponseDTO `json:"products"`
	Total    int                   `json:"total"`
	Page     int                   `json:"page"`
	PageSize int                   `json:"page_size"`
}

// ProductSearchRequestDTO for product search
type ProductSearchRequestDTO struct {
	Query    string                  `json:"query" validate:"omitempty,min=1,max=255"`
	Category string                  `json:"category" validate:"omitempty,min=2,max=100"`
	Brand    string                  `json:"brand" validate:"omitempty,max=100"`
	MinPrice *float64                `json:"min_price" validate:"omitempty,min=0"`
	MaxPrice *float64                `json:"max_price" validate:"omitempty,min=0"`
	InStock  *bool                   `json:"in_stock"`
	Status   *entities.ProductStatus `json:"status"`
	Page     int                     `json:"page" validate:"min=0"`
	PageSize int                     `json:"page_size" validate:"min=1,max=100"`
}

// StockUpdateRequestDTO for stock updates
type StockUpdateRequestDTO struct {
	Stock int `json:"stock" validate:"min=0"`
}

// PriceUpdateRequestDTO for price updates
type PriceUpdateRequestDTO struct {
	Price float64 `json:"price" validate:"min=0,max=999999.99"`
}

// Conversion methods
func (dto *CreateProductRequestDTO) ToEntity() (*entities.Product, error) {
	return entities.NewProduct(
		dto.Name,
		dto.Description,
		dto.SKU,
		dto.Category,
		dto.Brand,
		dto.Price,
		dto.Stock,
	)
}

func ProductToResponseDTO(product *entities.Product) *ProductResponseDTO {
	return &ProductResponseDTO{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		SKU:         product.SKU,
		Price:       product.Price,
		Category:    product.Category,
		Brand:       product.Brand,
		Stock:       product.Stock,
		Status:      product.Status,
		IsActive:    product.IsActive(),
		IsInStock:   product.IsInStock(),
		IsAvailable: product.IsAvailable(),
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

func ProductsToResponseDTOs(products []*entities.Product) []*ProductResponseDTO {
	dtos := make([]*ProductResponseDTO, 0, len(products))
	for _, product := range products {
		dtos = append(dtos, ProductToResponseDTO(product))
	}
	return dtos
}
