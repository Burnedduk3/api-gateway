package product_repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"product-service/internal/application/ports"
	"product-service/internal/domain/entities"
	domainErrors "product-service/internal/domain/errors"

	"gorm.io/gorm"
)

// ProductModel represents the database model for products
type ProductModel struct {
	ID          uint           `gorm:"primarykey"`
	Name        string         `gorm:"not null;size:255"`
	Description string         `gorm:"size:1000"`
	SKU         string         `gorm:"uniqueIndex;not null;size:50"`
	Price       float64        `gorm:"not null;type:decimal(10,2)"`
	Category    string         `gorm:"not null;size:100"`
	Brand       string         `gorm:"size:100"`
	Stock       int            `gorm:"not null;default:0"`
	Status      string         `gorm:"not null;default:'active';size:20"`
	CreatedAt   time.Time      `gorm:"autoCreateTime"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime"`
	DeletedAt   gorm.DeletedAt `gorm:"index"` // For soft deletes
}

// TableName specifies the table name for GORM
func (ProductModel) TableName() string {
	return "products"
}

// GormProductRepository implements the ProductRepository interface using GORM
type GormProductRepository struct {
	db *gorm.DB
}

// NewGormProductRepository creates a new GORM product repository
func NewGormProductRepository(db *gorm.DB) ports.ProductRepository {
	return &GormProductRepository{db: db}
}

// Create implements ports.ProductRepository
func (r *GormProductRepository) Create(ctx context.Context, product *entities.Product) (*entities.Product, error) {
	// Check if product already exists by SKU
	exists, err := r.ExistsBySKU(ctx, product.SKU)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domainErrors.ErrProductAlreadyExists
	}

	gormModel := r.toModel(product)

	// Create product in database
	if err := r.db.WithContext(ctx).Create(gormModel).Error; err != nil {
		return nil, r.handleError(err)
	}

	return r.toEntity(gormModel), nil
}

// GetByID implements ports.ProductRepository
func (r *GormProductRepository) GetByID(ctx context.Context, id uint) (*entities.Product, error) {
	var model ProductModel

	err := r.db.WithContext(ctx).Where("id = ?", id).First(&model).Error
	if err != nil {
		return nil, r.handleError(err)
	}

	return r.toEntity(&model), nil
}

// GetBySKU implements ports.ProductRepository
func (r *GormProductRepository) GetBySKU(ctx context.Context, sku string) (*entities.Product, error) {
	var model ProductModel

	err := r.db.WithContext(ctx).Where("sku = ?", sku).First(&model).Error
	if err != nil {
		return nil, r.handleError(err)
	}

	return r.toEntity(&model), nil
}

// ExistsBySKU implements ports.ProductRepository
func (r *GormProductRepository) ExistsBySKU(ctx context.Context, sku string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ProductModel{}).Where("sku = ?", sku).Count(&count).Error
	if err != nil {
		return false, domainErrors.ErrFailedToCheckProductExistance
	}

	return count > 0, nil
}

// Update implements ports.ProductRepository (additional method for completeness)
func (r *GormProductRepository) Update(ctx context.Context, product *entities.Product) (*entities.Product, error) {
	gormModel := r.toModel(product)

	err := r.db.WithContext(ctx).Model(gormModel).Where("id = ?", product.ID).Updates(gormModel).Error
	if err != nil {
		return nil, r.handleError(err)
	}

	// Fetch updated record to return
	return r.GetByID(ctx, product.ID)
}

// Delete implements ports.ProductRepository (additional method for completeness)
func (r *GormProductRepository) Delete(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).Delete(&ProductModel{}, id).Error
	if err != nil {
		return r.handleError(err)
	}
	return nil
}

// List implements ports.ProductRepository (additional method for completeness)
func (r *GormProductRepository) List(ctx context.Context, limit, offset int) ([]*entities.Product, error) {
	var models []ProductModel

	err := r.db.WithContext(ctx).Model(&ProductModel{}).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, r.handleError(err)
	}

	return r.toEntities(models), nil
}

// Search implements ports.ProductRepository (additional method for completeness)
func (r *GormProductRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Product, error) {
	var models []ProductModel

	searchQuery := "%" + query + "%"
	err := r.db.WithContext(ctx).Model(&ProductModel{}).
		Where("name ILIKE ? OR description ILIKE ? OR sku ILIKE ?", searchQuery, searchQuery, searchQuery).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, r.handleError(err)
	}

	return r.toEntities(models), nil
}

// GetByCategory implements ports.ProductRepository (additional method for completeness)
func (r *GormProductRepository) GetByCategory(ctx context.Context, category string, limit, offset int) ([]*entities.Product, error) {
	var models []ProductModel

	err := r.db.WithContext(ctx).Model(&ProductModel{}).
		Where("category = ?", category).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, r.handleError(err)
	}

	return r.toEntities(models), nil
}

// GetByBrand implements ports.ProductRepository (additional method for completeness)
func (r *GormProductRepository) GetByBrand(ctx context.Context, brand string, limit, offset int) ([]*entities.Product, error) {
	var models []ProductModel

	err := r.db.WithContext(ctx).Model(&ProductModel{}).
		Where("brand = ?", brand).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, r.handleError(err)
	}

	return r.toEntities(models), nil
}

