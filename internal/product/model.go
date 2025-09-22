package product

import "time"

type Product struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"not null" json:"name"`
	Price     int       `gorm:"not null" json:"price"`
	Stock     int       `gorm:"not null;default:0" json:"stock"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateProductInput struct {
	Name  string `json:"name" binding:"required"`
	Price int    `json:"price" binding:"required"`
	Stock int    `json:"stock" binding:"required"`
}

type UpdateProductInput struct {
	Name  *string `json:"name"`
	Price *int    `json:"price"`
	Stock *int    `json:"stock"`
}
