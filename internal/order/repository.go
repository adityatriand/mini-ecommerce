package order

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, order *Order) error
	FindAll(ctx context.Context) ([]Order, error)
	FindByID(ctx context.Context, id uint) (Order, error)
	Update(ctx context.Context, order *Order, updateFn func(*Order)) error
	Delete(ctx context.Context, id uint) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) Create(ctx context.Context, order *Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *repository) FindAll(ctx context.Context) ([]Order, error) {
	var orders []Order
	err := r.db.WithContext(ctx).Find(&orders).Error
	return orders, err
}

func (r *repository) FindByID(ctx context.Context, id uint) (Order, error) {
	var order Order
	err := r.db.WithContext(ctx).First(&order, id).Error
	return order, err
}

func (r *repository) Update(ctx context.Context, order *Order, updateFn func(*Order)) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if updateFn != nil {
			updateFn(order)
		}
		return tx.Save(order).Error
	})
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Order{}, id).Error
}
