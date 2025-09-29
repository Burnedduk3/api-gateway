package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"api-gateway/internal/application/dto"
	"api-gateway/internal/application/usecases"
	domainErrors "api-gateway/internal/domain/errors"
	"api-gateway/pkg/logger"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
)

type ProductHandler struct {
	productUseCases usecases.ProductUseCases
	validator       *validator.Validate
	logger          logger.Logger
}

func NewProductHandler(productUseCases usecases.ProductUseCases, log logger.Logger) *ProductHandler {
	return &ProductHandler{
		productUseCases: productUseCases,
		validator:       validator.New(),
		logger:          log.With("component", "product_handler"),
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// CreateProduct handles POST /api/v1/products
func (h *ProductHandler) CreateProduct(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Info("Create product request received",
		"request_id", requestID,
		"remote_ip", c.RealIP(),
		"user_agent", c.Request().UserAgent())

	// Parse request body
	var request dto.CreateProductRequestDTO
	if err := c.Bind(&request); err != nil {
		h.logger.Warn("Failed to bind request body",
			"request_id", requestID,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request body format",
		})
	}

	// Validate request
	if err := h.validator.Struct(request); err != nil {
		h.logger.Warn("Request validation failed",
			"request_id", requestID,
			"error", err)

		details := make(map[string]interface{})
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				details[fieldError.Field()] = getValidationErrorMessage(fieldError)
			}
		}

		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Request validation failed",
			Details: details,
		})
	}

	// Execute use case
	response, err := h.productUseCases.CreateProduct(c.Request().Context(), &request)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to create product")
	}

	h.logger.Info("Product created successfully",
		"request_id", requestID,
		"product_id", response.ID,
		"sku", response.SKU)

	return c.JSON(http.StatusCreated, response)
}

// GetProduct handles GET /api/v1/products/:id
func (h *ProductHandler) GetProduct(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	// Parse product ID from path parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid product ID parameter",
			"request_id", requestID,
			"id_param", idParam,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid product ID format",
		})
	}

	h.logger.Info("Get product request received",
		"request_id", requestID,
		"product_id", id,
		"remote_ip", c.RealIP())

	// Execute use case
	response, err := h.productUseCases.GetProductByID(c.Request().Context(), uint(id))
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to get product")
	}

	h.logger.Info("Product retrieved successfully",
		"request_id", requestID,
		"product_id", response.ID)

	return c.JSON(http.StatusOK, response)
}

// GetProductBySKU handles GET /api/v1/products/sku/:sku
func (h *ProductHandler) GetProductBySKU(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	sku := c.Param("sku")
	if sku == "" {
		h.logger.Warn("Empty SKU parameter",
			"request_id", requestID)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_SKU",
			Message: "SKU parameter is required",
		})
	}

	h.logger.Info("Get product by SKU request received",
		"request_id", requestID,
		"sku", sku,
		"remote_ip", c.RealIP())

	// Execute use case
	response, err := h.productUseCases.GetProductBySKU(c.Request().Context(), sku)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to get product by SKU")
	}

	h.logger.Info("Product retrieved by SKU successfully",
		"request_id", requestID,
		"product_id", response.ID,
		"sku", response.SKU)

	return c.JSON(http.StatusOK, response)
}

// UpdateProduct handles PUT /api/v1/products/:id
func (h *ProductHandler) UpdateProduct(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	// Parse product ID from path parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid product ID parameter",
			"request_id", requestID,
			"id_param", idParam,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid product ID format",
		})
	}

	// Parse request body
	var request dto.UpdateProductRequestDTO
	if err := c.Bind(&request); err != nil {
		h.logger.Warn("Failed to bind request body",
			"request_id", requestID,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request body format",
		})
	}

	// Validate request
	if err := h.validator.Struct(request); err != nil {
		h.logger.Warn("Request validation failed",
			"request_id", requestID,
			"error", err)

		details := make(map[string]interface{})
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			for _, fieldError := range validationErrors {
				details[fieldError.Field()] = getValidationErrorMessage(fieldError)
			}
		}

		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Request validation failed",
			Details: details,
		})
	}

	h.logger.Info("Update product request received",
		"request_id", requestID,
		"product_id", id,
		"remote_ip", c.RealIP())

	// Execute use case
	response, err := h.productUseCases.UpdateProduct(c.Request().Context(), uint(id), &request)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to update product")
	}

	h.logger.Info("Product updated successfully",
		"request_id", requestID,
		"product_id", response.ID)

	return c.JSON(http.StatusOK, response)
}

