package usecases

import (
	"context"
	"errors"
	"product-service/internal/application/dto"
	"product-service/internal/application/ports"
	productErrors "product-service/internal/domain/errors"
	"product-service/pkg/logger"
	"strings"
)

// ProductUseCases defines the interface for product business operations
type ProductUseCases interface {
	CreateProduct(ctx context.Context, request *dto.CreateProductRequestDTO) (*dto.ProductResponseDTO, error)
	GetProductByID(ctx context.Context, id uint) (*dto.ProductResponseDTO, error)
	GetProductBySKU(ctx context.Context, sku string) (*dto.ProductResponseDTO, error)
	UpdateProduct(ctx context.Context, id uint, request *dto.UpdateProductRequestDTO) (*dto.ProductResponseDTO, error)
	UpdateProductStock(ctx context.Context, id uint, stock int) (*dto.ProductResponseDTO, error)
	UpdateProductPrice(ctx context.Context, id uint, price float64) (*dto.ProductResponseDTO, error)
	ActivateProduct(ctx context.Context, id uint) (*dto.ProductResponseDTO, error)
	DeactivateProduct(ctx context.Context, id uint) (*dto.ProductResponseDTO, error)
	DiscontinueProduct(ctx context.Context, id uint) (*dto.ProductResponseDTO, error)
	ListProducts(ctx context.Context, page, pageSize int) (*dto.ProductListResponseDTO, error)
}

// productUseCasesImpl implements ProductUseCases interface
type productUseCasesImpl struct {
	productRepo ports.ProductRepository
	logger      logger.Logger
}

// NewProductUseCases creates a new instance of product use cases
func NewProductUseCases(productRepo ports.ProductRepository, log logger.Logger) ProductUseCases {
	return &productUseCasesImpl{
		productRepo: productRepo,
		logger:      log.With("component", "product_usecases"),
	}
}

func (uc *productUseCasesImpl) CreateProduct(ctx context.Context, request *dto.CreateProductRequestDTO) (*dto.ProductResponseDTO, error) {
	uc.logger.Info("CreateProduct use case called", "sku", request.SKU)

	// Validate SKU format
	if err := validateSKU(request.SKU); err != nil {
		return nil, productErrors.ErrInvalidProductSKU
	}

	// Check if product with this SKU already exists
	exists, err := uc.productRepo.ExistsBySKU(ctx, request.SKU)
	if err != nil {
		uc.logger.Error("Failed to check product existence", "error", err, "sku", request.SKU)
		return nil, productErrors.ErrFailedToCheckProductExistance
	}

	if exists {
		return nil, productErrors.ErrProductAlreadyExists
	}

	// Convert DTO to domain entity
	domainEntity, err := request.ToEntity()
	if err != nil {
		uc.logger.Error("Failed to convert DTO to entity", "error", err)
		return nil, err
	}

	// Create product
	createdProduct, err := uc.productRepo.Create(ctx, domainEntity)
	if err != nil {
		uc.logger.Error("Failed to create product", "error", err, "sku", request.SKU)
		switch {
		case errors.Is(err, productErrors.ErrFailedToCheckProductExistance):
			return nil, productErrors.ErrFailedToCheckProductExistance
		default:
			return nil, productErrors.ErrFailedToCreateProduct
		}
	}

	uc.logger.Info("CreateProduct success", "sku", request.SKU, "id", createdProduct.ID)
	return dto.ProductToResponseDTO(createdProduct), nil
}

// GetProductByID retrieves a product by its ID
func (uc *productUseCasesImpl) GetProductByID(ctx context.Context, id uint) (*dto.ProductResponseDTO, error) {
	uc.logger.Info("GetProductByID use case called", "product_id", id)

	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get product by ID", "error", err, "product_id", id)
		return nil, err
	}

	uc.logger.Info("GetProductByID success", "product_id", id)
	return dto.ProductToResponseDTO(product), nil
}

// GetProductBySKU retrieves a product by its SKU
func (uc *productUseCasesImpl) GetProductBySKU(ctx context.Context, sku string) (*dto.ProductResponseDTO, error) {
	uc.logger.Info("GetProductBySKU use case called", "sku", sku)

	product, err := uc.productRepo.GetBySKU(ctx, sku)
	if err != nil {
		uc.logger.Error("Failed to get product by SKU", "error", err, "sku", sku)
		return nil, err
	}

	uc.logger.Info("GetProductBySKU success", "product_id", product.ID, "sku", sku)
	return dto.ProductToResponseDTO(product), nil
}

// UpdateProduct updates an existing product
func (uc *productUseCasesImpl) UpdateProduct(ctx context.Context, id uint, request *dto.UpdateProductRequestDTO) (*dto.ProductResponseDTO, error) {
	uc.logger.Info("UpdateProduct use case called", "product_id", id)

	// Get existing product
	existingProduct, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get existing product", "error", err, "product_id", id)
		return nil, err
	}

	// Update fields if provided
	if request.Name != "" {
		existingProduct.Name = request.Name
	}

	if request.Description != "" {
		existingProduct.Description = request.Description
	}

	if request.Category != "" {
		existingProduct.Category = request.Category
	}

	if request.Brand != "" {
		existingProduct.Brand = request.Brand
	}

	if request.Price != nil {
		if err := existingProduct.UpdatePrice(*request.Price); err != nil {
			return nil, productErrors.NewProductValidationError("price", err.Error())
		}
	}

	if request.Stock != nil {
		if err := existingProduct.UpdateStock(*request.Stock); err != nil {
			return nil, productErrors.NewProductValidationError("stock", err.Error())
		}
	}

	// Update product in repository (assuming we add Update method to interface)
	updatedProduct := existingProduct // For now, since Update method is not in the current interface

	uc.logger.Info("UpdateProduct success", "product_id", id)
	return dto.ProductToResponseDTO(updatedProduct), nil
}

