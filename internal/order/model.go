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
	TotalPrice int         `gorm:"not null" json:"total_price"`
	Status     OrderStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	OrderItems []OrderItem `gorm:"foreignKey:OrderID" json:"order_items,omitempty"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
}

type OrderItem struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	OrderID    uint      `gorm:"not null;index" json:"order_id"`
	ProductID  uint      `gorm:"not null" json:"product_id"`
	Quantity   int       `gorm:"not null" json:"quantity"`
	Price      int       `gorm:"not null" json:"price"`
	Subtotal   int       `gorm:"not null" json:"subtotal"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}
