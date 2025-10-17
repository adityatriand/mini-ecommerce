package repository

import (
	"context"
	"fmt"
	"mini-e-commerce/services/order/internal/model"
	"strings"

	"gorm.io/gorm"
)

// OrderRepository defines the order repository interface
type OrderRepository interface {
	Create(ctx context.Context, order *model.Order) error
	CreateWithTransaction(ctx context.Context, order *model.Order, fn func(tx *gorm.DB) error) error
	FindByID(ctx context.Context, id uint) (model.Order, error)
	FindByUserID(ctx context.Context, userID uint, query model.OrderQuery) ([]model.Order, int64, error)
	FindAll(ctx context.Context, query model.OrderQuery) ([]model.Order, int64, error)
	Update(ctx context.Context, order *model.Order) error
	UpdateWithTransaction(ctx context.Context, order *model.Order, updateFn func(*model.Order), fn func(tx *gorm.DB) error) error
	Delete(ctx context.Context, id uint) error
	DeleteWithTransaction(ctx context.Context, id uint, fn func(tx *gorm.DB) error) error
}

// orderRepository implements OrderRepository interface
type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{
		db: db,
	}
}

func (r *orderRepository) Create(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) CreateWithTransaction(ctx context.Context, order *model.Order, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create order
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// Execute additional transaction logic
		if fn != nil {
			return fn(tx)
		}

		return nil
	})
}

// FindByID finds an order by ID
func (r *orderRepository) FindByID(ctx context.Context, id uint) (model.Order, error) {
	var order model.Order
	err := r.db.WithContext(ctx).Preload("OrderItems").First(&order, id).Error
	return order, err
}

// FindByUserID finds orders by user ID with filtering and pagination
func (r *orderRepository) FindByUserID(ctx context.Context, userID uint, query model.OrderQuery) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	// Build query
	db := r.db.WithContext(ctx).Model(&model.Order{}).Where("user_id = ?", userID)

	// Apply filters
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderBy := r.buildOrderBy(query.SortBy, query.Order)
	db = db.Order(orderBy)

	// Apply pagination
	if query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		db = db.Limit(query.PageSize).Offset(offset)
	}

	// Execute query with preload
	err := db.Preload("OrderItems").Find(&orders).Error
	return orders, total, err
}

// FindAll finds all orders with filtering and pagination
func (r *orderRepository) FindAll(ctx context.Context, query model.OrderQuery) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64

	// Build query
	db := r.db.WithContext(ctx).Model(&model.Order{})

	// Apply filters
	if query.Status != "" {
		db = db.Where("status = ?", query.Status)
	}
	if query.UserID != nil {
		db = db.Where("user_id = ?", *query.UserID)
	}

	// Count total records
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply sorting
	orderBy := r.buildOrderBy(query.SortBy, query.Order)
	db = db.Order(orderBy)

	// Apply pagination
	if query.Page > 0 && query.PageSize > 0 {
		offset := (query.Page - 1) * query.PageSize
		db = db.Limit(query.PageSize).Offset(offset)
	}

	// Execute query with preload
	err := db.Preload("OrderItems").Find(&orders).Error
	return orders, total, err
}

// Update updates an order
func (r *orderRepository) Update(ctx context.Context, order *model.Order) error {
	return r.db.WithContext(ctx).Save(order).Error
}

// UpdateWithTransaction updates an order within a transaction
func (r *orderRepository) UpdateWithTransaction(ctx context.Context, order *model.Order, updateFn func(*model.Order), fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Apply update function
		if updateFn != nil {
			updateFn(order)
		}

		// Update order
		if err := tx.Save(order).Error; err != nil {
			return err
		}

		// Execute additional transaction logic
		if fn != nil {
			return fn(tx)
		}

		return nil
	})
}

// Delete deletes an order by ID
func (r *orderRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Order{}, id).Error
}

// DeleteWithTransaction deletes an order within a transaction
func (r *orderRepository) DeleteWithTransaction(ctx context.Context, id uint, fn func(tx *gorm.DB) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Execute additional transaction logic first
		if fn != nil {
			if err := fn(tx); err != nil {
				return err
			}
		}

		// Delete order
		return tx.Delete(&model.Order{}, id).Error
	})
}

// buildOrderBy builds the ORDER BY clause
func (r *orderRepository) buildOrderBy(sortBy, order string) string {
	// Default sorting
	if sortBy == "" {
		sortBy = "created_at"
	}
	if order == "" {
		order = "desc"
	}

	// Validate sortBy field
	validFields := map[string]bool{
		"id": true, "user_id": true, "total_price": true, "status": true, "created_at": true,
	}
	if !validFields[sortBy] {
		sortBy = "created_at"
	}

	// Validate order
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	return fmt.Sprintf("%s %s", sortBy, strings.ToUpper(order))
}
