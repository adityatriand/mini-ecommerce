package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

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
		mockRepo.On("Create", ctx, mock.AnythingOfType("*auth.User")).Return(nil)

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
		assert.Equal(t, ErrInvalidCredentials, err)
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
		assert.Equal(t, ErrInvalidCredentials, err)
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
