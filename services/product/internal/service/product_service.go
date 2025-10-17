package service

import (
	"context"
	"errors"
	"fmt"
	"mini-e-commerce/services/product/internal/model"
	"mini-e-commerce/services/product/internal/repository"
	"mini-e-commerce/shared/cache"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	ErrProductNotFound  = "product not found"
	ErrInsufficientStock = "insufficient stock"
	ErrSKUAlreadyExists = "SKU already exists"
	CacheKeyProductByID = "product:id:%d"
	CacheKeyProductList = "product:list:%d:%d:%s:%s:%s" // page:pageSize:sortBy:order:category
	CacheTTLProduct     = 5 * time.Minute
	CacheTTLProductList = 2 * time.Minute
)

// ProductService defines the product service interface
type ProductService interface {
	CreateProduct(ctx context.Context, input model.CreateProductRequest) (*model.Product, error)
	GetProductByID(ctx context.Context, id uint) (*model.Product, error)
	GetAllProducts(ctx context.Context, query model.ProductQuery) (*model.ProductListResponse, error)
	UpdateProduct(ctx context.Context, id uint, input model.UpdateProductRequest) (*model.Product, error)
	DeleteProduct(ctx context.Context, id uint) error
	UpdateStock(ctx context.Context, id uint, stockDelta int) error
	UpdateStockWithTx(tx *gorm.DB, id uint, stockDelta int) error
	GetProductsByIDs(ctx context.Context, ids []uint) ([]model.Product, error)
	GetProductBySKU(ctx context.Context, sku string) (*model.Product, error)
}

// productService implements ProductService interface
type productService struct {
	repo      repository.ProductRepository
	cache     cache.Interface
	validator *validator.Validate
	logger    *zap.Logger
}

func NewProductService(repo repository.ProductRepository, cache cache.Interface, logger *zap.Logger) ProductService {
	return &productService{
		repo:      repo,
		cache:     cache,
		validator: validator.New(),
		logger:    logger,
	}
}

func (s *productService) CreateProduct(ctx context.Context, input model.CreateProductRequest) (*model.Product, error) {
	// Validate input
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	// Check if SKU already exists (if provided)
	if input.SKU != "" {
		_, err := s.repo.FindBySKU(ctx, input.SKU)
		if err == nil {
			return nil, errors.New(ErrSKUAlreadyExists)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
	}

	// Create product
	product := &model.Product{
		Name:        input.Name,
		Description: input.Description,
		Price:       input.Price,
		Stock:       input.Stock,
		Category:    input.Category,
		SKU:         input.SKU,
		IsActive:    true,
	}

	if err := s.repo.Create(ctx, product); err != nil {
		return nil, err
	}

	// Invalidate cache
	s.invalidateProductListCache(ctx)

	s.logger.Info("Product created successfully", zap.Uint("product_id", product.ID), zap.String("name", product.Name))
	return product, nil
}

// GetProductByID retrieves a product by ID
func (s *productService) GetProductByID(ctx context.Context, id uint) (*model.Product, error) {
	// Try cache first
	cacheKey := fmt.Sprintf(CacheKeyProductByID, id)
	var product model.Product
	err := s.cache.Get(ctx, cacheKey, &product)
	if err == nil {
		return &product, nil
	}

	if err != redis.Nil {
		s.logger.Warn("Cache error on GetProductByID, falling back to database",
			zap.Uint("product_id", id),
			zap.Error(err),
		)
	}

	// Get from database
	product, err = s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrProductNotFound)
		}
		return nil, err
	}

	// Cache the result
	_ = s.cache.Set(ctx, cacheKey, product, CacheTTLProduct)

	return &product, nil
}

// GetProductBySKU retrieves a product by SKU
func (s *productService) GetProductBySKU(ctx context.Context, sku string) (*model.Product, error) {
	product, err := s.repo.FindBySKU(ctx, sku)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrProductNotFound)
		}
		return nil, err
	}

	return &product, nil
}