// GetByStatus implements ports.ProductRepository (additional method for completeness)
func (r *GormProductRepository) GetByStatus(ctx context.Context, status entities.ProductStatus, limit, offset int) ([]*entities.Product, error) {
	var models []ProductModel

	err := r.db.WithContext(ctx).Model(&ProductModel{}).
		Where("status = ?", string(status)).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, r.handleError(err)
	}

	return r.toEntities(models), nil
}

// GetLowStockProducts implements ports.ProductRepository (additional method for completeness)
func (r *GormProductRepository) GetLowStockProducts(ctx context.Context, threshold int, limit, offset int) ([]*entities.Product, error) {
	var models []ProductModel

	err := r.db.WithContext(ctx).Model(&ProductModel{}).
		Where("stock <= ?", threshold).
		Limit(limit).
		Offset(offset).
		Order("stock ASC").
		Find(&models).Error

	if err != nil {
		return nil, r.handleError(err)
	}

	return r.toEntities(models), nil
}

// UpdateStock implements ports.ProductRepository (additional method for completeness)
func (r *GormProductRepository) UpdateStock(ctx context.Context, id uint, stock int) error {
	err := r.db.WithContext(ctx).Model(&ProductModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"stock":      stock,
			"updated_at": time.Now(),
		}).Error

	return r.handleError(err)
}

// UpdatePrice implements ports.ProductRepository (additional method for completeness)
func (r *GormProductRepository) UpdatePrice(ctx context.Context, id uint, price float64) error {
	err := r.db.WithContext(ctx).Model(&ProductModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"price":      price,
			"updated_at": time.Now(),
		}).Error

	return r.handleError(err)
}

// UpdateStatus implements ports.ProductRepository (additional method for completeness)
func (r *GormProductRepository) UpdateStatus(ctx context.Context, id uint, status entities.ProductStatus) error {
	err := r.db.WithContext(ctx).Model(&ProductModel{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"status":     string(status),
			"updated_at": time.Now(),
		}).Error

	return r.handleError(err)
}

// GetAvailableProducts implements ports.ProductRepository (additional method for completeness)
func (r *GormProductRepository) GetAvailableProducts(ctx context.Context, limit, offset int) ([]*entities.Product, error) {
	var models []ProductModel

	err := r.db.WithContext(ctx).Model(&ProductModel{}).
		Where("status = ? AND stock > 0", string(entities.ProductStatusActive)).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&models).Error

	if err != nil {
		return nil, r.handleError(err)
	}

	return r.toEntities(models), nil
}

// Count implements ports.ProductRepository (additional method for completeness)
func (r *GormProductRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ProductModel{}).Count(&count).Error
	if err != nil {
		return 0, r.handleError(err)
	}
	return count, nil
}

// CountByCategory implements ports.ProductRepository (additional method for completeness)
func (r *GormProductRepository) CountByCategory(ctx context.Context, category string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ProductModel{}).
		Where("category = ?", category).
		Count(&count).Error
	if err != nil {
		return 0, r.handleError(err)
	}
	return count, nil
}

// CountByStatus implements ports.ProductRepository (additional method for completeness)
func (r *GormProductRepository) CountByStatus(ctx context.Context, status entities.ProductStatus) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&ProductModel{}).
		Where("status = ?", string(status)).
		Count(&count).Error
	if err != nil {
		return 0, r.handleError(err)
	}
	return count, nil
}

// Helper functions for conversion between domain entities and GORM models

func (r *GormProductRepository) toModel(product *entities.Product) *ProductModel {
	return &ProductModel{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		SKU:         product.SKU,
		Price:       product.Price,
		Category:    product.Category,
		Brand:       product.Brand,
		Stock:       product.Stock,
		Status:      string(product.Status),
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
}

func (r *GormProductRepository) toEntity(model *ProductModel) *entities.Product {
	return &entities.Product{
		ID:          model.ID,
		Name:        model.Name,
		Description: model.Description,
		SKU:         model.SKU,
		Price:       model.Price,
		Category:    model.Category,
		Brand:       model.Brand,
		Stock:       model.Stock,
		Status:      entities.ProductStatus(model.Status),
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func (r *GormProductRepository) toEntities(models []ProductModel) []*entities.Product {
	products := make([]*entities.Product, 0, len(models))
	for _, model := range models {
		products = append(products, r.toEntity(&model))
	}
	return products
}

// Helper to convert GORM errors to domain errors
func (r *GormProductRepository) handleError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domainErrors.ErrProductNotFound
	}

	// Handle unique constraint violation for SKU
	if errors.Is(err, gorm.ErrDuplicatedKey) ||
		(err.Error() != "" && (strings.Contains(err.Error(), "duplicate key") ||
			strings.Contains(err.Error(), "UNIQUE constraint") ||
			strings.Contains(err.Error(), "sku"))) {
		return domainErrors.ErrProductAlreadyExists
	}

	// Return original error for other cases (can be enhanced with more specific error mapping)
	return err
}
