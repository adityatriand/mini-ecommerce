package product

import (
	"context"
	"errors"
	"mini-e-commerce/internal/dto"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

const (
	ErrProductNotFound = "product not found"
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
	validator *validator.Validate
}

func NewService(repo Repository) Service {
	return &service{
		repo:      repo,
		validator: validator.New(),
	}
}

func (s *service) GetAllProducts(ctx context.Context) ([]Product, error) {
	return s.repo.FindAll(ctx)
}

func (s *service) GetProductByID(ctx context.Context, id uint) (*Product, error) {
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrProductNotFound)
		}
		return nil, err
	}
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
	return s.repo.Delete(ctx, id)
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

	return s.repo.Update(ctx, &product)
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

	return tx.Save(&product).Error
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

	offset := (page - 1) * pageSize

	products, total, err := s.repo.FindAllWithPagination(ctx, offset, pageSize, sortBy, order)
	if err != nil {
		return nil, err
	}

	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))

	response := &ProductListResponse{
		Data: products,
		Pagination: dto.PaginationMetadata{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	return response, nil
}
