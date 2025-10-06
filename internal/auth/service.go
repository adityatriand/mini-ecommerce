package auth

import (
	"context"
	"errors"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
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
	LoginUser(ctx context.Context, input LoginRequest) (*AuthResponse, error)
	RefreshToken(ctx context.Context, userID uint, sessionID, refreshToken string) (*AuthResponse, error)
	LogoutUser(ctx context.Context, userID uint, sessionID string) error
	GetUserByID(ctx context.Context, id uint) (*User, error)
	UpdateUser(ctx context.Context, id uint, input UpdateUserRequest) (*User, error)
	DeleteUser(ctx context.Context, id uint) error
	GetAllUsers(ctx context.Context) ([]User, error)
}

type service struct {
	repo           Repository
	jwtManager     *JWTManager
	sessionManager *SessionManager
	validator      *validator.Validate
	logger         *zap.Logger
	jwtExpiration  time.Duration
	refreshExp     time.Duration
}

func NewService(repo Repository, jwtManager *JWTManager, sessionManager *SessionManager, logger *zap.Logger, jwtExp, refreshExp time.Duration) Service {
	return &service{
		repo:           repo,
		jwtManager:     jwtManager,
		sessionManager: sessionManager,
		validator:      validator.New(),
		logger:         logger,
		jwtExpiration:  jwtExp,
		refreshExp:     refreshExp,
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

func (s *service) LoginUser(ctx context.Context, input LoginRequest) (*AuthResponse, error) {
	if err := s.validator.Struct(input); err != nil {
		s.logger.Warn("Login validation failed", zap.Error(err))
		return nil, err
	}

	user, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			s.logger.Warn("Login attempt with non-existent email", zap.String("email", input.Email))
			return nil, ErrInvalidCredentials
		}
		s.logger.Error("Failed to find user by email", zap.Error(err))
		return nil, err
	}

	if !CheckPassword(user.Password, input.Password) {
		s.logger.Warn("Invalid password attempt", zap.Uint("user_id", user.ID))
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.jwtManager.Generate(user.ID)
	if err != nil {
		s.logger.Error("Failed to generate access token", zap.Error(err), zap.Uint("user_id", user.ID))
		return nil, err
	}

	sessionID := uuid.New().String()
	refreshToken := uuid.New().String()

	if err := s.sessionManager.StoreRefreshToken(ctx, user.ID, sessionID, refreshToken, s.refreshExp); err != nil {
		s.logger.Error("Failed to store refresh token", zap.Error(err), zap.Uint("user_id", user.ID))
		return nil, err
	}

	s.logger.Info("User logged in successfully",
		zap.Uint("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("session_id", sessionID),
	)

	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    sessionID,
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, userID uint, sessionID, refreshToken string) (*AuthResponse, error) {
	if err := s.sessionManager.ValidateRefreshToken(ctx, userID, sessionID, refreshToken); err != nil {
		s.logger.Warn("Invalid refresh token attempt",
			zap.Error(err),
			zap.Uint("user_id", userID),
			zap.String("session_id", sessionID),
		)
		return nil, errors.New("invalid or expired refresh token")
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to find user during token refresh", zap.Error(err), zap.Uint("user_id", userID))
		return nil, errors.New(ErrUserNotFound)
	}

	newAccessToken, err := s.jwtManager.Generate(user.ID)
	if err != nil {
		s.logger.Error("Failed to generate new access token", zap.Error(err), zap.Uint("user_id", user.ID))
		return nil, err
	}

	s.logger.Info("Access token refreshed successfully",
		zap.Uint("user_id", user.ID),
		zap.String("session_id", sessionID),
	)

	return &AuthResponse{
		User:         user,
		AccessToken:  newAccessToken,
		RefreshToken: refreshToken,
		SessionID:    sessionID,
	}, nil
}

func (s *service) LogoutUser(ctx context.Context, userID uint, sessionID string) error {
	if err := s.sessionManager.DeleteRefreshToken(ctx, userID, sessionID); err != nil {
		s.logger.Error("Failed to delete refresh token", zap.Error(err), zap.Uint("user_id", userID))
		return err
	}

	s.logger.Info("User logged out successfully", zap.Uint("user_id", userID), zap.String("session_id", sessionID))
	return nil
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