// UpdateProductStock handles PATCH /api/v1/products/:id/stock
func (h *ProductHandler) UpdateProductStock(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	// Parse product ID from path parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid product ID parameter",
			"request_id", requestID,
			"id_param", idParam,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid product ID format",
		})
	}

	// Parse request body
	var request dto.StockUpdateRequestDTO
	if err := c.Bind(&request); err != nil {
		h.logger.Warn("Failed to bind request body",
			"request_id", requestID,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request body format",
		})
	}

	// Validate request
	if err := h.validator.Struct(request); err != nil {
		h.logger.Warn("Request validation failed",
			"request_id", requestID,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Request validation failed",
		})
	}

	h.logger.Info("Update product stock request received",
		"request_id", requestID,
		"product_id", id,
		"new_stock", request.Stock)

	// Execute use case
	response, err := h.productUseCases.UpdateProductStock(c.Request().Context(), uint(id), request.Stock)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to update product stock")
	}

	h.logger.Info("Product stock updated successfully",
		"request_id", requestID,
		"product_id", response.ID,
		"new_stock", response.Stock)

	return c.JSON(http.StatusOK, response)
}

// UpdateProductPrice handles PATCH /api/v1/products/:id/price
func (h *ProductHandler) UpdateProductPrice(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	// Parse product ID from path parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid product ID parameter",
			"request_id", requestID,
			"id_param", idParam,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid product ID format",
		})
	}

	// Parse request body
	var request dto.PriceUpdateRequestDTO
	if err := c.Bind(&request); err != nil {
		h.logger.Warn("Failed to bind request body",
			"request_id", requestID,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_REQUEST",
			Message: "Invalid request body format",
		})
	}

	// Validate request
	if err := h.validator.Struct(request); err != nil {
		h.logger.Warn("Request validation failed",
			"request_id", requestID,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "VALIDATION_ERROR",
			Message: "Request validation failed",
		})
	}

	h.logger.Info("Update product price request received",
		"request_id", requestID,
		"product_id", id,
		"new_price", request.Price)

	// Execute use case
	response, err := h.productUseCases.UpdateProductPrice(c.Request().Context(), uint(id), request.Price)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to update product price")
	}

	h.logger.Info("Product price updated successfully",
		"request_id", requestID,
		"product_id", response.ID,
		"new_price", response.Price)

	return c.JSON(http.StatusOK, response)
}

// ActivateProduct handles PATCH /api/v1/products/:id/activate
func (h *ProductHandler) ActivateProduct(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	// Parse product ID from path parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid product ID parameter",
			"request_id", requestID,
			"id_param", idParam,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid product ID format",
		})
	}

	h.logger.Info("Activate product request received",
		"request_id", requestID,
		"product_id", id)

	// Execute use case
	response, err := h.productUseCases.ActivateProduct(c.Request().Context(), uint(id))
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to activate product")
	}

	h.logger.Info("Product activated successfully",
		"request_id", requestID,
		"product_id", response.ID)

	return c.JSON(http.StatusOK, response)
}

// DeactivateProduct handles PATCH /api/v1/products/:id/deactivate
func (h *ProductHandler) DeactivateProduct(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	// Parse product ID from path parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid product ID parameter",
			"request_id", requestID,
			"id_param", idParam,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid product ID format",
		})
	}

	h.logger.Info("Deactivate product request received",
		"request_id", requestID,
		"product_id", id)

	// Execute use case
	response, err := h.productUseCases.DeactivateProduct(c.Request().Context(), uint(id))
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to deactivate product")
	}

	h.logger.Info("Product deactivated successfully",
		"request_id", requestID,
		"product_id", response.ID)

	return c.JSON(http.StatusOK, response)
}

