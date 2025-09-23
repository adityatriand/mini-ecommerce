package product

import (
	"context"
	"errors"
	"strconv"

	"github.com/go-playground/validator/v10"
	"gorm.io/gorm"
)

type Service struct {
	repo      *Repository
	validator *validator.Validate
}

func NewService(repo *Repository) *Service {
	return &Service{
		repo:      repo,
		validator: validator.New(),
	}
}


func (s *Service) GetAllProducts(ctx context.Context) ([]Product, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) GetProductByID(ctx context.Context, id uint) (*Product, error) {
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrProductNotFound)
		}
		return nil, err
	}
	return &product, nil
}

func (s *Service) CreateProduct(ctx context.Context, input CreateProductRequest) (*Product, error) {
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

func (s *Service) UpdateProduct(ctx context.Context, id uint, input UpdateProductRequest) (*Product, error) {
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

func (s *Service) DeleteProduct(ctx context.Context, id uint) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(ErrProductNotFound)
		}
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) ParseIDFromString(idStr string) (uint, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

const (
	ErrProductNotFound = "product not found"
)

