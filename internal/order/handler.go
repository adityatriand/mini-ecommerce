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

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	repo := NewRepository(db)
	service := NewService(repo, db)

	group := r.Group("/orders", middleware.AuthMiddleware(rdb))
	{
		group.POST("", func(c *gin.Context) {
			var input CreateOrderInput

			if err := c.ShouldBindJSON(&input); err != nil {
				response.Error(c, http.StatusBadRequest, "Invalid input request", response.ErrCodeValidationError, err.Error())
				return
			}

			userIDStr, ok := c.Get("user_id")
			if !ok {
				response.Error(c, http.StatusUnauthorized, "Unauthorized", response.ErrCodeUnauthorized, "Missing user_id in context")
				return
			}

			userID, err := service.ParseUserIDFromString(userIDStr.(string))
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Invalid user id in context", response.ErrCodeInternalServer, err.Error())
				return
			}

			order, err := service.CreateOrder(c.Request.Context(), input, userID)
			if err != nil {
				if err.Error() == "product not found" {
					response.Error(c, http.StatusNotFound, "Product not found", response.ErrCodeDataNotFound, err.Error())
					return
				}
				if err.Error() == "insufficient stock" {
					response.Error(c, http.StatusBadRequest, "Stock product not available", response.ErrCodeValidationError, err.Error())
					return
				}
				response.Error(c, http.StatusInternalServerError, "Failed to process order", response.ErrCodeInternalServer, err.Error())
				return
			}

			response.Success(c, "Order created successfully", order)
		})

		group.GET("", func(c *gin.Context) {
			orders, err := service.GetAll(c.Request.Context())
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to fetch orders", response.ErrCodeInternalServer, err.Error())
				return
			}
			response.Success(c, "Orders fetched successfully", orders)
		})

		group.GET("/:id", func(c *gin.Context) {
			id, err := service.ParseIDFromString(c.Param("id"))
			if err != nil {
				response.Error(c, http.StatusBadRequest, "Invalid order ID", response.ErrCodeValidationError, err.Error())
				return
			}

			ord, err := service.GetByID(c.Request.Context(), id)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					response.Error(c, http.StatusNotFound, "Order not found", response.ErrCodeDataNotFound, "No order with given ID")
					return
				}
				response.Error(c, http.StatusInternalServerError, "Failed to fetch order", response.ErrCodeInternalServer, err.Error())
				return
			}
			response.Success(c, "Order fetched successfully", ord)
		})

		group.DELETE("/:id", func(c *gin.Context) {
			id, err := service.ParseIDFromString(c.Param("id"))
			if err != nil {
				response.Error(c, http.StatusBadRequest, "Invalid order ID", response.ErrCodeValidationError, err.Error())
				return
			}

			err = service.DeleteOrder(c.Request.Context(), id)
			if err != nil {
				if err.Error() == "order not found" {
					response.Error(c, http.StatusNotFound, "Order not found", response.ErrCodeDataNotFound, err.Error())
					return
				}
				response.Error(c, http.StatusInternalServerError, "Failed to delete order", response.ErrCodeInternalServer, err.Error())
				return
			}

			response.Success(c, "Order deleted successfully", nil)
		})

		group.PATCH("/:id", func(c *gin.Context) {
			var input UpdateOrderInput
			if err := c.ShouldBindJSON(&input); err != nil {
				response.Error(c, http.StatusBadRequest, "Invalid input request", response.ErrCodeValidationError, err.Error())
				return
			}

			id, err := service.ParseIDFromString(c.Param("id"))
			if err != nil {
				response.Error(c, http.StatusBadRequest, "Invalid order ID", response.ErrCodeValidationError, err.Error())
				return
			}

			userIDStr, ok := c.Get("user_id")
			if !ok {
				response.Error(c, http.StatusUnauthorized, "Unauthorized", response.ErrCodeUnauthorized, "Missing user_id in context")
				return
			}

			userID, err := service.ParseUserIDFromString(userIDStr.(string))
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Invalid user id in context", response.ErrCodeInternalServer, err.Error())
				return
			}

			order, err := service.UpdateOrder(c.Request.Context(), id, input, userID)
			if err != nil {
				if err.Error() == "order not found" {
					response.Error(c, http.StatusNotFound, "Order not found", response.ErrCodeDataNotFound, err.Error())
					return
				}
				if err.Error() == "not authorized to update this order" {
					response.Error(c, http.StatusForbidden, "Not allowed to update this order", response.ErrCodeForbidden, err.Error())
					return
				}
				if err.Error() == "invalid status value" {
					response.Error(c, http.StatusBadRequest, "Invalid status value", response.ErrCodeValidationError, err.Error())
					return
				}
				if err.Error() == "insufficient stock for quantity update" {
					response.Error(c, http.StatusBadRequest, "Stock product not available", response.ErrCodeValidationError, err.Error())
					return
				}
				response.Error(c, http.StatusInternalServerError, "Failed to update order", response.ErrCodeInternalServer, err.Error())
				return
			}

			response.Success(c, "Order updated successfully", order)
		})

	}
}
