package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"mini-e-commerce/services/user/internal/model"
)

// Type aliases for the test
type User = model.User
type RegisterRequest = model.RegisterRequest
type LoginRequest = model.LoginRequest
type UserResponse = model.UserResponse
type AuthResponse = model.AuthResponse
type UpdateUserRequest = model.UpdateUserRequest

// UserClaims for JWT
type UserClaims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
}

// Error constants
var (
	ErrEmailAlreadyExists = "email already exists"
	ErrInvalidCredentials = "invalid credentials"
	ErrUserNotFound       = "user not found"
)

// Mock service interface
type Service interface {
	RegisterUser(ctx context.Context, input RegisterRequest) (*UserResponse, error)
	LoginUser(ctx context.Context, input LoginRequest) (*AuthResponse, error)
	RefreshToken(ctx context.Context, userID uint, sessionID, refreshToken string) (*AuthResponse, error)
	LogoutUser(ctx context.Context, userID uint, sessionID string) error
	GetUserByID(ctx context.Context, userID uint) (*UserResponse, error)
}

// Mock service implementation
type mockService struct {
	repo    *MockRepository
	jwt     *MockJWTManager
	session *MockSessionManager
	logger  *zap.Logger
}

func NewService(repo *MockRepository, jwt *MockJWTManager, session *MockSessionManager, logger *zap.Logger, accessTTL, refreshTTL time.Duration) Service {
	return &mockService{
		repo:    repo,
		jwt:     jwt,
		session: session,
		logger:  logger,
	}
}

func (s *mockService) RegisterUser(ctx context.Context, input RegisterRequest) (*UserResponse, error) {
	// Check if user already exists
	_, err := s.repo.FindByEmail(ctx, input.Email)
	if err == nil {
		return nil, errors.New(ErrEmailAlreadyExists)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	// Create new user
	hashedPassword, _ := HashPassword(input.Password)
	user := &User{
		Email:     input.Email,
		Password:  hashedPassword,
		FirstName: input.FirstName,
		LastName:  input.LastName,
		IsActive:  true,
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user.ToResponse(), nil
}

func (s *mockService) LoginUser(ctx context.Context, input LoginRequest) (*AuthResponse, error) {
	user, err := s.repo.FindByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrInvalidCredentials)
		}
		return nil, err
	}

	if !CheckPassword(user.Password, input.Password) {
		return nil, errors.New(ErrInvalidCredentials)
	}

	accessToken, err := s.jwt.Generate(user.ID)
	if err != nil {
		return nil, err
	}

	sessionID := "session-" + time.Now().Format("20060102150405")
	refreshToken := "refresh-" + time.Now().Format("20060102150405")

	err = s.session.StoreRefreshToken(ctx, user.ID, sessionID, refreshToken, 7*24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    sessionID,
	}, nil
}

func (s *mockService) RefreshToken(ctx context.Context, userID uint, sessionID, refreshToken string) (*AuthResponse, error) {
	err := s.session.ValidateRefreshToken(ctx, userID, sessionID, refreshToken)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.jwt.Generate(userID)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User:         &user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		SessionID:    sessionID,
	}, nil
}

func (s *mockService) LogoutUser(ctx context.Context, userID uint, sessionID string) error {
	return s.session.DeleteRefreshToken(ctx, userID, sessionID)
}

func (s *mockService) GetUserByID(ctx context.Context, userID uint) (*UserResponse, error) {
	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(ErrUserNotFound)
		}
		return nil, err
	}

	return user.ToResponse(), nil
}

// Helper functions
func HashPassword(password string) (string, error) {
	// Simple mock implementation
	return "hashed-" + password, nil
}

func CheckPassword(hashedPassword, password string) bool {
	// Simple mock implementation
	return hashedPassword == "hashed-"+password
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(User), args.Error(1)
}

func (m *MockRepository) FindByID(ctx context.Context, id uint) (User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(User), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, user *User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) FindAll(ctx context.Context) ([]User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]User), args.Error(1)
}

type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) Generate(userID uint) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) Verify(tokenStr string) (*UserClaims, error) {
	args := m.Called(tokenStr)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*UserClaims), args.Error(1)
}

type MockSessionManager struct {
	mock.Mock
}

func (m *MockSessionManager) StoreRefreshToken(ctx context.Context, userID uint, sessionID, token string, ttl time.Duration) error {
	args := m.Called(ctx, userID, sessionID, token, ttl)
	return args.Error(0)
}

func (m *MockSessionManager) ValidateRefreshToken(ctx context.Context, userID uint, sessionID, token string) error {
	args := m.Called(ctx, userID, sessionID, token)
	return args.Error(0)
}

func (m *MockSessionManager) DeleteRefreshToken(ctx context.Context, userID uint, sessionID string) error {
	args := m.Called(ctx, userID, sessionID)
	return args.Error(0)
}

func (m *MockSessionManager) GetSessionKey(userID uint, sessionID string) string {
	args := m.Called(userID, sessionID)
	return args.String(0)
}

