package entities

import (
	"errors"
	"strings"
	"time"
)

type ProductStatus string

const (
	ProductStatusActive       ProductStatus = "active"
	ProductStatusInactive     ProductStatus = "inactive"
	ProductStatusDiscontinued ProductStatus = "discontinued"
)

type Product struct {
	ID          uint          `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	SKU         string        `json:"sku"`
	Price       float64       `json:"price"`
	Category    string        `json:"category"`
	Brand       string        `json:"brand"`
	Stock       int           `json:"stock"`
	Status      ProductStatus `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`
}

func (p *Product) IsActive() bool {
	return p.Status == ProductStatusActive
}

func (p *Product) IsInStock() bool {
	return p.Stock > 0
}

func (p *Product) IsAvailable() bool {
	return p.IsActive() && p.IsInStock()
}

func (p *Product) Activate() {
	p.Status = ProductStatusActive
	p.UpdatedAt = time.Now()
}

func (p *Product) Deactivate() {
	p.Status = ProductStatusInactive
	p.UpdatedAt = time.Now()
}

func (p *Product) Discontinue() {
	p.Status = ProductStatusDiscontinued
	p.UpdatedAt = time.Now()
}

func (p *Product) UpdateStock(quantity int) error {
	if quantity < 0 {
		return errors.New("stock quantity cannot be negative")
	}
	p.Stock = quantity
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Product) ReduceStock(quantity int) error {
	if quantity <= 0 {
		return errors.New("reduction quantity must be positive")
	}
	if p.Stock < quantity {
		return errors.New("insufficient stock")
	}
	p.Stock -= quantity
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Product) AddStock(quantity int) error {
	if quantity <= 0 {
		return errors.New("addition quantity must be positive")
	}
	p.Stock += quantity
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Product) UpdatePrice(price float64) error {
	if price < 0 {
		return errors.New("price cannot be negative")
	}
	p.Price = price
	p.UpdatedAt = time.Now()
	return nil
}

func NewProduct(name, description, sku, category, brand string, price float64, stock int) (*Product, error) {
	if err := validateProductName(name); err != nil {
		return nil, err
	}

	if err := validateSKU(sku); err != nil {
		return nil, err
	}

	if err := validatePrice(price); err != nil {
		return nil, err
	}

	if err := validateStock(stock); err != nil {
		return nil, err
	}

	if strings.TrimSpace(category) == "" {
		return nil, errors.New("category is required")
	}

	now := time.Now()

	return &Product{
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		SKU:         strings.ToUpper(strings.TrimSpace(sku)),
		Price:       price,
		Category:    strings.TrimSpace(category),
		Brand:       strings.TrimSpace(brand),
		Stock:       stock,
		Status:      ProductStatusActive,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func validateProductName(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		return errors.New("product name is required")
	}
	if len(name) < 2 {
		return errors.New("product name must be at least 2 characters long")
	}
	if len(name) > 255 {
		return errors.New("product name must be less than 255 characters")
	}
	return nil
}

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

func validatePrice(price float64) error {
	if price < 0 {
		return errors.New("price cannot be negative")
	}
	if price > 999999.99 {
		return errors.New("price cannot exceed 999,999.99")
	}
	return nil
}

func validateStock(stock int) error {
	if stock < 0 {
		return errors.New("stock cannot be negative")
	}
	return nil
}
