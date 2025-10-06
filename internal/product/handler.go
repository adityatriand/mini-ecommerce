package product

import (
	"mini-e-commerce/internal/auth"
	"mini-e-commerce/internal/logger"
	"mini-e-commerce/internal/middleware"
	"mini-e-commerce/internal/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	ErrMsgInvalidProductID = "Invalid product ID"
	ErrMsgFailedToCreate   = "Failed to create product"
	ErrMsgFailedToFetch    = "Failed to fetch products"
	ErrMsgFailedToUpdate   = "Failed to update product"
	ErrMsgFailedToDelete   = "Failed to delete product"
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
	group := r.Group("/products", authMiddleware)
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
		h.responseHelper.BadRequest(c, response.ErrCodeValidationError, err.Error())
		return
	}

	product, err := h.service.CreateProduct(c.Request.Context(), input)
	if err != nil {
		h.responseHelper.InternalServerError(c, ErrMsgFailedToCreate, err.Error())
		return
	}

	ctxLogger := h.logger.WithContext(c)
	ctxLogger.Info("Product added to inventory",
		zap.Uint("product_id", product.ID),
		zap.String("product_name", product.Name),
		zap.Int("price", product.Price),
		zap.Int("initial_stock", product.Stock),
	)

	h.responseHelper.SuccessCreated(c, "Product created successfully", product)
}

// GetAllProducts godoc
// @Summary Get all products
// @Description Get a list of all products with pagination and filtering
// @Tags Products
// @Accept  json
// @Produce  json
// @Param page query int false "Page number" minimum(1)
// @Param page_size query int false "Page size" minimum(1) maximum(100)
// @Param order query string false "Sort order" Enums(asc, desc)
// @Param sort_by query string false "Sort by field" Enums(id, name, price, stock, created_at)
// @Success 200 {object} response.SuccessResponse{data=ProductListResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /products [get]
func (h *Handler) GetAllProducts(c *gin.Context) {
	var query ProductQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.responseHelper.BadRequest(c, response.ErrCodeValidationError, err.Error())
		return
	}

	result, err := h.service.GetAllProductsWithQuery(c.Request.Context(), query)
	if err != nil {
		h.responseHelper.InternalServerError(c, ErrMsgFailedToFetch, err.Error())
		return
	}
	h.responseHelper.SuccessPaginated(c, "List product retrieved successfully", result.Data, result.Pagination)

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
		h.responseHelper.BadRequest(c, ErrMsgInvalidProductID, err.Error())
		return
	}

	product, err := h.service.GetProductByID(c.Request.Context(), id)
	if err != nil {
		h.responseHelper.NotFound(c, response.ErrCodeDataNotFound, err.Error())
		return
	}

	h.responseHelper.SuccessOK(c, "Product retrieved successfully", product)

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
		h.responseHelper.BadRequest(c, ErrMsgInvalidProductID, err.Error())
		return
	}

	var input UpdateProductRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		h.responseHelper.BadRequest(c, response.ErrCodeValidationError, err.Error())
		return
	}

	product, err := h.service.UpdateProduct(c.Request.Context(), id, input)
	if err != nil {
		h.responseHelper.InternalServerError(c, ErrMsgFailedToUpdate, err.Error())
		return
	}

	ctxLogger := h.logger.WithContext(c)
	ctxLogger.Info("Product inventory updated",
		zap.Uint("product_id", product.ID),
		zap.String("product_name", product.Name),
		zap.Int("new_price", product.Price),
		zap.Int("new_stock", product.Stock),
	)

	h.responseHelper.SuccessOK(c, "Product updated successfully", product)
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
		h.responseHelper.BadRequest(c, ErrMsgInvalidProductID, err.Error())
		return
	}

	if err := h.service.DeleteProduct(c.Request.Context(), id); err != nil {
		h.responseHelper.InternalServerError(c, ErrMsgFailedToDelete, err.Error())
		return
	}

	ctxLogger := h.logger.WithContext(c)
	ctxLogger.Info("Product removed from inventory",
		zap.Uint("product_id", id),
	)

	h.responseHelper.SuccessOK(c, "Product deleted successfully", nil)
}
