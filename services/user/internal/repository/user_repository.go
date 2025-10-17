package repository

import (
	"context"
	"mini-e-commerce/services/user/internal/model"

	"gorm.io/gorm"
)

// UserRepository defines the user repository interface
type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	FindByID(ctx context.Context, id uint) (model.User, error)
	FindByEmail(ctx context.Context, email string) (model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
	FindAll(ctx context.Context, limit, offset int) ([]model.User, int64, error)
	FindActiveUsers(ctx context.Context, limit, offset int) ([]model.User, int64, error)
}

// userRepository implements UserRepository interface
type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// FindByID finds a user by ID
func (r *userRepository) FindByID(ctx context.Context, id uint) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	return user, err
}

// FindByEmail finds a user by email
func (r *userRepository) FindByEmail(ctx context.Context, email string) (model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return user, err
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete deletes a user by ID
func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

// FindAll finds all users with pagination
func (r *userRepository) FindAll(ctx context.Context, limit, offset int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	// Count total records
	if err := r.db.WithContext(ctx).Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error

	return users, total, err
}

// FindActiveUsers finds active users with pagination
func (r *userRepository) FindActiveUsers(ctx context.Context, limit, offset int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	// Count total active users
	if err := r.db.WithContext(ctx).Model(&model.User{}).Where("is_active = ?", true).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Get paginated results
	err := r.db.WithContext(ctx).
		Where("is_active = ?", true).
		Limit(limit).
		Offset(offset).
		Order("created_at DESC").
		Find(&users).Error

	return users, total, err
}
