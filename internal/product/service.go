package product

import (
	"context"
	"errors"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

const (
	ErrProductNotFound = "product not found"
)

type Service interface {
	CreateProduct(ctx context.Context, input CreateProductRequest) (*Product, error)
	GetAllProducts(ctx context.Context) ([]Product, error)
	GetProductByID(ctx context.Context, id uint) (*Product, error)
	UpdateProduct(ctx context.Context, id uint, input UpdateProductRequest) (*Product, error)
	DeleteProduct(ctx context.Context, id uint) error
	UpdateStock(ctx context.Context, id uint, stockDelta int) error
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
