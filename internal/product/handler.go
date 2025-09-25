package product

import (
	"net/http"

	"mini-e-commerce/internal/middleware"
	"mini-e-commerce/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	ErrMsgInvalidProductID = "Invalid product ID"
	ErrMsgInvalidInput     = "Invalid input"
	ErrMsgFailedToCreate   = "Failed to create product"
	ErrMsgFailedToFetch    = "Failed to fetch products"
	ErrMsgFailedToUpdate   = "Failed to update product"
	ErrMsgFailedToDelete   = "Failed to delete product"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup, rdb *redis.Client) {
	group := r.Group("/products", middleware.AuthMiddleware(rdb))
	group.POST("", h.CreateProduct)
	group.GET("", h.GetAllProducts)
	group.GET("/:id", h.GetProductByID)
	group.PUT("/:id", h.UpdateProduct)
	group.DELETE("/:id", h.DeleteProduct)
}

func (h *Handler) CreateProduct(c *gin.Context) {
	var input CreateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, ErrMsgInvalidInput, response.ErrCodeValidationError, err.Error())
		return
	}

	product, err := h.service.CreateProduct(c.Request.Context(), input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, ErrMsgFailedToCreate, response.ErrCodeInternalServer, err.Error())
		return
	}

	response.Success(c, "Product created successfully", product)
}

func (h *Handler) GetAllProducts(c *gin.Context) {
	products, err := h.service.GetAllProducts(c.Request.Context())
	if err != nil {
		response.Error(c, http.StatusInternalServerError, ErrMsgFailedToFetch, response.ErrCodeInternalServer, err.Error())
		return
	}
	response.Success(c, "Products fetched successfully", products)
}

func (h *Handler) GetProductByID(c *gin.Context) {
	id, err := ParseIDFromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrMsgInvalidProductID, response.ErrCodeValidationError, err.Error())
		return
	}

	product, err := h.service.GetProductByID(c.Request.Context(), id)
	if err != nil {
		response.Error(c, http.StatusNotFound, "Product not found", response.ErrCodeDataNotFound, err.Error())
		return
	}

	response.Success(c, "Product fetched successfully", product)
}

func (h *Handler) UpdateProduct(c *gin.Context) {
	id, err := ParseIDFromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrMsgInvalidProductID, response.ErrCodeValidationError, err.Error())
		return
	}

	var input UpdateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, ErrMsgInvalidInput, response.ErrCodeValidationError, err.Error())
		return
	}

	product, err := h.service.UpdateProduct(c.Request.Context(), id, input)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, ErrMsgFailedToUpdate, response.ErrCodeInternalServer, err.Error())
		return
	}

	response.Success(c, "Product updated successfully", product)
}

func (h *Handler) DeleteProduct(c *gin.Context) {
	id, err := ParseIDFromString(c.Param("id"))
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrMsgInvalidProductID, response.ErrCodeValidationError, err.Error())
		return
	}

	if err := h.service.DeleteProduct(c.Request.Context(), id); err != nil {
		response.Error(c, http.StatusInternalServerError, ErrMsgFailedToDelete, response.ErrCodeInternalServer, err.Error())
		return
	}

	response.Success(c, "Product deleted successfully", nil)
}
