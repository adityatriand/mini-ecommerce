package product

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, product *Product) error
	FindAll(ctx context.Context) ([]Product, error)
	FindByID(ctx context.Context, id uint) (Product, error)
	Update(ctx context.Context, product *Product) error
	Delete(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, p *Product) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *repository) FindAll(ctx context.Context) ([]Product, error) {
	var products []Product
	err := r.db.WithContext(ctx).Find(&products).Error
	return products, err
}

func (r *repository) FindByID(ctx context.Context, id uint) (Product, error) {
	var p Product
	err := r.db.WithContext(ctx).First(&p, id).Error
	return p, err
}

func (r *repository) Update(ctx context.Context, p *Product) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Product{}, id).Error
}
