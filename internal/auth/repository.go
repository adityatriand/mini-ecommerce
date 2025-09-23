package auth

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

func (r *Repository) Create(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	return user, err
}

func (r *Repository) FindByID(ctx context.Context, id uint) (User, error) {
	var user User
	err := r.db.WithContext(ctx).First(&user, id).Error
	return user, err
}

func (r *Repository) Update(ctx context.Context, user *User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *Repository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&User{}, id).Error
}

func (r *Repository) FindAll(ctx context.Context) ([]User, error) {
	var users []User
	err := r.db.WithContext(ctx).Find(&users).Error
	return users, err
}
