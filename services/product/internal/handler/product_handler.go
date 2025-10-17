package handler

import (
	"errors"
	"strconv"

	"mini-e-commerce/services/product/internal/model"
	"mini-e-commerce/services/product/internal/service"
	"mini-e-commerce/shared/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type ProductHandler struct {
	service        service.ProductService
	logger         *zap.Logger
	responseHelper *response.ResponseHelper
}

func NewProductHandler(service service.ProductService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		service:        service,
		logger:         logger,
		responseHelper: response.NewResponseHelper(logger),
	}
}

func (h *ProductHandler) RegisterRoutes(r *gin.RouterGroup) {
	products := r.Group("/products")
	{
		products.GET("", h.GetAllProducts)
		products.GET("/:id", h.GetProductByID)
		products.POST("", h.CreateProduct)
		products.PUT("/:id", h.UpdateProduct)
		products.DELETE("/:id", h.DeleteProduct)
		products.PATCH("/:id/stock", h.UpdateStock)
		products.GET("/sku/:sku", h.GetProductBySKU)
		products.POST("/bulk", h.GetProductsByIDs)
	}
}

// CreateProduct handles product creation
func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var req model.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHelper.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	product, err := h.service.CreateProduct(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, model.ErrSKUAlreadyExists) {
			h.responseHelper.Conflict(c, "SKU already exists", err.Error())
		} else {
			h.responseHelper.BadRequest(c, "Failed to create product", err.Error())
		}
		return
	}

	h.responseHelper.Created(c, product, "Product created successfully")
}

// GetProductByID handles getting product by ID
func (h *ProductHandler) GetProductByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.responseHelper.BadRequest(c, "Invalid product ID", err.Error())
		return
	}

	product, err := h.service.GetProductByID(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, model.ErrProductNotFound) {
			h.responseHelper.NotFound(c, "Product not found")
		} else {
			h.responseHelper.InternalError(c, "Failed to get product", err.Error())
		}
		return
	}

	h.responseHelper.Success(c, product)
}

// GetProductBySKU handles getting product by SKU
func (h *ProductHandler) GetProductBySKU(c *gin.Context) {
	sku := c.Param("sku")
	if sku == "" {
		h.responseHelper.BadRequest(c, "SKU is required", nil)
		return
	}

	product, err := h.service.GetProductBySKU(c.Request.Context(), sku)
	if err != nil {
		if errors.Is(err, model.ErrProductNotFound) {
			h.responseHelper.NotFound(c, "Product not found")
		} else {
			h.responseHelper.InternalError(c, "Failed to get product", err.Error())
		}
		return
	}

	h.responseHelper.Success(c, product)
}

// GetAllProducts handles getting all products with filtering and pagination
func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	var query model.ProductQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		h.responseHelper.BadRequest(c, "Invalid query parameters", err.Error())
		return
	}

	response, err := h.service.GetAllProducts(c.Request.Context(), query)
	if err != nil {
		h.responseHelper.InternalError(c, "Failed to get products", err.Error())
		return
	}

	h.responseHelper.SuccessWithMeta(c, response.Data, response.Pagination)
}

// UpdateProduct handles product update
func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.responseHelper.BadRequest(c, "Invalid product ID", err.Error())
		return
	}

	var req model.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHelper.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	product, err := h.service.UpdateProduct(c.Request.Context(), uint(id), req)
	if err != nil {
		if errors.Is(err, model.ErrProductNotFound) {
			h.responseHelper.NotFound(c, "Product not found")
		} else if errors.Is(err, model.ErrSKUAlreadyExists) {
			h.responseHelper.Conflict(c, "SKU already exists", err.Error())
		} else {
			h.responseHelper.InternalError(c, "Failed to update product", err.Error())
		}
		return
	}

	h.responseHelper.Success(c, product, "Product updated successfully")
}

// DeleteProduct handles product deletion
func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.responseHelper.BadRequest(c, "Invalid product ID", err.Error())
		return
	}

	err = h.service.DeleteProduct(c.Request.Context(), uint(id))
	if err != nil {
		if errors.Is(err, model.ErrProductNotFound) {
			h.responseHelper.NotFound(c, "Product not found")
		} else {
			h.responseHelper.InternalError(c, "Failed to delete product", err.Error())
		}
		return
	}

	h.responseHelper.Success(c, nil, "Product deleted successfully")
}

// UpdateStock handles stock update
func (h *ProductHandler) UpdateStock(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.responseHelper.BadRequest(c, "Invalid product ID", err.Error())
		return
	}

	var req model.UpdateStockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHelper.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	err = h.service.UpdateStock(c.Request.Context(), uint(id), req.Stock)
	if err != nil {
		if errors.Is(err, model.ErrProductNotFound) {
			h.responseHelper.NotFound(c, "Product not found")
		} else if errors.Is(err, model.ErrInsufficientStock) {
			h.responseHelper.BadRequest(c, "Insufficient stock", err.Error())
		} else {
			h.responseHelper.InternalError(c, "Failed to update stock", err.Error())
		}
		return
	}

	h.responseHelper.Success(c, nil, "Stock updated successfully")
}

// GetProductsByIDs handles bulk product retrieval by IDs
func (h *ProductHandler) GetProductsByIDs(c *gin.Context) {
	var req model.BulkProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHelper.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	products, err := h.service.GetProductsByIDs(c.Request.Context(), req.ProductIDs)
	if err != nil {
		h.responseHelper.InternalError(c, "Failed to get products", err.Error())
		return
	}

	// Convert to ProductInfo format for order service
	productInfos := make([]model.ProductInfo, len(products))
	for i, product := range products {
		productInfos[i] = model.ProductInfo{
			ID:    product.ID,
			Name:  product.Name,
			Price: product.Price,
			Stock: product.Stock,
		}
	}

	h.responseHelper.Success(c, gin.H{"data": productInfos})
}