// GetAllProducts retrieves products with filtering and pagination
func (s *productService) GetAllProducts(ctx context.Context, query model.ProductQuery) (*model.ProductListResponse, error) {
	// Set defaults
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 10
	}
	if query.PageSize > 100 {
		query.PageSize = 100
	}

	// Try cache first
	cacheKey := fmt.Sprintf(CacheKeyProductList, query.Page, query.PageSize, query.SortBy, query.Order, query.Category)
	var response model.ProductListResponse
	err := s.cache.Get(ctx, cacheKey, &response)
	if err == nil {
		return &response, nil
	}

	if err != redis.Nil {
		s.logger.Warn("Cache error on GetAllProducts, falling back to database",
			zap.Int("page", query.Page),
			zap.Int("page_size", query.PageSize),
			zap.Error(err),
		)
	}

	// Get from database
	products, total, err := s.repo.FindAll(ctx, query)
	if err != nil {
		return nil, err
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(query.PageSize) - 1) / int64(query.PageSize))

	response = model.ProductListResponse{
		Data: products,
		Pagination: model.PaginationMetadata{
			Page:       query.Page,
			PageSize:   query.PageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	// Cache the result
	_ = s.cache.Set(ctx, cacheKey, response, CacheTTLProductList)

	return &response, nil
}

// UpdateProduct updates a product
func (s *productService) UpdateProduct(ctx context.Context, id uint, input model.UpdateProductRequest) (*model.Product, error) {
	// Validate input
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	// Find product
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrProductNotFound)
		}
		return nil, err
	}

	// Update fields if provided
	if input.Name != nil {
		product.Name = *input.Name
	}
	if input.Description != nil {
		product.Description = *input.Description
	}
	if input.Price != nil {
		product.Price = *input.Price
	}
	if input.Stock != nil {
		product.Stock = *input.Stock
	}
	if input.Category != nil {
		product.Category = *input.Category
	}
	if input.SKU != nil && *input.SKU != product.SKU {
		// Check if new SKU already exists
		_, err := s.repo.FindBySKU(ctx, *input.SKU)
		if err == nil {
			return nil, errors.New(ErrSKUAlreadyExists)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		product.SKU = *input.SKU
	}
	if input.IsActive != nil {
		product.IsActive = *input.IsActive
	}

	// Save changes
	if err := s.repo.Update(ctx, &product); err != nil {
		return nil, err
	}

	// Invalidate cache
	s.invalidateProductCache(ctx, id)
	s.invalidateProductListCache(ctx)

	s.logger.Info("Product updated successfully", zap.Uint("product_id", product.ID))
	return &product, nil
}

// DeleteProduct deletes a product
func (s *productService) DeleteProduct(ctx context.Context, id uint) error {
	// Check if product exists
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(ErrProductNotFound)
		}
		return err
	}

	// Delete product
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache
	s.invalidateProductCache(ctx, id)
	s.invalidateProductListCache(ctx)

	s.logger.Info("Product deleted successfully", zap.Uint("product_id", id))
	return nil
}

// UpdateStock updates product stock
func (s *productService) UpdateStock(ctx context.Context, id uint, stockDelta int) error {
	// Check if product exists
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(ErrProductNotFound)
		}
		return err
	}

	// Check if stock would go negative
	if product.Stock+stockDelta < 0 {
		return errors.New(ErrInsufficientStock)
	}

	// Update stock
	if err := s.repo.UpdateStock(ctx, id, stockDelta); err != nil {
		return err
	}

	// Invalidate cache
	s.invalidateProductCache(ctx, id)

	s.logger.Info("Product stock updated", zap.Uint("product_id", id), zap.Int("stock_delta", stockDelta))
	return nil
}

// UpdateStockWithTx updates product stock within a transaction
func (s *productService) UpdateStockWithTx(tx *gorm.DB, id uint, stockDelta int) error {
	err := s.repo.UpdateStockOptimizedWithTx(tx, id, stockDelta)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(ErrProductNotFound)
		}
		return err
	}

	// Invalidate cache (use background context since we're in a transaction)
	s.invalidateProductCache(context.Background(), id)
	return nil
}

// GetProductsByIDs retrieves multiple products by IDs
func (s *productService) GetProductsByIDs(ctx context.Context, ids []uint) ([]model.Product, error) {
	if len(ids) == 0 {
		return []model.Product{}, nil
	}

	// Limit the number of IDs to prevent potential DoS
	if len(ids) > 1000 {
		return nil, errors.New("too many IDs requested")
	}

	// Try to get from cache first
	cachedProducts := make([]model.Product, 0, len(ids))
	uncachedIDs := make([]uint, 0, len(ids))

	for _, id := range ids {
		cacheKey := fmt.Sprintf(CacheKeyProductByID, id)
		var product model.Product
		err := s.cache.Get(ctx, cacheKey, &product)
		if err == nil {
			cachedProducts = append(cachedProducts, product)
		} else {
			uncachedIDs = append(uncachedIDs, id)
		}
	}

	// Get uncached products from database
	var dbProducts []model.Product
	if len(uncachedIDs) > 0 {
		var err error
		dbProducts, err = s.repo.FindMultipleByIDs(ctx, uncachedIDs)
		if err != nil {
			return nil, err
		}

		// Cache the results
		for _, product := range dbProducts {
			cacheKey := fmt.Sprintf(CacheKeyProductByID, product.ID)
			_ = s.cache.Set(ctx, cacheKey, product, CacheTTLProduct)
		}
	}

	// Combine cached and database results
	allProducts := make([]model.Product, 0, len(cachedProducts)+len(dbProducts))
	allProducts = append(allProducts, cachedProducts...)
	allProducts = append(allProducts, dbProducts...)

	return allProducts, nil
}

// Helper methods

// invalidateProductCache invalidates cache for a specific product
func (s *productService) invalidateProductCache(ctx context.Context, id uint) {
	cacheKey := fmt.Sprintf(CacheKeyProductByID, id)
	_ = s.cache.Delete(ctx, cacheKey)
}

// invalidateProductListCache invalidates product list cache
func (s *productService) invalidateProductListCache(ctx context.Context) {
	_ = s.cache.DeletePattern(ctx, "product:list:*")
}
