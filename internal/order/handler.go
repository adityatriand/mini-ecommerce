package order

import (
	"errors"
	"net/http"

	"mini-e-commerce/internal/auth"
	"mini-e-commerce/internal/logger"
	"mini-e-commerce/internal/middleware"
	"mini-e-commerce/internal/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	ErrMsgInvalidOrderID     = "Invalid order ID"
	ErrMsgOrderNotFound      = "Order not found"
	ErrMsgProductNotFound    = "Product not found"
	ErrMsgInsufficientStock  = "Stock product not available"
	ErrMsgNotAuthorized      = "Not allowed to update this order"
	ErrMsgInvalidStatus      = "Invalid status value"
	ErrMsgInvalidUserContext = "Invalid user id in context"
	ErrMsgFailedToProcess    = "Failed to process order"
	ErrMsgFailedToFetch      = "Failed to fetch order"
	ErrMsgFailedToDelete     = "Failed to delete order"
	ErrMsgFailedToUpdate     = "Failed to update order"
)

type Handler struct {
	service        Service
	logger         logger.Logger
	responseHelper *response.ResponseHelper
}

func NewHandler(service Service, log logger.Logger) *Handler {
	return &Handler{
		service:        service,
		logger:         log,
		responseHelper: response.NewResponseHelper(log),
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup, jwtManager *auth.JWTManager, sessionManager *auth.SessionManager, logger *zap.Logger) {
	authMiddleware := middleware.AuthMiddleware(jwtManager, sessionManager, logger)
	group := r.Group("/orders", authMiddleware)

	group.POST("", h.CreateOrder)
	group.GET("", h.GetOrders)
	group.GET("/:id", h.GetOrderByID)
	group.DELETE("/:id", h.DeleteOrder)
	group.PATCH("/:id", h.UpdateOrder)
}

// CreateOrder godoc
// @Summary Create new order
// @Description Create new order with one product and quantity
// @Tags Orders
// @Accept  json
// @Produce  json
// @Param   request body CreateOrderRequest true "Order body request"
// @Success 201 {object} response.SuccessResponse{data=Order}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	var input CreateOrderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.responseHelper.BadRequest(c, response.ErrCodeValidationError, err.Error())
		return
	}

	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		if err.Error() == "missing user_id in context" {
			h.responseHelper.Error(c, http.StatusUnauthorized, response.ErrCodeUnauthorized, response.ErrCodeUnauthorized, err.Error())
		} else {
			h.responseHelper.InternalServerError(c, ErrMsgInvalidUserContext, err.Error())
		}
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), input, userID)
	if err != nil {
		if err.Error() == ErrProductNotFound {
			h.responseHelper.NotFound(c, ErrMsgProductNotFound, err.Error())
			return
		}
		if err.Error() == ErrInsufficientStock {
			h.responseHelper.BadRequest(c, ErrMsgInsufficientStock, err.Error())
			return
		}
		h.responseHelper.InternalServerError(c, ErrMsgFailedToProcess, err.Error())
		return
	}

	ctxLogger := h.logger.WithContext(c)
	ctxLogger.Info("Order placed",
		zap.Uint("order_id", order.ID),
		zap.Uint("user_id", userID),
		zap.Uint("product_id", input.ProductID),
		zap.Int("quantity", input.Quantity),
		zap.Int("total_amount", order.TotalPrice),
	)

	h.responseHelper.SuccessCreated(c, "Order created successfully", order)

}

// GetOrders godoc
// @Summary Get all list order
// @Description Get all list order
// @Tags Orders
// @Accept  json
// @Produce  json
// @Param page query int false "Page number" minimum(1)
// @Param page_size query int false "Page size" minimum(1) maximum(100)
// @Param order query string false "Sort order" Enums(asc, desc)
// @Param sort_by query string false "Sort by field" Enums(id, user_id, product_id, quantity, total_price, status, created_at)
// @Success 200 {object} response.SuccessResponse{data=OrderListResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /orders [get]
func (h *Handler) GetOrders(c *gin.Context) {
	var query OrderQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.responseHelper.BadRequest(c, response.ErrCodeValidationError, err.Error())
		return
	}

	result, err := h.service.GetAllOrdersWithQuery(c.Request.Context(), query)
	if err != nil {
		h.responseHelper.InternalServerError(c, ErrMsgFailedToFetch, err.Error())
		return
	}
	h.responseHelper.SuccessPaginated(c, "List Order retrieved successfully", result.Data, result.Pagination)
}

