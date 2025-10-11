package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"mini-e-commerce/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) RegisterUser(ctx context.Context, input RegisterRequest) (*User, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) LoginUser(ctx context.Context, input LoginRequest) (*AuthResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResponse), args.Error(1)
}

func (m *MockService) RefreshToken(ctx context.Context, userID uint, sessionID, refreshToken string) (*AuthResponse, error) {
	args := m.Called(ctx, userID, sessionID, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*AuthResponse), args.Error(1)
}

func (m *MockService) LogoutUser(ctx context.Context, userID uint, sessionID string) error {
	args := m.Called(ctx, userID, sessionID)
	return args.Error(0)
}

func (m *MockService) GetUserByID(ctx context.Context, id uint) (*User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) UpdateUser(ctx context.Context, id uint, input UpdateUserRequest) (*User, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockService) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockService) GetAllUsers(ctx context.Context) ([]User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]User), args.Error(1)
}

func setupLogger() logger.Logger {
	logConfig := &logger.Config{
		ServiceName: "test",
		AppVersion:  "test",
		LogLevel:    zapcore.FatalLevel,
		Mode:        "development",
	}
	log, _ := logger.NewLogger(logConfig)
	return log
}

func TestHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should register user successfully", func(t *testing.T) {
		mockService := new(MockService)
		log := setupLogger()
		handler := NewHandler(mockService, log)

		input := RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		expectedUser := &User{
			ID:    1,
			Email: input.Email,
		}

		mockService.On("RegisterUser", mock.Anything, input).Return(expectedUser, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body, _ := json.Marshal(input)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Register(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid JSON", func(t *testing.T) {
		mockService := new(MockService)
		log := setupLogger()
		handler := NewHandler(mockService, log)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Register(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error when email already exists", func(t *testing.T) {
		mockService := new(MockService)
		log := setupLogger()
		handler := NewHandler(mockService, log)

		input := RegisterRequest{
			Email:    "existing@example.com",
			Password: "password123",
		}

		mockService.On("RegisterUser", mock.Anything, input).Return(nil, errors.New(ErrEmailAlreadyExists))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body, _ := json.Marshal(input)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Register(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should login user successfully", func(t *testing.T) {
		mockService := new(MockService)
		log := setupLogger()
		handler := NewHandler(mockService, log)

		input := LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		authResp := &AuthResponse{
			User: User{
				ID:    1,
				Email: input.Email,
			},
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			SessionID:    "session-id",
		}

		mockService.On("LoginUser", mock.Anything, input).Return(authResp, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body, _ := json.Marshal(input)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Login(c)

		assert.Equal(t, http.StatusOK, w.Code)

		cookies := w.Result().Cookies()
		require.NotEmpty(t, cookies)

		cookieNames := make(map[string]bool)
		for _, cookie := range cookies {
			cookieNames[cookie.Name] = true
		}
		assert.True(t, cookieNames["session_id"])
		assert.True(t, cookieNames["refresh_token"])
		assert.True(t, cookieNames["user_id"])

		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid credentials", func(t *testing.T) {
		mockService := new(MockService)
		log := setupLogger()
		handler := NewHandler(mockService, log)

		input := LoginRequest{
			Email:    "test@example.com",
			Password: "wrong-password",
		}

		mockService.On("LoginUser", mock.Anything, input).Return(nil, ErrInvalidCredentials)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body, _ := json.Marshal(input)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.Login(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should logout user successfully", func(t *testing.T) {
		mockService := new(MockService)
		log := setupLogger()
		handler := NewHandler(mockService, log)

		mockService.On("LogoutUser", mock.Anything, uint(1), "session-123").Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
		c.Request.AddCookie(&http.Cookie{Name: "session_id", Value: "session-123"})
		c.Request.AddCookie(&http.Cookie{Name: "user_id", Value: "1"})

		handler.Logout(c)

		assert.Equal(t, http.StatusOK, w.Code)

		cookies := w.Result().Cookies()
		require.NotEmpty(t, cookies)

		for _, cookie := range cookies {
			if cookie.Name == "session_id" || cookie.Name == "refresh_token" || cookie.Name == "user_id" {
				assert.Equal(t, -1, cookie.MaxAge)
				assert.Equal(t, "", cookie.Value)
			}
		}

		mockService.AssertExpectations(t)
	})

	t.Run("should return error when session_id cookie is missing", func(t *testing.T) {
		mockService := new(MockService)
		log := setupLogger()
		handler := NewHandler(mockService, log)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodPost, "/auth/logout", nil)

		handler.Logout(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
