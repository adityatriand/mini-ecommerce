package model

import (
	"errors"
	"time"
)

// Error definitions
var (
	ErrOrderNotFound                    = errors.New("order not found")
	ErrNotAuthorizedToUpdate            = errors.New("not authorized to update this order")
	ErrInvalidStatusValue               = errors.New("invalid status value")
	ErrCannotChangePaidOrderToPending   = errors.New("cannot change paid order to pending")
	ErrCannotChangeCancelledOrderStatus = errors.New("cannot change cancelled order status")
	ErrProductNotFound                  = errors.New("product not found")
	ErrInsufficientStock                = errors.New("insufficient stock")
	ErrProductServiceUnavailable        = errors.New("product service unavailable")
)

// OrderStatus represents the status of an order
type OrderStatus string

const (
	StatusPending   OrderStatus = "PENDING"
	StatusPaid      OrderStatus = "PAID"
	StatusShipped   OrderStatus = "SHIPPED"
	StatusDelivered OrderStatus = "DELIVERED"
	StatusCancelled OrderStatus = "CANCELLED"
)

// Order represents an order in the system
type Order struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	UserID     uint        `gorm:"not null" json:"user_id"`
	TotalPrice int         `gorm:"not null" json:"total_price"` // Price in cents
	Status     OrderStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	OrderItems []OrderItem `gorm:"foreignKey:OrderID" json:"order_items,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// OrderItem represents an item within an order
type OrderItem struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	OrderID    uint      `gorm:"not null;index" json:"order_id"`
	ProductID  uint      `gorm:"not null" json:"product_id"`
	Quantity   int       `gorm:"not null" json:"quantity"`
	Price      int       `gorm:"not null" json:"price"` // Price in cents
	Subtotal   int       `gorm:"not null" json:"subtotal"` // Subtotal in cents
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// CreateOrderRequest represents order creation request
type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items" validate:"required,min=1,dive"`
}

// OrderItemRequest represents an order item in creation request
type OrderItemRequest struct {
	ProductID uint `json:"product_id" validate:"required"`
	Quantity  int  `json:"quantity" validate:"required,min=1"`
}

// UpdateOrderRequest represents order update request
type UpdateOrderRequest struct {
	Status *OrderStatus `json:"status,omitempty" validate:"omitempty,oneof=PENDING PAID SHIPPED DELIVERED CANCELLED"`
}

// OrderQuery represents order search/filter query
type OrderQuery struct {
	Page     int         `form:"page" binding:"omitempty,min=1"`
	PageSize int         `form:"page_size" binding:"omitempty,min=1,max=100"`
	SortBy   string      `form:"sort_by" binding:"omitempty,oneof=id user_id total_price status created_at"`
	Order    string      `form:"order" binding:"omitempty,oneof=asc desc"`
	Status   OrderStatus `form:"status" binding:"omitempty,oneof=PENDING PAID SHIPPED DELIVERED CANCELLED"`
	UserID   *uint       `form:"user_id" binding:"omitempty"`
}

// OrderListResponse represents paginated order list response
type OrderListResponse struct {
	Data       []Order            `json:"data"`
	Pagination PaginationMetadata `json:"pagination"`
}

// PaginationMetadata represents pagination information
type PaginationMetadata struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ProductInfo represents product information for order items
type ProductInfo struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Stock int    `json:"stock"`
}

// OrderWithProductInfo represents order with product information
type OrderWithProductInfo struct {
	ID         uint        `json:"id"`
	UserID     uint        `json:"user_id"`
	TotalPrice int         `json:"total_price"`
	Status     OrderStatus `json:"status"`
	OrderItems []OrderItemWithProduct `json:"order_items,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

// OrderItemWithProduct represents order item with product information
type OrderItemWithProduct struct {
	ID        uint        `json:"id"`
	OrderID   uint        `json:"order_id"`
	ProductID uint        `json:"product_id"`
	Quantity  int         `json:"quantity"`
	Price     int         `json:"price"`
	Subtotal  int         `json:"subtotal"`
	Product   ProductInfo `json:"product"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}
