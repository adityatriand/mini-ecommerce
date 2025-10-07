package product

import (
	"context"
	"errors"
	"fmt"
	"mini-e-commerce/internal/cache"
	"mini-e-commerce/internal/dto"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	ErrProductNotFound  = "product not found"
	CacheKeyProductByID = "product:id:%d"
	CacheKeyProductList = "product:list:%d:%d:%s:%s" // page:pageSize:sortBy:order
	CacheTTLProduct     = 5 * time.Minute
	CacheTTLProductList = 2 * time.Minute
)

type Service interface {
	CreateProduct(ctx context.Context, input CreateProductRequest) (*Product, error)
	GetAllProducts(ctx context.Context) ([]Product, error)
	GetAllProductsWithQuery(ctx context.Context, query ProductQuery) (*ProductListResponse, error)
	GetProductByID(ctx context.Context, id uint) (*Product, error)
	UpdateProduct(ctx context.Context, id uint, input UpdateProductRequest) (*Product, error)
	DeleteProduct(ctx context.Context, id uint) error
	UpdateStock(ctx context.Context, id uint, stockDelta int) error
	UpdateStockWithTx(tx *gorm.DB, id uint, stockDelta int) error
}
type service struct {
	repo      Repository
	cache     *cache.RedisCache
	validator *validator.Validate
	logger    *zap.Logger
}

func NewService(repo Repository, cache *cache.RedisCache, logger *zap.Logger) Service {
	return &service{
		repo:      repo,
		cache:     cache,
		validator: validator.New(),
		logger:    logger,
	}
}

func (s *service) invalidateProductCache(ctx context.Context, id uint) {
	cacheKey := fmt.Sprintf(CacheKeyProductByID, id)
	_ = s.cache.Delete(ctx, cacheKey)
	_ = s.cache.DeletePattern(ctx, "product:list:*")
}

func (s *service) invalidateProductListCache(ctx context.Context) {
	_ = s.cache.DeletePattern(ctx, "product:list:*")
}

func (s *service) GetAllProducts(ctx context.Context) ([]Product, error) {
	return s.repo.FindAll(ctx)
}

func (s *service) GetProductByID(ctx context.Context, id uint) (*Product, error) {
	cacheKey := fmt.Sprintf(CacheKeyProductByID, id)
	var product Product
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

	product, err = s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrProductNotFound)
		}
		return nil, err
	}

	_ = s.cache.Set(ctx, cacheKey, product, CacheTTLProduct)

	return &product, nil
}

func (s *service) CreateProduct(ctx context.Context, input CreateProductRequest) (*Product, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	product := Product{
		Name:  input.Name,
		Price: input.Price,
		Stock: input.Stock,
	}
	if err := s.repo.Create(ctx, &product); err != nil {
		return nil, err
	}

	s.invalidateProductListCache(ctx)

	return &product, nil
}

func (s *service) UpdateProduct(ctx context.Context, id uint, input UpdateProductRequest) (*Product, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrProductNotFound)
		}
		return nil, err
	}

	if input.Name != nil {
		product.Name = *input.Name
	}
	if input.Price != nil {
		product.Price = *input.Price
	}
	if input.Stock != nil {
		product.Stock = *input.Stock
	}
	if err := s.repo.Update(ctx, &product); err != nil {
		return nil, err
	}

	s.invalidateProductCache(ctx, id)

	return &product, nil
}

func (s *service) DeleteProduct(ctx context.Context, id uint) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(ErrProductNotFound)
		}
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.invalidateProductCache(ctx, id)

	return nil
}

func (s *service) UpdateStock(ctx context.Context, id uint, stockDelta int) error {
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(ErrProductNotFound)
		}
		return err
	}

	product.Stock += stockDelta
	if product.Stock < 0 {
		return errors.New("insufficient stock")
	}

	if err := s.repo.Update(ctx, &product); err != nil {
		return err
	}

	s.invalidateProductCache(ctx, id)

	return nil
}

func (s *service) UpdateStockWithTx(tx *gorm.DB, id uint, stockDelta int) error {
	var product Product
	if err := tx.First(&product, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(ErrProductNotFound)
		}
		return err
	}

	product.Stock += stockDelta
	if product.Stock < 0 {
		return errors.New("insufficient stock")
	}

	if err := tx.Save(&product).Error; err != nil {
		return err
	}

	s.invalidateProductCache(context.Background(), id)

	return nil
}

func (s *service) GetAllProductsWithQuery(ctx context.Context, query ProductQuery) (*ProductListResponse, error) {
	page := query.Page
	if page <= 0 {
		page = 1
	}

	pageSize := query.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}

	if pageSize > 100 {
		pageSize = 100
	}

	order := query.Order
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	sortBy := query.SortBy
	validSortFields := map[string]bool{
		"id": true, "name": true, "price": true, "stock": true, "created_at": true,
	}
	if sortBy != "" && !validSortFields[sortBy] {
		sortBy = "created_at"
	}

	cacheKey := fmt.Sprintf(CacheKeyProductList, page, pageSize, sortBy, order)
	var response ProductListResponse
	err := s.cache.Get(ctx, cacheKey, &response)
	if err == nil {
		return &response, nil
	}

	if err != redis.Nil {
		s.logger.Warn("Cache error on GetAllProductsWithQuery, falling back to database",
			zap.Int("page", page),
			zap.Int("page_size", pageSize),
			zap.Error(err),
		)
	}

	offset := (page - 1) * pageSize

	products, total, err := s.repo.FindAllWithPagination(ctx, offset, pageSize, sortBy, order)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	response = ProductListResponse{
		Data: products,
		Pagination: dto.PaginationMetadata{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	_ = s.cache.Set(ctx, cacheKey, response, CacheTTLProductList)

	return &response, nil
}