func TestService_RegisterUser(t *testing.T) {
	ctx := context.Background()

	t.Run("should register user successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockJWT := new(MockJWTManager)
		mockSession := new(MockSessionManager)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockJWT, mockSession, logger, time.Hour, 7*24*time.Hour)

		input := RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		mockRepo.On("FindByEmail", ctx, input.Email).Return(User{}, gorm.ErrRecordNotFound)
		mockRepo.On("Create", ctx, mock.AnythingOfType("*model.User")).Return(nil)

		user, err := service.RegisterUser(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, input.Email, user.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when email already exists", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockJWT := new(MockJWTManager)
		mockSession := new(MockSessionManager)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockJWT, mockSession, logger, time.Hour, 7*24*time.Hour)

		input := RegisterRequest{
			Email:    "existing@example.com",
			Password: "password123",
		}

		existingUser := User{ID: 1, Email: input.Email}
		mockRepo.On("FindByEmail", ctx, input.Email).Return(existingUser, nil)

		user, err := service.RegisterUser(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrEmailAlreadyExists, err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestService_LoginUser(t *testing.T) {
	ctx := context.Background()

	t.Run("should login user successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockJWT := new(MockJWTManager)
		mockSession := new(MockSessionManager)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockJWT, mockSession, logger, time.Hour, 7*24*time.Hour)

		hashedPassword, _ := HashPassword("password123")
		user := User{
			ID:       1,
			Email:    "test@example.com",
			Password: hashedPassword,
		}

		input := LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		mockRepo.On("FindByEmail", ctx, input.Email).Return(user, nil)
		mockJWT.On("Generate", user.ID).Return("access-token", nil)
		mockSession.On("StoreRefreshToken", ctx, user.ID, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("time.Duration")).Return(nil)

		authResp, err := service.LoginUser(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, authResp)
		assert.Equal(t, user.Email, authResp.User.Email)
		assert.Equal(t, "access-token", authResp.AccessToken)
		assert.NotEmpty(t, authResp.RefreshToken)
		assert.NotEmpty(t, authResp.SessionID)
		mockRepo.AssertExpectations(t)
		mockJWT.AssertExpectations(t)
		mockSession.AssertExpectations(t)
	})

	t.Run("should return error for non-existent user", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockJWT := new(MockJWTManager)
		mockSession := new(MockSessionManager)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockJWT, mockSession, logger, time.Hour, 7*24*time.Hour)

		input := LoginRequest{
			Email:    "nonexistent@example.com",
			Password: "password123",
		}

		mockRepo.On("FindByEmail", ctx, input.Email).Return(User{}, gorm.ErrRecordNotFound)

		authResp, err := service.LoginUser(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, authResp)
		assert.Equal(t, ErrInvalidCredentials, err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error for incorrect password", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockJWT := new(MockJWTManager)
		mockSession := new(MockSessionManager)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockJWT, mockSession, logger, time.Hour, 7*24*time.Hour)

		hashedPassword, _ := HashPassword("correct-password")
		user := User{
			ID:       1,
			Email:    "test@example.com",
			Password: hashedPassword,
		}

		input := LoginRequest{
			Email:    "test@example.com",
			Password: "wrong-password",
		}

		mockRepo.On("FindByEmail", ctx, input.Email).Return(user, nil)

		authResp, err := service.LoginUser(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, authResp)
		assert.Equal(t, ErrInvalidCredentials, err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestService_RefreshToken(t *testing.T) {
	ctx := context.Background()

	t.Run("should refresh token successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockJWT := new(MockJWTManager)
		mockSession := new(MockSessionManager)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockJWT, mockSession, logger, time.Hour, 7*24*time.Hour)

		userID := uint(1)
		sessionID := "session-123"
		refreshToken := "refresh-token"

		user := User{
			ID:    userID,
			Email: "test@example.com",
		}

		mockSession.On("ValidateRefreshToken", ctx, userID, sessionID, refreshToken).Return(nil)
		mockRepo.On("FindByID", ctx, userID).Return(user, nil)
		mockJWT.On("Generate", userID).Return("new-access-token", nil)

		authResp, err := service.RefreshToken(ctx, userID, sessionID, refreshToken)

		require.NoError(t, err)
		assert.NotNil(t, authResp)
		assert.Equal(t, "new-access-token", authResp.AccessToken)
		mockSession.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
		mockJWT.AssertExpectations(t)
	})
}

func TestService_LogoutUser(t *testing.T) {
	ctx := context.Background()

	t.Run("should logout user successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockJWT := new(MockJWTManager)
		mockSession := new(MockSessionManager)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockJWT, mockSession, logger, time.Hour, 7*24*time.Hour)

		userID := uint(1)
		sessionID := "session-123"

		mockSession.On("DeleteRefreshToken", ctx, userID, sessionID).Return(nil)

		err := service.LogoutUser(ctx, userID, sessionID)

		require.NoError(t, err)
		mockSession.AssertExpectations(t)
	})
}

func TestService_GetUserByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get user by ID successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockJWT := new(MockJWTManager)
		mockSession := new(MockSessionManager)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockJWT, mockSession, logger, time.Hour, 7*24*time.Hour)

		userID := uint(1)
		expectedUser := User{
			ID:    userID,
			Email: "test@example.com",
		}

		mockRepo.On("FindByID", ctx, userID).Return(expectedUser, nil)

		user, err := service.GetUserByID(ctx, userID)

		require.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, expectedUser.Email, user.Email)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockJWT := new(MockJWTManager)
		mockSession := new(MockSessionManager)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockJWT, mockSession, logger, time.Hour, 7*24*time.Hour)

		userID := uint(999)

		mockRepo.On("FindByID", ctx, userID).Return(User{}, gorm.ErrRecordNotFound)

		user, err := service.GetUserByID(ctx, userID)

		assert.Error(t, err)
		assert.Nil(t, user)
		assert.Equal(t, ErrUserNotFound, err.Error())
		mockRepo.AssertExpectations(t)
	})
}