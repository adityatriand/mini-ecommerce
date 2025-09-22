package product

import (
	"context"
	"strconv"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, p *Product) error {
	return s.repo.Create(ctx, p)
}

func (s *Service) GetAll(ctx context.Context) ([]Product, error) {
	return s.repo.FindAll(ctx)
}

func (s *Service) GetByID(ctx context.Context, id uint) (Product, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, p *Product) error {
	return s.repo.Update(ctx, p)
}

func (s *Service) Delete(ctx context.Context, id uint) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) CreateProduct(ctx context.Context, input CreateProductInput) (*Product, error) {
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

func (s *Service) UpdateProduct(ctx context.Context, id uint, input UpdateProductInput) (*Product, error) {
	product, err := s.repo.FindByID(ctx, id)
	if err != nil {
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

func (s *Service) ParseIDFromString(idStr string) (uint, error) {
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}
