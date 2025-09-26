package order

import (
	"errors"

	"mini-e-commerce/internal/constants"
	"mini-e-commerce/internal/middleware"
	"mini-e-commerce/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	// Domain-specific error messages (keep local)
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
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup, rdb *redis.Client) {
	group := r.Group("/orders", middleware.AuthMiddleware(rdb))

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
		response.BadRequest(c, constants.InvalidInputMessage, err.Error())
		return
	}

	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		if err.Error() == "missing user_id in context" {
			response.Error(c, constants.StatusUnauthorized, constants.UnauthorizedMessage, constants.ErrorCodeValidation, err.Error())
		} else {
			response.InternalServerError(c, ErrMsgInvalidUserContext, err.Error())
		}
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), input, userID)
	if err != nil {
		if err.Error() == ErrProductNotFound {
			response.NotFound(c, ErrMsgProductNotFound, err.Error())
			return
		}
		if err.Error() == ErrInsufficientStock {
			response.BadRequest(c, ErrMsgInsufficientStock, err.Error())
			return
		}
		response.InternalServerError(c, ErrMsgFailedToProcess, err.Error())
		return
	}

	response.SuccessCreated(c, constants.OrderCreatedMessage, order)
}

// GetOrders godoc
// @Summary Get all list order
// @Description Get all list order
// @Tags Orders
// @Accept  json
// @Produce  json
// @Success 200 {object} response.SuccessResponse{data=[]Order}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /orders [get]
func (h *Handler) GetOrders(c *gin.Context) {
	orders, err := h.service.GetAllOrders(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, ErrMsgFailedToFetch, err.Error())
		return
	}
	response.SuccessOK(c, constants.OrdersRetrievedMessage, orders)
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
		response.BadRequest(c, ErrMsgInvalidOrderID, err.Error())
		return
	}

	order, err := h.service.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.NotFound(c, ErrMsgOrderNotFound, "No order with given ID")
			return
		}
		response.InternalServerError(c, ErrMsgFailedToFetch, err.Error())
		return
	}
	response.SuccessOK(c, constants.OrderRetrievedMessage, order)
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
		response.BadRequest(c, ErrMsgInvalidOrderID, err.Error())
		return
	}

	err = h.service.DeleteOrder(c.Request.Context(), id)
	if err != nil {
		if err.Error() == ErrOrderNotFound {
			response.NotFound(c, ErrMsgOrderNotFound, err.Error())
			return
		}
		response.InternalServerError(c, ErrMsgFailedToDelete, err.Error())
		return
	}

	response.SuccessOK(c, constants.OrderDeletedMessage, nil)
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
		response.BadRequest(c, constants.InvalidInputMessage, err.Error())
		return
	}

	id, err := ParseIDFromString(c.Param("id"))
	if err != nil {
		response.BadRequest(c, ErrMsgInvalidOrderID, err.Error())
		return
	}

	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		if err.Error() == "missing user_id in context" {
			response.Error(c, constants.StatusUnauthorized, constants.UnauthorizedMessage, constants.ErrorCodeValidation, err.Error())
		} else {
			response.InternalServerError(c, ErrMsgInvalidUserContext, err.Error())
		}
		return
	}

	order, err := h.service.UpdateOrder(c.Request.Context(), id, input, userID)
	if err != nil {
		if err.Error() == ErrOrderNotFound {
			response.NotFound(c, ErrMsgOrderNotFound, err.Error())
			return
		}
		if err.Error() == ErrNotAuthorizedToUpdate {
			response.Error(c, constants.StatusUnauthorized, ErrMsgNotAuthorized, constants.ErrorCodeValidation, err.Error())
			return
		}
		if err.Error() == ErrInvalidStatusValue {
			response.BadRequest(c, ErrMsgInvalidStatus, err.Error())
			return
		}
		if err.Error() == ErrInsufficientStockForUpdate {
			response.BadRequest(c, ErrMsgInsufficientStock, err.Error())
			return
		}
		response.InternalServerError(c, ErrMsgFailedToUpdate, err.Error())
		return
	}

	response.SuccessOK(c, constants.OrderUpdatedMessage, order)
}

// Helpers
func (h *Handler) getUserIDFromContext(c *gin.Context) (uint, error) {
	userIDStr, ok := c.Get("user_id")
	if !ok {
		return 0, errors.New("missing user_id in context")
	}
	return ParseUserIDFromString(userIDStr.(string))
}
