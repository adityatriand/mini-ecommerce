package order

import (
	"context"

	"mini-e-commerce/internal/product"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, order *Order) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *Repository) FindAll(ctx context.Context) ([]Order, error) {
	var orders []Order
	err := r.db.WithContext(ctx).Find(&orders).Error
	return orders, err
}

func (r *Repository) FindByID(ctx context.Context, id uint) (Order, error) {
	var order Order
	err := r.db.WithContext(ctx).First(&order, id).Error
	return order, err
}

func (r *Repository) Update(ctx context.Context, order *Order, product *product.Product, updateFn func(*Order)) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if updateFn != nil {
			updateFn(order)
		}
		if err := tx.Save(order).Error; err != nil {
			return err
		}
		if product != nil {
			if err := tx.Save(product).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Order{}, id).Error
}
