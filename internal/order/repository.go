package order

import (
	"context"

	"gorm.io/gorm"
)

type Repository interface {
	Create(ctx context.Context, order *Order) error
	CreateWithTransaction(ctx context.Context, order *Order, txFunc func(*gorm.DB) error) error
	FindAll(ctx context.Context) ([]Order, error)
	FindAllWithPagination(ctx context.Context, offset, limit int, sortBy, order string) ([]Order, int64, error)
	FindByID(ctx context.Context, id uint) (Order, error)
	Update(ctx context.Context, order *Order, updateFn func(*Order)) error
	UpdateWithTransaction(ctx context.Context, order *Order, updateFn func(*Order), txFunc func(*gorm.DB) error) error
	Delete(ctx context.Context, id uint) error
	DeleteWithTransaction(ctx context.Context, id uint, txFunc func(*gorm.DB) error) error
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

func (r *repository) CreateWithTransaction(ctx context.Context, order *Order, txFunc func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if txFunc != nil {
			if err := txFunc(tx); err != nil {
				return err
			}
		}
		return tx.Create(order).Error
	})
}

func (r *repository) FindAll(ctx context.Context) ([]Order, error) {
	var orders []Order
	err := r.db.WithContext(ctx).Preload("OrderItems").Find(&orders).Error
	return orders, err
}

func (r *repository) FindByID(ctx context.Context, id uint) (Order, error) {
	var order Order
	err := r.db.WithContext(ctx).Preload("OrderItems").First(&order, id).Error
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

func (r *repository) UpdateWithTransaction(ctx context.Context, order *Order, updateFn func(*Order), txFunc func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if txFunc != nil {
			if err := txFunc(tx); err != nil {
				return err
			}
		}
		if updateFn != nil {
			updateFn(order)
		}
		return tx.Save(order).Error
	})
}

func (r *repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Order{}, id).Error
}

func (r *repository) DeleteWithTransaction(ctx context.Context, id uint, txFunc func(*gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if txFunc != nil {
			if err := txFunc(tx); err != nil {
				return err
			}
		}
		return tx.Delete(&Order{}, id).Error
	})
}

func (r *repository) FindAllWithPagination(ctx context.Context, offset, limit int, sortBy, order string) ([]Order, int64, error) {
	var orders []Order
	var total int64

	db := r.db.WithContext(ctx).Model(&Order{})

	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if sortBy != "" && order != "" {
		db = db.Order(sortBy + " " + order)
	} else {
		db = db.Order("created_at desc")
	}

	err := db.Preload("OrderItems").Offset(offset).Limit(limit).Find(&orders).Error
	return orders, total, err
}
