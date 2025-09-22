package product

import (
	"net/http"

	"mini-e-commerce/internal/middleware"
	"mini-e-commerce/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	repo := NewRepository(db)
	service := NewService(repo)

	group := r.Group("/products", middleware.AuthMiddleware(rdb))
	{
		group.POST("", func(c *gin.Context) {
			var input CreateProductInput

			if err := c.ShouldBindJSON(&input); err != nil {
				response.Error(c, http.StatusBadRequest,
					"Invalid input request",
					response.ErrCodeValidationError,
					err.Error())
				return
			}

			product, err := service.CreateProduct(c.Request.Context(), input)
			if err != nil {
				response.Error(c, http.StatusInternalServerError,
					"Failed to create product",
					response.ErrCodeInternalServer,
					err.Error())
				return
			}

			response.Success(c, "Product created successfully", product)
		})

		group.GET("", func(c *gin.Context) {
			products, err := service.GetAll(c.Request.Context())
			if err != nil {
				response.Error(c, http.StatusInternalServerError,
					"Failed to fetch products",
					response.ErrCodeInternalServer,
					err.Error())
				return
			}
			response.Success(c, "Products fetched successfully", products)
		})

		group.GET("/:id", func(c *gin.Context) {
			id, err := service.ParseIDFromString(c.Param("id"))
			if err != nil {
				response.Error(c, http.StatusBadRequest,
					"Invalid product ID",
					response.ErrCodeValidationError,
					err.Error())
				return
			}

			product, err := service.GetByID(c.Request.Context(), id)
			if err != nil {
				response.Error(c, http.StatusNotFound,
					"Product not found",
					response.ErrCodeDataNotFound,
					err.Error())
				return
			}
			response.Success(c, "Product fetched successfully", product)
		})

		group.DELETE("/:id", func(c *gin.Context) {
			id, err := service.ParseIDFromString(c.Param("id"))
			if err != nil {
				response.Error(c, http.StatusBadRequest,
					"Invalid product ID",
					response.ErrCodeValidationError,
					err.Error())
				return
			}

			if err := service.Delete(c.Request.Context(), id); err != nil {
				response.Error(c, http.StatusInternalServerError,
					"Failed to delete product",
					response.ErrCodeInternalServer,
					err.Error())
				return
			}

			response.Success(c, "Product deleted successfully", nil)
		})

		group.PATCH("/:id", func(c *gin.Context) {
			id, err := service.ParseIDFromString(c.Param("id"))
			if err != nil {
				response.Error(c, http.StatusBadRequest,
					"Invalid product ID",
					response.ErrCodeValidationError,
					err.Error())
				return
			}

			var input UpdateProductInput
			if err := c.ShouldBindJSON(&input); err != nil {
				response.Error(c, http.StatusBadRequest,
					"Invalid input request",
					response.ErrCodeValidationError,
					err.Error())
				return
			}

			product, err := service.UpdateProduct(c.Request.Context(), id, input)
			if err != nil {
				response.Error(c, http.StatusInternalServerError,
					"Failed to update product",
					response.ErrCodeInternalServer,
					err.Error())
				return
			}

			response.Success(c, "Product updated successfully", product)
		})
	}
}