// GetOrderByID godoc
// @Summary Get single order
// @Description Get an order by id
// @Tags Orders
// @Accept  json
// @Produce  json
// @Param   id path string true "Order ID"
// @Success 200 {object} response.SuccessResponse{data=Order}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /orders/{id} [get]
func (h *Handler) GetOrderByID(c *gin.Context) {
	id, err := ParseIDFromString(c.Param("id"))
	if err != nil {
		h.responseHelper.BadRequest(c, ErrMsgInvalidOrderID, err.Error())
		return
	}

	order, err := h.service.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			h.responseHelper.NotFound(c, ErrMsgOrderNotFound, "No order with given ID")
			return
		}
		h.responseHelper.InternalServerError(c, ErrMsgFailedToFetch, err.Error())
		return
	}
	h.responseHelper.SuccessOK(c, "Order retrieved successfully", order)
}

// DeleteOrder godoc
// @Summary Delete single product
// @Description Delete an order by id
// @Tags Orders
// @Accept  json
// @Produce  json
// @Param   id path string true "Order ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /orders/{id} [delete]
func (h *Handler) DeleteOrder(c *gin.Context) {
	id, err := ParseIDFromString(c.Param("id"))
	if err != nil {
		h.responseHelper.BadRequest(c, ErrMsgInvalidOrderID, err.Error())
		return
	}

	err = h.service.DeleteOrder(c.Request.Context(), id)
	if err != nil {
		if err.Error() == ErrOrderNotFound {
			h.responseHelper.NotFound(c, ErrMsgOrderNotFound, err.Error())
			return
		}
		h.responseHelper.InternalServerError(c, ErrMsgFailedToDelete, err.Error())
		return
	}

	ctxLogger := h.logger.WithContext(c)
	ctxLogger.Info("Order cancelled",
		zap.Uint("order_id", id),
	)

	h.responseHelper.SuccessOK(c, "Order deleted successfully", nil)
}

// UpdateProduct godoc
// @Summary Update an order
// @Description Update an order by Id
// @Tags Orders
// @Accept  json
// @Produce  json
// @Param   id path string true "Order ID"
// @Param   request body UpdateOrderRequest true "Order body request"
// @Success 200 {object} response.SuccessResponse{data=Order}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /orders/{id} [patch]
func (h *Handler) UpdateOrder(c *gin.Context) {
	var input UpdateOrderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.responseHelper.BadRequest(c, response.ErrCodeValidationError, err.Error())
		return
	}

	id, err := ParseIDFromString(c.Param("id"))
	if err != nil {
		h.responseHelper.BadRequest(c, ErrMsgInvalidOrderID, err.Error())
		return
	}

	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		if err.Error() == "missing user_id in context" {
			h.responseHelper.Error(c, http.StatusUnauthorized, response.ErrCodeUnauthorized, response.ErrCodeUnauthorized, err.Error())
		} else {
			h.responseHelper.InternalServerError(c, ErrMsgInvalidUserContext, err.Error())
		}
		return
	}

	order, err := h.service.UpdateOrder(c.Request.Context(), id, input, userID)
	if err != nil {
		if err.Error() == ErrOrderNotFound {
			h.responseHelper.NotFound(c, ErrMsgOrderNotFound, err.Error())
			return
		}
		if err.Error() == ErrNotAuthorizedToUpdate {
			h.responseHelper.Error(c, http.StatusUnauthorized, ErrMsgNotAuthorized, response.ErrCodeUnauthorized, err.Error())
			return
		}
		if err.Error() == ErrInvalidStatusValue {
			h.responseHelper.BadRequest(c, ErrMsgInvalidStatus, err.Error())
			return
		}
		if err.Error() == ErrInsufficientStockForUpdate {
			h.responseHelper.BadRequest(c, ErrMsgInsufficientStock, err.Error())
			return
		}
		h.responseHelper.InternalServerError(c, ErrMsgFailedToUpdate, err.Error())
		return
	}

	ctxLogger := h.logger.WithContext(c)
	ctxLogger.Info("Order status changed",
		zap.Uint("order_id", order.ID),
		zap.Uint("user_id", userID),
		zap.Any("new_status", order.Status),
	)

	h.responseHelper.SuccessOK(c, "Order updated successfully", order)
}

// Helpers
func (h *Handler) getUserIDFromContext(c *gin.Context) (uint, error) {
	userIDStr, ok := c.Get("user_id")
	if !ok {
		return 0, errors.New("missing user_id in context")
	}
	return ParseUserIDFromString(userIDStr.(string))
}
