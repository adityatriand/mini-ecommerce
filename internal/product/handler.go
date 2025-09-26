package product

import (
	"mini-e-commerce/internal/constants"
	"mini-e-commerce/internal/middleware"
	"mini-e-commerce/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

const (
	// Domain-specific error messages (keep local)
	ErrMsgInvalidProductID = "Invalid product ID"
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
	group.PATCH("/:id", h.UpdateProduct)
	group.DELETE("/:id", h.DeleteProduct)
}

// CreateProduct godoc
// @Summary Create a new product
// @Description Create a new product with name, price, and stock
// @Tags Products
// @Accept  json
// @Produce  json
// @Param   request body CreateProductRequest true "Product request body"
// @Success 201 {object} response.SuccessResponse{data=Product}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products [post]
func (h *Handler) CreateProduct(c *gin.Context) {
	var input CreateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, constants.InvalidInputMessage, err.Error())
		return
	}

	product, err := h.service.CreateProduct(c.Request.Context(), input)
	if err != nil {
		response.InternalServerError(c, ErrMsgFailedToCreate, err.Error())
		return
	}

	response.SuccessCreated(c, constants.ProductCreatedMessage, product)
}

// GetAllProducts godoc
// @Summary Get all products
// @Description Get a list of all products
// @Tags Products
// @Accept  json
// @Produce  json
// @Success 200 {object} response.SuccessResponse{data=[]Product}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products [get]
func (h *Handler) GetAllProducts(c *gin.Context) {
	products, err := h.service.GetAllProducts(c.Request.Context())
	if err != nil {
		response.InternalServerError(c, ErrMsgFailedToFetch, err.Error())
		return
	}
	response.SuccessOK(c, constants.ProductsRetrievedMessage, products)
}

// GetProductByID godoc
// @Summary Get single product
// @Description Get product by id
// @Tags Products
// @Accept  json
// @Produce  json
// @Param   id path string true "Product ID"
// @Success 200 {object} response.SuccessResponse{data=Product}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products/{id} [get]
func (h *Handler) GetProductByID(c *gin.Context) {
	id, err := ParseIDFromString(c.Param("id"))
	if err != nil {
		response.BadRequest(c, ErrMsgInvalidProductID, err.Error())
		return
	}

	product, err := h.service.GetProductByID(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, constants.ProductNotFoundMessage, err.Error())
		return
	}

	response.SuccessOK(c, constants.ProductsRetrievedMessage, product)
}

// UpdateProduct godoc
// @Summary Update exist product
// @Description Update single product
// @Tags Products
// @Accept  json
// @Produce  json
// @Param   id path string true "Product ID"
// @Param   request body UpdateProductRequest true "Product request body"
// @Success 200 {object} response.SuccessResponse{data=Product}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products/{id} [patch]
func (h *Handler) UpdateProduct(c *gin.Context) {
	id, err := ParseIDFromString(c.Param("id"))
	if err != nil {
		response.BadRequest(c, ErrMsgInvalidProductID, err.Error())
		return
	}

	var input UpdateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, constants.InvalidInputMessage, err.Error())
		return
	}

	product, err := h.service.UpdateProduct(c.Request.Context(), id, input)
	if err != nil {
		response.InternalServerError(c, ErrMsgFailedToUpdate, err.Error())
		return
	}

	response.SuccessOK(c, constants.ProductUpdatedMessage, product)
}

// DeleteProduct godoc
// @Summary Delete exist product
// @Description Delete exist single product
// @Tags Products
// @Accept  json
// @Produce  json
// @Param   id path string true "Product ID"
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products/{id} [delete]
func (h *Handler) DeleteProduct(c *gin.Context) {
	id, err := ParseIDFromString(c.Param("id"))
	if err != nil {
		response.BadRequest(c, ErrMsgInvalidProductID, err.Error())
		return
	}

	if err := h.service.DeleteProduct(c.Request.Context(), id); err != nil {
		response.InternalServerError(c, ErrMsgFailedToDelete, err.Error())
		return
	}

	response.SuccessOK(c, constants.ProductDeletedMessage, nil)
}
