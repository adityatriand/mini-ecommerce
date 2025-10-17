package repository

import (
	"context"
	"fmt"
	"mini-e-commerce/services/product/internal/model"
	"strings"

	"gorm.io/gorm"
)

// ProductRepository defines the product repository interface
type ProductRepository interface {
	Create(ctx context.Context, product *model.Product) error
	FindByID(ctx context.Context, id uint) (model.Product, error)
	FindBySKU(ctx context.Context, sku string) (model.Product, error)
	Update(ctx context.Context, product *model.Product) error
	Delete(ctx context.Context, id uint) error
	FindAll(ctx context.Context, query model.ProductQuery) ([]model.Product, int64, error)
	FindMultipleByIDs(ctx context.Context, ids []uint) ([]model.Product, error)
	UpdateStock(ctx context.Context, id uint, stockDelta int) error
	UpdateStockOptimizedWithTx(tx *gorm.DB, id uint, stockDelta int) error
}

// productRepository implements ProductRepository interface
type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &productRepository{
		db: db,
	}
}

func (r *productRepository) Create(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

// FindByID finds a product by ID
func (r *productRepository) FindByID(ctx context.Context, id uint) (model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).First(&product, id).Error
	return product, err
}

// FindBySKU finds a product by SKU
func (r *productRepository) FindBySKU(ctx context.Context, sku string) (model.Product, error) {
	var product model.Product
	err := r.db.WithContext(ctx).Where("sku = ?", sku).First(&product).Error
	return product, err
}

// Update updates a product
func (r *productRepository) Update(ctx context.Context, product *model.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

// Delete deletes a product by ID
func (r *productRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Product{}, id).Error
}

// FindAll finds products with filtering and pagination
func (r *productRepository) FindAll(ctx context.Context, query model.ProductQuery) ([]model.Product, int64, error) {
	var products []model.Product
	var total int64

	// Build query
	db := r.db.WithContext(ctx).Model(&model.Product{})

	// Apply filters
	if query.Category != "" {
		db = db.Where("category = ?", query.Category)
	}
	if query.Search != "" {
		searchTerm := "%" + query.Search + "%"
		db = db.Where("name ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
	}
	if query.IsActive != nil {
		db = db.Where("is_active = ?", *query.IsActive)
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

	// Execute query
	err := db.Find(&products).Error
	return products, total, err
}

// FindMultipleByIDs finds multiple products by IDs
func (r *productRepository) FindMultipleByIDs(ctx context.Context, ids []uint) ([]model.Product, error) {
	var products []model.Product
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&products).Error
	return products, err
}

// UpdateStock updates product stock
func (r *productRepository) UpdateStock(ctx context.Context, id uint, stockDelta int) error {
	return r.db.WithContext(ctx).Model(&model.Product{}).
		Where("id = ?", id).
		Update("stock", gorm.Expr("stock + ?", stockDelta)).Error
}

// UpdateStockOptimizedWithTx updates product stock within a transaction
func (r *productRepository) UpdateStockOptimizedWithTx(tx *gorm.DB, id uint, stockDelta int) error {
	// First, get current stock to check if it would go negative
	var currentStock int
	if err := tx.Model(&model.Product{}).Select("stock").Where("id = ?", id).Scan(&currentStock).Error; err != nil {
		return err
	}

	if currentStock+stockDelta < 0 {
		return fmt.Errorf("insufficient stock: current=%d, requested=%d", currentStock, -stockDelta)
	}

	// Update stock
	return tx.Model(&model.Product{}).
		Where("id = ?", id).
		Update("stock", gorm.Expr("stock + ?", stockDelta)).Error
}

// buildOrderBy builds the ORDER BY clause
func (r *productRepository) buildOrderBy(sortBy, order string) string {
	// Default sorting
	if sortBy == "" {
		sortBy = "created_at"
	}
	if order == "" {
		order = "desc"
	}

	// Validate sortBy field
	validFields := map[string]bool{
		"id": true, "name": true, "price": true, "stock": true, "created_at": true,
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
