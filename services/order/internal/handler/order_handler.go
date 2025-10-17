package handler

import (
	"errors"
	"strconv"

	"mini-e-commerce/services/order/internal/model"
	"mini-e-commerce/services/order/internal/service"
	"mini-e-commerce/shared/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type OrderHandler struct {
	service        service.OrderService
	logger         *zap.Logger
	responseHelper *response.ResponseHelper
}

func NewOrderHandler(service service.OrderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		service:        service,
		logger:         logger,
		responseHelper: response.NewResponseHelper(logger),
	}
}

func (h *OrderHandler) RegisterRoutes(r *gin.RouterGroup) {
	orders := r.Group("/orders")
	{
		orders.POST("", h.CreateOrder)
		orders.GET("", h.GetOrders)
		orders.GET("/:id", h.GetOrderByID)
		orders.PUT("/:id", h.UpdateOrder)
		orders.DELETE("/:id", h.DeleteOrder)
	}
}

// CreateOrder handles order creation
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		h.responseHelper.Unauthorized(c, "User ID not found in context")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		h.responseHelper.Unauthorized(c, "Invalid user ID format")
		return
	}

	var req model.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHelper.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), req, userIDUint)
	if err != nil {
		if errors.Is(err, model.ErrProductNotFound) {
			h.responseHelper.NotFound(c, "Product not found")
		} else if errors.Is(err, model.ErrInsufficientStock) {
			h.responseHelper.BadRequest(c, "Insufficient stock", err.Error())
		} else if errors.Is(err, model.ErrProductServiceUnavailable) {
			h.responseHelper.ServiceUnavailable(c, "Product service unavailable")
		} else {
			h.responseHelper.BadRequest(c, "Failed to create order", err.Error())
		}
		return
	}

	h.responseHelper.Created(c, order, "Order created successfully")
}

// GetOrderByID handles getting order by ID
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		h.responseHelper.Unauthorized(c, "User ID not found in context")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		h.responseHelper.Unauthorized(c, "Invalid user ID format")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.responseHelper.BadRequest(c, "Invalid order ID", err.Error())
		return
	}

	order, err := h.service.GetOrderByID(c.Request.Context(), uint(id), userIDUint)
	if err != nil {
		if errors.Is(err, model.ErrOrderNotFound) {
			h.responseHelper.NotFound(c, "Order not found")
		} else if errors.Is(err, model.ErrNotAuthorizedToUpdate) {
			h.responseHelper.Forbidden(c, "Not authorized to view this order")
		} else {
			h.responseHelper.InternalError(c, "Failed to get order", err.Error())
		}
		return
	}

	h.responseHelper.Success(c, order)
}

// GetOrders handles getting orders with filtering and pagination
func (h *OrderHandler) GetOrders(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		h.responseHelper.Unauthorized(c, "User ID not found in context")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		h.responseHelper.Unauthorized(c, "Invalid user ID format")
		return
	}

	var query model.OrderQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.responseHelper.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	// Check if this is a request for user's own orders or all orders
	// For simplicity, we'll use GetOrdersByUserID for now
	// In a real application, you might check user roles here
	response, err := h.service.GetOrdersByUserID(c.Request.Context(), userIDUint, query)
	if err != nil {
		h.responseHelper.InternalError(c, "Failed to get orders", err.Error())
		return
	}

	h.responseHelper.SuccessWithMeta(c, response.Data, response.Pagination)
}

// UpdateOrder handles order update
func (h *OrderHandler) UpdateOrder(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		h.responseHelper.Unauthorized(c, "User ID not found in context")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		h.responseHelper.Unauthorized(c, "Invalid user ID format")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.responseHelper.BadRequest(c, "Invalid order ID", err.Error())
		return
	}

	var req model.UpdateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHelper.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	order, err := h.service.UpdateOrder(c.Request.Context(), uint(id), req, userIDUint)
	if err != nil {
		if errors.Is(err, model.ErrOrderNotFound) {
			h.responseHelper.NotFound(c, "Order not found")
		} else if errors.Is(err, model.ErrNotAuthorizedToUpdate) {
			h.responseHelper.Forbidden(c, "Not authorized to update this order")
		} else if errors.Is(err, model.ErrInvalidStatusValue) {
			h.responseHelper.BadRequest(c, "Invalid status value", err.Error())
		} else if errors.Is(err, model.ErrCannotChangePaidOrderToPending) {
			h.responseHelper.BadRequest(c, "Cannot change paid order back to pending", err.Error())
		} else if errors.Is(err, model.ErrCannotChangeCancelledOrderStatus) {
			h.responseHelper.BadRequest(c, "Cannot change cancelled order status", err.Error())
		} else {
			h.responseHelper.InternalError(c, "Failed to update order", err.Error())
		}
		return
	}

	h.responseHelper.Success(c, order, "Order updated successfully")
}

// DeleteOrder handles order deletion
func (h *OrderHandler) DeleteOrder(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		h.responseHelper.Unauthorized(c, "User ID not found in context")
		return
	}

	userIDUint, ok := userID.(uint)
	if !ok {
		h.responseHelper.Unauthorized(c, "Invalid user ID format")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.responseHelper.BadRequest(c, "Invalid order ID", err.Error())
		return
	}

	err = h.service.DeleteOrder(c.Request.Context(), uint(id), userIDUint)
	if err != nil {
		if errors.Is(err, model.ErrOrderNotFound) {
			h.responseHelper.NotFound(c, "Order not found")
		} else if errors.Is(err, model.ErrNotAuthorizedToUpdate) {
			h.responseHelper.Forbidden(c, "Not authorized to delete this order")
		} else {
			h.responseHelper.InternalError(c, "Failed to delete order", err.Error())
		}
		return
	}

	h.responseHelper.Success(c, nil, "Order deleted successfully")
}
