package order

import (
	"context"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, p *Order) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *Repository) FindAll(ctx context.Context) ([]Order, error) {
	var products []Order
	err := r.db.WithContext(ctx).Find(&products).Error
	return products, err
}

func (r *Repository) FindByID(ctx context.Context, id uint) (Order, error) {
	var p Order
	err := r.db.WithContext(ctx).First(&p, id).Error
	return p, err
}

func (r *Repository) Update(ctx context.Context, p *Order) error {
	return r.db.WithContext(ctx).Save(p).Error
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Order{}, id).Error
}