// DiscontinueProduct handles PATCH /api/v1/products/:id/discontinue
func (h *ProductHandler) DiscontinueProduct(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	// Parse product ID from path parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		h.logger.Warn("Invalid product ID parameter",
			"request_id", requestID,
			"id_param", idParam,
			"error", err)
		return c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "INVALID_ID",
			Message: "Invalid product ID format",
		})
	}

	h.logger.Info("Discontinue product request received",
		"request_id", requestID,
		"product_id", id)

	// Execute use case
	response, err := h.productUseCases.DiscontinueProduct(c.Request().Context(), uint(id))
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to discontinue product")
	}

	h.logger.Info("Product discontinued successfully",
		"request_id", requestID,
		"product_id", response.ID)

	return c.JSON(http.StatusOK, response)
}

// ListProducts handles GET /api/v1/products
func (h *ProductHandler) ListProducts(c echo.Context) error {
	requestID := c.Response().Header().Get(echo.HeaderXRequestID)

	h.logger.Info("List products request received",
		"request_id", requestID,
		"remote_ip", c.RealIP())

	// Parse query parameters
	page := 0
	pageSize := 10

	if pageParam := c.QueryParam("page"); pageParam != "" {
		if p, err := strconv.Atoi(pageParam); err == nil && p >= 0 {
			page = p
		}
	}

	if sizeParam := c.QueryParam("page_size"); sizeParam != "" {
		if ps, err := strconv.Atoi(sizeParam); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	}

	h.logger.Info("List products parameters",
		"request_id", requestID,
		"page", page,
		"page_size", pageSize)

	// Execute use case
	response, err := h.productUseCases.ListProducts(c.Request().Context(), page, pageSize)
	if err != nil {
		return h.handleError(c, err, requestID, "Failed to list products")
	}

	h.logger.Info("Products listed successfully",
		"request_id", requestID,
		"count", len(response.Products),
		"page", page)

	return c.JSON(http.StatusOK, response)
}

// handleError handles different types of errors and returns appropriate HTTP responses
func (h *ProductHandler) handleError(c echo.Context, err error, requestID, logMessage string) error {
	h.logger.Error(logMessage,
		"request_id", requestID,
		"error", err)

	// Handle domain errors
	var domainErr *domainErrors.DomainError
	if errors.As(err, &domainErr) {
		switch domainErr.Code {
		case domainErrors.ErrProductNotFound.Code:
			return c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   domainErr.Code,
				Message: domainErr.Message,
			})
		case domainErrors.ErrProductAlreadyExists.Code:
			return c.JSON(http.StatusConflict, ErrorResponse{
				Error:   domainErr.Code,
				Message: domainErr.Message,
			})
		case domainErrors.ErrInvalidProductName.Code,
			domainErrors.ErrInvalidProductSKU.Code,
			domainErrors.ErrInvalidProductPrice.Code,
			domainErrors.ErrInvalidProductStock.Code,
			domainErrors.ErrInvalidProductCategory.Code,
			domainErrors.ErrInsufficientStock.Code:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   domainErr.Code,
				Message: domainErr.Message,
			})
		case domainErrors.ErrProductInactive.Code,
			domainErrors.ErrProductDiscontinued.Code,
			domainErrors.ErrProductOutOfStock.Code,
			domainErrors.ErrProductNotAvailable.Code:
			return c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
				Error:   domainErr.Code,
				Message: domainErr.Message,
			})
		default:
			return c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   domainErr.Code,
				Message: domainErr.Message,
			})
		}
	}

	// Handle generic errors
	return c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error:   "INTERNAL_ERROR",
		Message: "An internal error occurred",
	})
}

// getValidationErrorMessage returns a user-friendly validation error message
func getValidationErrorMessage(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "This field is required"
	case "min":
		return "Minimum value is " + fieldError.Param()
	case "max":
		return "Maximum value is " + fieldError.Param()
	case "gte":
		return "Value must be greater than or equal to " + fieldError.Param()
	case "lte":
		return "Value must be less than or equal to " + fieldError.Param()
	default:
		return "Invalid value"
	}
}
