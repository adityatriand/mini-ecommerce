package order

import "mini-e-commerce/internal/dto"

type OrderQuery struct {
	dto.PaginationQuery
	SortBy string `form:"sort_by" binding:"omitempty,oneof=id user_id product_id quantity total_price status created_at"`
}

type OrderItemInput struct {
	ProductID uint `json:"product_id" binding:"required" validate:"required"`
	Quantity  int  `json:"quantity" binding:"required,gt=0" validate:"required,gt=0"`
}

type CreateOrderRequest struct {
	Items []OrderItemInput `json:"items" binding:"required,min=1,dive" validate:"required,min=1,dive"`
}

type UpdateOrderRequest struct {
	Status *OrderStatus `json:"status" validate:"omitempty,oneof=PENDING PAID CANCELLED"`
}

type OrderListResponse struct {
	Data       []Order                `json:"data"`
	Pagination dto.PaginationMetadata `json:"pagination"`
}
