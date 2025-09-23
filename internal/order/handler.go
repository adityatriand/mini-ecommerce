package order

import (
	"errors"
	"net/http"

	"mini-e-commerce/internal/middleware"
	"mini-e-commerce/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	ErrMsgInvalidOrderID     = "Invalid order ID"
	ErrMsgOrderNotFound      = "Order not found"
	ErrMsgProductNotFound    = "Product not found"
	ErrMsgInsufficientStock  = "Stock product not available"
	ErrMsgNotAuthorized      = "Not allowed to update this order"
	ErrMsgInvalidStatus      = "Invalid status value"
	ErrMsgUnauthorized       = "Unauthorized"
	ErrMsgInvalidInput       = "Invalid input request"
	ErrMsgInvalidUserContext = "Invalid user id in context"
	ErrMsgFailedToProcess    = "Failed to process order"
	ErrMsgFailedToFetch      = "Failed to fetch order"
	ErrMsgFailedToDelete     = "Failed to delete order"
	ErrMsgFailedToUpdate     = "Failed to update order"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
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

func (h *Handler) CreateOrder(c *gin.Context) {
	var input CreateOrderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, ErrMsgInvalidInput, response.ErrCodeValidationError, err.Error())
		return
	}

	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		if err.Error() == "missing user_id in context" {
			response.Error(c, http.StatusUnauthorized, ErrMsgUnauthorized, response.ErrCodeUnauthorized, err.Error())
		} else {
			response.Error(c, http.StatusInternalServerError, ErrMsgInvalidUserContext, response.ErrCodeInternalServer, err.Error())
		}
		return
	}

	order, err := h.service.CreateOrder(c.Request.Context(), input, userID)
	if err != nil {
		if err.Error() == ErrProductNotFound {
			response.Error(c, http.StatusNotFound, ErrMsgProductNotFound, response.ErrCodeDataNotFound, err.Error())
			return
		}
		if err.Error() == ErrInsufficientStock {
			response.Error(c, http.StatusBadRequest, ErrMsgInsufficientStock, response.ErrCodeValidationError, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, ErrMsgFailedToProcess, response.ErrCodeInternalServer, err.Error())
		return
	}

	response.Success(c, "Order created successfully", order)
}

func (h *Handler) GetOrders(c *gin.Context) {
	orders, err := h.service.GetAllOrders(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, ErrMsgFailedToFetch, response.ErrCodeInternalServer, err.Error())
		return
	}
	response.Success(c, "Orders fetched successfully", orders)
}

func (h *Handler) GetOrderByID(c *gin.Context) {
	id, err := h.service.ParseIDFromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrMsgInvalidOrderID, response.ErrCodeValidationError, err.Error())
		return
	}

	order, err := h.service.GetOrderByID(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, http.StatusNotFound, ErrMsgOrderNotFound, response.ErrCodeDataNotFound, "No order with given ID")
			return
		}
		response.Error(c, http.StatusInternalServerError, ErrMsgFailedToFetch, response.ErrCodeInternalServer, err.Error())
		return
	}
	response.Success(c, "Order fetched successfully", order)
}

func (h *Handler) DeleteOrder(c *gin.Context) {
	id, err := h.service.ParseIDFromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrMsgInvalidOrderID, response.ErrCodeValidationError, err.Error())
		return
	}

	err = h.service.DeleteOrder(c.Request.Context(), id)
	if err != nil {
		if err.Error() == ErrOrderNotFound {
			response.Error(c, http.StatusNotFound, ErrMsgOrderNotFound, response.ErrCodeDataNotFound, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, ErrMsgFailedToDelete, response.ErrCodeInternalServer, err.Error())
		return
	}

	response.Success(c, "Order deleted successfully", nil)
}

func (h *Handler) UpdateOrder(c *gin.Context) {
	var input UpdateOrderRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, ErrMsgInvalidInput, response.ErrCodeValidationError, err.Error())
		return
	}

	id, err := h.service.ParseIDFromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrMsgInvalidOrderID, response.ErrCodeValidationError, err.Error())
		return
	}

	userID, err := h.getUserIDFromContext(c)
	if err != nil {
		if err.Error() == "missing user_id in context" {
			response.Error(c, http.StatusUnauthorized, ErrMsgUnauthorized, response.ErrCodeUnauthorized, err.Error())
		} else {
			response.Error(c, http.StatusInternalServerError, ErrMsgInvalidUserContext, response.ErrCodeInternalServer, err.Error())
		}
		return
	}

	order, err := h.service.UpdateOrder(c.Request.Context(), id, input, userID)
	if err != nil {
		if err.Error() == ErrOrderNotFound {
			response.Error(c, http.StatusNotFound, ErrMsgOrderNotFound, response.ErrCodeDataNotFound, err.Error())
			return
		}
		if err.Error() == ErrNotAuthorizedToUpdate {
			response.Error(c, http.StatusForbidden, ErrMsgNotAuthorized, response.ErrCodeForbidden, err.Error())
			return
		}
		if err.Error() == ErrInvalidStatusValue {
			response.Error(c, http.StatusBadRequest, ErrMsgInvalidStatus, response.ErrCodeValidationError, err.Error())
			return
		}
		if err.Error() == ErrInsufficientStockForUpdate {
			response.Error(c, http.StatusBadRequest, ErrMsgInsufficientStock, response.ErrCodeValidationError, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, ErrMsgFailedToUpdate, response.ErrCodeInternalServer, err.Error())
		return
	}

	response.Success(c, "Order updated successfully", order)
}

func (h *Handler) getUserIDFromContext(c *gin.Context) (uint, error) {
	userIDStr, ok := c.Get("user_id")
	if !ok {
		return 0, errors.New("missing user_id in context")
	}
	return h.service.ParseUserIDFromString(userIDStr.(string))
}
