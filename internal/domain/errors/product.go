package errors

import "fmt"

type DomainError struct {
	Code    string
	Message string
	Field   string
}

func (e *DomainError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s (field: %s)", e.Code, e.Message, e.Field)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Product-specific domain errors
var (
	ErrProductNotFound = &DomainError{
		Code:    "PRODUCT_NOT_FOUND",
		Message: "Product not found",
	}

	ErrProductAlreadyExists = &DomainError{
		Code:    "PRODUCT_ALREADY_EXISTS",
		Message: "Product with this SKU already exists",
		Field:   "sku",
	}

	ErrInvalidProductName = &DomainError{
		Code:    "INVALID_PRODUCT_NAME",
		Message: "Invalid product name",
		Field:   "name",
	}

	ErrInvalidProductSKU = &DomainError{
		Code:    "INVALID_SKU",
		Message: "Invalid SKU format",
		Field:   "sku",
	}

	ErrInvalidProductPrice = &DomainError{
		Code:    "INVALID_PRICE",
		Message: "Invalid product price",
		Field:   "price",
	}

	ErrInvalidProductStock = &DomainError{
		Code:    "INVALID_STOCK",
		Message: "Invalid stock quantity",
		Field:   "stock",
	}

	ErrInvalidProductCategory = &DomainError{
		Code:    "INVALID_CATEGORY",
		Message: "Invalid product category",
		Field:   "category",
	}

	ErrProductInactive = &DomainError{
		Code:    "PRODUCT_INACTIVE",
		Message: "Product is inactive",
	}

	ErrProductDiscontinued = &DomainError{
		Code:    "PRODUCT_DISCONTINUED",
		Message: "Product has been discontinued",
	}

	ErrProductOutOfStock = &DomainError{
		Code:    "PRODUCT_OUT_OF_STOCK",
		Message: "Product is out of stock",
	}

	ErrInsufficientStock = &DomainError{
		Code:    "INSUFFICIENT_STOCK",
		Message: "Insufficient stock quantity",
		Field:   "stock",
	}

	ErrProductNotAvailable = &DomainError{
		Code:    "PRODUCT_NOT_AVAILABLE",
		Message: "Product is not available for purchase",
	}

	ErrFailedToCheckProductExistance = &DomainError{
		Code:    "FAILED_TO_CHECK_PRODUCT_EXISTENCE",
		Message: "failed to check product existence",
	}

	ErrFailedToCreateProduct = &DomainError{
		Code:    "FAILED_TO_CREATE_PRODUCT",
		Message: "failed to create product",
	}

	ErrFailedToUpdateProduct = &DomainError{
		Code:    "FAILED_TO_UPDATE_PRODUCT",
		Message: "failed to update product",
	}

	ErrFailedToDeleteProduct = &DomainError{
		Code:    "FAILED_TO_DELETE_PRODUCT",
		Message: "failed to delete product",
	}

	ErrFailedToListProducts = &DomainError{
		Code:    "FAILED_TO_LIST_PRODUCTS",
		Message: "failed to list products",
	}

	ErrFailedToSearchProducts = &DomainError{
		Code:    "FAILED_TO_SEARCH_PRODUCTS",
		Message: "failed to search products",
	}

	ErrFailedToUpdateStock = &DomainError{
		Code:    "FAILED_TO_UPDATE_STOCK",
		Message: "failed to update product stock",
	}

	ErrFailedToUpdatePrice = &DomainError{
		Code:    "FAILED_TO_UPDATE_PRICE",
		Message: "failed to update product price",
	}
)

func NewProductValidationError(field, message string) *DomainError {
	return &DomainError{
		Code:    "VALIDATION_ERROR",
		Message: message,
		Field:   field,
	}
}

func NewProductBusinessRuleError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

func NewProductOperationError(operation, message string) *DomainError {
	return &DomainError{
		Code:    fmt.Sprintf("FAILED_TO_%s_PRODUCT", operation),
		Message: fmt.Sprintf("failed to %s product: %s", operation, message),
	}
}
