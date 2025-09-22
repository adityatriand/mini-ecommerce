package order

import "time"

type OrderStatus string

const (
	StatusPending   OrderStatus = "PENDING"
	StatusPaid      OrderStatus = "PAID"
	StatusCancelled OrderStatus = "CANCELLED"
)

type Order struct {
	ID         uint        `gorm:"primaryKey" json:"id"`
	UserID     uint        `gorm:"not null" json:"user_id"`
	ProductID  uint        `gorm:"not null" json:"product_id"`
	Quantity   int         `gorm:"not null" json:"quantity"`
	TotalPrice int         `gorm:"not null" json:"total_price"`
	Status     OrderStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type CreateOrderInput struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0"`
}

type UpdateOrderInput struct {
	Quantity *int         `json:"quantity"`
	Status   *OrderStatus `json:"status"`
}
