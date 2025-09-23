package product

type CreateProductRequest struct {
	Name  string `json:"name" binding:"required" validate:"required"`
	Price int    `json:"price" binding:"required" validate:"required,gt=0"`
	Stock int    `json:"stock" binding:"required" validate:"gte=0"`
}

type UpdateProductRequest struct {
	Name  *string `json:"name" validate:"omitempty"`
	Price *int    `json:"price" validate:"omitempty,gt=0"`
	Stock *int    `json:"stock" validate:"omitempty,gte=0"`
}