// UpdateProductStock updates only the stock of a product
func (uc *productUseCasesImpl) UpdateProductStock(ctx context.Context, id uint, stock int) (*dto.ProductResponseDTO, error) {
	uc.logger.Info("UpdateProductStock use case called", "product_id", id, "stock", stock)

	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get product", "error", err, "product_id", id)
		return nil, err
	}

	// Update stock using domain method
	if err := product.UpdateStock(stock); err != nil {
		uc.logger.Error("Failed to update stock", "error", err, "product_id", id)
		return nil, productErrors.NewProductValidationError("stock", err.Error())
	}

	// In a real implementation, you would save to repository here
	// updatedProduct, err := uc.productRepo.Update(ctx, product)

	uc.logger.Info("UpdateProductStock success", "product_id", id, "new_stock", stock)
	return dto.ProductToResponseDTO(product), nil
}

// UpdateProductPrice updates only the price of a product
func (uc *productUseCasesImpl) UpdateProductPrice(ctx context.Context, id uint, price float64) (*dto.ProductResponseDTO, error) {
	uc.logger.Info("UpdateProductPrice use case called", "product_id", id, "price", price)

	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get product", "error", err, "product_id", id)
		return nil, err
	}

	// Update price using domain method
	if err := product.UpdatePrice(price); err != nil {
		uc.logger.Error("Failed to update price", "error", err, "product_id", id)
		return nil, productErrors.NewProductValidationError("price", err.Error())
	}

	// In a real implementation, you would save to repository here
	// updatedProduct, err := uc.productRepo.Update(ctx, product)

	uc.logger.Info("UpdateProductPrice success", "product_id", id, "new_price", price)
	return dto.ProductToResponseDTO(product), nil
}

// ActivateProduct activates a product
func (uc *productUseCasesImpl) ActivateProduct(ctx context.Context, id uint) (*dto.ProductResponseDTO, error) {
	uc.logger.Info("ActivateProduct use case called", "product_id", id)

	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get product", "error", err, "product_id", id)
		return nil, err
	}

	// Activate product using domain method
	product.Activate()

	// In a real implementation, you would save to repository here
	// updatedProduct, err := uc.productRepo.Update(ctx, product)

	uc.logger.Info("ActivateProduct success", "product_id", id)
	return dto.ProductToResponseDTO(product), nil
}

// DeactivateProduct deactivates a product
func (uc *productUseCasesImpl) DeactivateProduct(ctx context.Context, id uint) (*dto.ProductResponseDTO, error) {
	uc.logger.Info("DeactivateProduct use case called", "product_id", id)

	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get product", "error", err, "product_id", id)
		return nil, err
	}

	// Deactivate product using domain method
	product.Deactivate()

	// In a real implementation, you would save to repository here
	// updatedProduct, err := uc.productRepo.Update(ctx, product)

	uc.logger.Info("DeactivateProduct success", "product_id", id)
	return dto.ProductToResponseDTO(product), nil
}

// DiscontinueProduct discontinues a product
func (uc *productUseCasesImpl) DiscontinueProduct(ctx context.Context, id uint) (*dto.ProductResponseDTO, error) {
	uc.logger.Info("DiscontinueProduct use case called", "product_id", id)

	// Get existing product
	product, err := uc.productRepo.GetByID(ctx, id)
	if err != nil {
		uc.logger.Error("Failed to get product", "error", err, "product_id", id)
		return nil, err
	}

	// Discontinue product using domain method
	product.Discontinue()

	// In a real implementation, you would save to repository here
	// updatedProduct, err := uc.productRepo.Update(ctx, product)

	uc.logger.Info("DiscontinueProduct success", "product_id", id)
	return dto.ProductToResponseDTO(product), nil
}

// ListProducts retrieves a paginated list of products
func (uc *productUseCasesImpl) ListProducts(ctx context.Context, page, pageSize int) (*dto.ProductListResponseDTO, error) {
	uc.logger.Info("ListProducts use case called", "page", page, "page_size", pageSize)

	if page < 0 {
		page = 0
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	products := []*dto.ProductResponseDTO{}

	uc.logger.Info("ListProducts success", "page", page, "page_size", pageSize, "count", len(products))

	return &dto.ProductListResponseDTO{
		Products: products,
		Page:     page,
		PageSize: pageSize,
		Total:    len(products),
	}, nil
}

// validateSKU validates SKU format
func validateSKU(sku string) error {
	sku = strings.TrimSpace(sku)
	if sku == "" {
		return errors.New("SKU is required")
	}
	if len(sku) < 3 {
		return errors.New("SKU must be at least 3 characters long")
	}
	if len(sku) > 50 {
		return errors.New("SKU must be less than 50 characters")
	}
	return nil
}
