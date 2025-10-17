package model

import (
	"errors"
	"time"
)

// Error definitions
var (
	ErrProductNotFound      = errors.New("product not found")
	ErrSKUAlreadyExists     = errors.New("SKU already exists")
	ErrInsufficientStock    = errors.New("insufficient stock")
	ErrProductServiceUnavailable = errors.New("product service unavailable")
)

// Product represents a product in the system
type Product struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"not null" json:"name"`
	Description string    `gorm:"type:text" json:"description,omitempty"`
	Price       int       `gorm:"not null" json:"price"` // Price in cents
	Stock       int       `gorm:"not null;default:0" json:"stock"`
	Category    string    `gorm:"size:100" json:"category,omitempty"`
	SKU         string    `gorm:"uniqueIndex;size:100" json:"sku,omitempty"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateProductRequest represents product creation request
type CreateProductRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty"`
	Price       int    `json:"price" validate:"required,min=1"`
	Stock       int    `json:"stock" validate:"min=0"`
	Category    string `json:"category,omitempty"`
	SKU         string `json:"sku,omitempty"`
}

// UpdateProductRequest represents product update request
type UpdateProductRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Description *string `json:"description,omitempty"`
	Price       *int    `json:"price,omitempty" validate:"omitempty,min=1"`
	Stock       *int    `json:"stock,omitempty" validate:"omitempty,min=0"`
	Category    *string `json:"category,omitempty"`
	SKU         *string `json:"sku,omitempty"`
	IsActive    *bool   `json:"is_active,omitempty"`
}

// ProductQuery represents product search/filter query
type ProductQuery struct {
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortBy   string `form:"sort_by" binding:"omitempty,oneof=id name price stock created_at"`
	Order    string `form:"order" binding:"omitempty,oneof=asc desc"`
	Category string `form:"category" binding:"omitempty"`
	Search   string `form:"search" binding:"omitempty"`
	IsActive *bool  `form:"is_active" binding:"omitempty"`
}

// ProductListResponse represents paginated product list response
type ProductListResponse struct {
	Data       []Product           `json:"data"`
	Pagination PaginationMetadata  `json:"pagination"`
}

// PaginationMetadata represents pagination information
type PaginationMetadata struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// UpdateStockRequest represents stock update request
type UpdateStockRequest struct {
	Stock int `json:"stock" validate:"required"`
}

// BulkProductRequest represents bulk product operations
type BulkProductRequest struct {
	ProductIDs []uint `json:"product_ids" validate:"required,min=1,max=100"`
}

// ProductInfo represents product information for order service
type ProductInfo struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Stock int    `json:"stock"`
}
