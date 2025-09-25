package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	SessionTimeout    = 3600 * time.Second
	SessionKeyPrefix  = "session:"
	MinPasswordLength = 8

	// Error constants
	ErrEmailAlreadyExists = "email already exists"
	ErrUserNotFound       = "user not found"
	ErrWeakPassword       = "password must be at least 8 characters long"
	ErrInvalidEmailFormat = "invalid email format"
	ErrPasswordRequired   = "password is required"
)

type Service interface {
	RegisterUser(ctx context.Context, input RegisterRequest) (*User, error)
	LoginUser(ctx context.Context, input LoginRequest) (*User, string, error)
	LogoutUser(ctx context.Context, sessionID string) error
	ValidateSession(ctx context.Context, sessionID string) (uint, error)
	GetUserByID(ctx context.Context, id uint) (*User, error)
	UpdateUser(ctx context.Context, id uint, input UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id uint) error
	GetAllUsers(ctx context.Context) ([]User, error)
}

type service struct {
	repo      Repository
	rdb       *redis.Client
	validator *validator.Validate
}

func NewService(repo Repository, rdb *redis.Client) Service {
	return &service{
		repo:      repo,
		rdb:       rdb,
		validator: validator.New(),
	}
}

func (s *service) RegisterUser(ctx context.Context, input RegisterRequest) (*User, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	// Check if email already exists
	_, err := s.repo.FindByEmail(ctx, input.Email)
	if err == nil {
		return nil, errors.New(ErrEmailAlreadyExists)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashed, err := HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := User{
		Email:    input.Email,
		Password: hashed,
	}

	if err := s.repo.Create(ctx, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *service) LoginUser(ctx context.Context, input LoginRequest) (*User, string, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, "", err
	}

	user, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", ErrInvalidCredentials
		}
		return nil, "", err
	}

	if !CheckPassword(user.Password, input.Password) {
		return nil, "", ErrInvalidCredentials
	}

	sessionID := uuid.New().String()
	if err := s.rdb.Set(ctx, SessionKeyPrefix+sessionID, user.ID, SessionTimeout).Err(); err != nil {
		return nil, "", err
	}

	return &user, sessionID, nil
}

func (s *service) LogoutUser(ctx context.Context, sessionID string) error {
	return s.rdb.Del(ctx, SessionKeyPrefix+sessionID).Err()
}

func (s *service) ValidateSession(ctx context.Context, sessionID string) (uint, error) {
	userIDStr, err := s.rdb.Get(ctx, SessionKeyPrefix+sessionID).Result()
	if err != nil {
		return 0, err
	}

	// Parse user ID from redis value
	userID := 0
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		return 0, err
	}

	return uint(userID), nil
}

func (s *service) GetUserByID(ctx context.Context, id uint) (*User, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrUserNotFound)
		}
		return nil, err
	}
	return &user, nil
}

func (s *service) UpdateUser(ctx context.Context, id uint, input UpdateUserRequest) (*User, error) {
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrUserNotFound)
		}
		return nil, err
	}

	if input.Email != nil && *input.Email != user.Email {
		// Check if new email already exists
		_, err := s.repo.FindByEmail(ctx, *input.Email)
		if err == nil {
			return nil, errors.New(ErrEmailAlreadyExists)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		user.Email = *input.Email
	}

	if err := s.repo.Update(ctx, &user); err != nil {
		return nil, err
	}

	return &user, nil
}

func (s *service) DeleteUser(ctx context.Context, id uint) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New(ErrUserNotFound)
		}
		return err
	}

	return s.repo.Delete(ctx, id)
}

func (s *service) GetAllUsers(ctx context.Context) ([]User, error) {
	return s.repo.FindAll(ctx)
}
