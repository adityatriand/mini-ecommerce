package order

type CreateOrderRequest struct {
	ProductID uint `json:"product_id" binding:"required" validate:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0" validate:"required,gt=0"`
}

type UpdateOrderRequest struct {
	Quantity *int         `json:"quantity" validate:"omitempty,gt=0"`
	Status   *OrderStatus `json:"status" validate:"omitempty,oneof=pending paid cancelled"`
}
