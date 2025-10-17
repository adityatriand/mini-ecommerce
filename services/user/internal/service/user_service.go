package service

import (
	"context"
	"errors"
	"fmt"
	"mini-e-commerce/services/user/internal/model"
	"mini-e-commerce/services/user/internal/repository"
	"mini-e-commerce/shared/cache"
	"mini-e-commerce/shared/config"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserService defines the user service interface
type UserService interface {
	RegisterUser(ctx context.Context, input model.RegisterRequest) (*model.AuthResponse, error)
	LoginUser(ctx context.Context, input model.LoginRequest) (*model.AuthResponse, error)
	GetUserByID(ctx context.Context, id uint) (*model.UserResponse, error)
	UpdateUser(ctx context.Context, id uint, input model.UpdateUserRequest) (*model.UserResponse, error)
	DeleteUser(ctx context.Context, id uint) error
	GetAllUsers(ctx context.Context, limit, offset int) ([]model.UserResponse, int64, error)
	RefreshToken(ctx context.Context, userID uint, sessionID, refreshToken string) (*model.AuthResponse, error)
	LogoutUser(ctx context.Context, userID uint, sessionID string) error
	ValidateToken(token string) (*jwt.Token, error)
}

// userService implements UserService interface
type userService struct {
	repo      repository.UserRepository
	cache     cache.Interface
	validator *validator.Validate
	logger    *zap.Logger
	jwtConfig config.JWTConfig
}

func NewUserService(repo repository.UserRepository, cache cache.Interface, logger *zap.Logger, jwtConfig config.JWTConfig) UserService {
	return &userService{
		repo:      repo,
		cache:     cache,
		validator: validator.New(),
		logger:    logger,
		jwtConfig: jwtConfig,
	}
}

// RegisterUser registers a new user
func (s *userService) RegisterUser(ctx context.Context, input model.RegisterRequest) (*model.AuthResponse, error) {
	// Validate input
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	// Check if email already exists
	_, err := s.repo.FindByEmail(ctx, input.Email)
	if err == nil {
		return nil, errors.New("email already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Hash password
	hashedPassword, err := s.hashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &model.User{
		Email:     input.Email,
		Password:  hashedPassword,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		IsActive:  true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate tokens
	authResponse, err := s.generateAuthResponse(ctx, user)
	if err != nil {
		return nil, err
	}

	s.logger.Info("User registered successfully", zap.Uint("user_id", user.ID), zap.String("email", user.Email))
	return authResponse, nil
}

// LoginUser authenticates a user
func (s *userService) LoginUser(ctx context.Context, input model.LoginRequest) (*model.AuthResponse, error) {
	// Validate input
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	// Find user by email
	user, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid credentials")
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify password
	if !s.verifyPassword(user.Password, input.Password) {
		s.logger.Warn("Invalid login attempt", zap.Uint("user_id", user.ID), zap.String("email", input.Email))
		return nil, errors.New("invalid credentials")
	}

	// Generate tokens
	authResponse, err := s.generateAuthResponse(ctx, &user)
	if err != nil {
		return nil, err
	}

	s.logger.Info("User logged in successfully", zap.Uint("user_id", user.ID), zap.String("email", user.Email))
	return authResponse, nil
}

// GetUserByID retrieves a user by ID
func (s *userService) GetUserByID(ctx context.Context, id uint) (*model.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return user.ToResponse(), nil
}

// UpdateUser updates a user
func (s *userService) UpdateUser(ctx context.Context, id uint, input model.UpdateUserRequest) (*model.UserResponse, error) {
	// Validate input
	if err := s.validator.Struct(input); err != nil {
		return nil, err
	}

	// Find user
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Update fields if provided
	if input.Email != nil && *input.Email != user.Email {
		// Check if new email already exists
		_, err := s.repo.FindByEmail(ctx, *input.Email)
		if err == nil {
			return nil, errors.New("email already exists")
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		user.Email = *input.Email
	}

	if input.FirstName != nil {
		user.FirstName = *input.FirstName
	}
	if input.LastName != nil {
		user.LastName = *input.LastName
	}
	if input.IsActive != nil {
		user.IsActive = *input.IsActive
	}

	// Save changes
	if err := s.repo.Update(ctx, &user); err != nil {
		return nil, err
	}

	s.logger.Info("User updated successfully", zap.Uint("user_id", user.ID))
	return user.ToResponse(), nil
}

// DeleteUser deletes a user
func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	// Check if user exists
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Delete user
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	s.logger.Info("User deleted successfully", zap.Uint("user_id", id))
	return nil
}

// GetAllUsers retrieves all users with pagination
func (s *userService) GetAllUsers(ctx context.Context, limit, offset int) ([]model.UserResponse, int64, error) {
	users, total, err := s.repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]model.UserResponse, len(users))
	for i, user := range users {
		responses[i] = *user.ToResponse()
	}

	return responses, total, nil
}

// RefreshToken refreshes user tokens
func (s *userService) RefreshToken(ctx context.Context, userID uint, sessionID, refreshToken string) (*model.AuthResponse, error) {
	// Validate refresh token in cache
	cacheKey := s.getSessionCacheKey(userID, sessionID)
	var cachedToken string
	if err := s.cache.Get(ctx, cacheKey, &cachedToken); err != nil {
		return nil, errors.New("invalid or expired refresh token")
	}

	if cachedToken != refreshToken {
		return nil, errors.New("invalid refresh token")
	}

	// Get user
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	// Generate new tokens
	authResponse, err := s.generateAuthResponse(ctx, &user)
	if err != nil {
		return nil, err
	}

	// Delete old session
	s.cache.Delete(ctx, cacheKey)

	s.logger.Info("Token refreshed successfully", zap.Uint("user_id", userID))
	return authResponse, nil
}

// LogoutUser logs out a user
func (s *userService) LogoutUser(ctx context.Context, userID uint, sessionID string) error {
	cacheKey := s.getSessionCacheKey(userID, sessionID)
	if err := s.cache.Delete(ctx, cacheKey); err != nil {
		return err
	}

	s.logger.Info("User logged out successfully", zap.Uint("user_id", userID))
	return nil
}

// ValidateToken validates a JWT token
func (s *userService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(s.jwtConfig.Secret), nil
	})
}

// Helper methods

// hashPassword hashes a password using bcrypt
func (s *userService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// verifyPassword verifies a password against its hash
func (s *userService) verifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// generateAuthResponse generates authentication response with tokens
func (s *userService) generateAuthResponse(ctx context.Context, user *model.User) (*model.AuthResponse, error) {
	// Generate JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"exp":     time.Now().Add(s.jwtConfig.Expiration).Unix(),
		"iat":     time.Now().Unix(),
	})

	accessToken, err := token.SignedString([]byte(s.jwtConfig.Secret))
	if err != nil {
		return nil, err
	}

	// Generate refresh token and session ID
	refreshToken := uuid.New().String()
	sessionID := uuid.New().String()

	// Store refresh token in cache
	cacheKey := s.getSessionCacheKey(user.ID, sessionID)
	if err := s.cache.Set(ctx, cacheKey, refreshToken, s.jwtConfig.RefreshExpiration); err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    sessionID,
		ExpiresIn:    int64(s.jwtConfig.Expiration.Seconds()),
	}, nil
}

// getSessionCacheKey generates cache key for session
func (s *userService) getSessionCacheKey(userID uint, sessionID string) string {
	return fmt.Sprintf("session:%d:%s", userID, sessionID)
}
