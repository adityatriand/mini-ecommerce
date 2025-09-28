package order

import "mini-e-commerce/internal/dto"

type OrderQuery struct {
	dto.PaginationQuery
	SortBy string `form:"sort_by" binding:"omitempty,oneof=id user_id product_id quantity total_price status created_at"`
}

type CreateOrderRequest struct {
	ProductID uint `json:"product_id" binding:"required" validate:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0" validate:"required,gt=0"`
}

type UpdateOrderRequest struct {
	Quantity *int         `json:"quantity" validate:"omitempty,gt=0"`
	Status   *OrderStatus `json:"status" validate:"omitempty,oneof=pending paid cancelled"`
}

type OrderListResponse struct {
	Data       []Order                `json:"data"`
	Pagination dto.PaginationMetadata `json:"pagination"`
}
