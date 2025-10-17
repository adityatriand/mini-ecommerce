package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"mini-e-commerce/services/user/internal/handler"
	"mini-e-commerce/services/user/internal/model"
	"mini-e-commerce/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) RegisterUser(ctx context.Context, input model.RegisterRequest) (*model.AuthResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuthResponse), args.Error(1)
}

func (m *MockService) LoginUser(ctx context.Context, input model.LoginRequest) (*model.AuthResponse, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuthResponse), args.Error(1)
}

func (m *MockService) RefreshToken(ctx context.Context, userID uint, sessionID, refreshToken string) (*model.AuthResponse, error) {
	args := m.Called(ctx, userID, sessionID, refreshToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AuthResponse), args.Error(1)
}

func (m *MockService) LogoutUser(ctx context.Context, userID uint, sessionID string) error {
	args := m.Called(ctx, userID, sessionID)
	return args.Error(0)
}

func (m *MockService) GetUserByID(ctx context.Context, id uint) (*model.UserResponse, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserResponse), args.Error(1)
}

func (m *MockService) UpdateUser(ctx context.Context, id uint, input model.UpdateUserRequest) (*model.UserResponse, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.UserResponse), args.Error(1)
}

func (m *MockService) DeleteUser(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockService) GetAllUsers(ctx context.Context, limit, offset int) ([]model.UserResponse, int64, error) {
	args := m.Called(ctx, limit, offset)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.UserResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockService) ValidateToken(token string) (*jwt.Token, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*jwt.Token), args.Error(1)
}

func setupLogger() *zap.Logger {
	logConfig := logger.Config{
		Level:      "info",
		Format:     "console",
		OutputPath: "",
	}
	log, _ := logger.NewLogger(logConfig, "test")
	return log.GetZapLogger()
}

func TestHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should register user successfully", func(t *testing.T) {
		mockService := new(MockService)
		log := setupLogger()
		handler := handler.NewUserHandler(mockService, log)

		input := model.RegisterRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		expectedAuthResp := &model.AuthResponse{
			User: &model.User{
				ID:    1,
				Email: input.Email,
			},
			AccessToken:  "access-token",
			RefreshToken: "refresh-token",
			SessionID:    "session-id",
		}

		mockService.On("RegisterUser", mock.Anything, input).Return(expectedAuthResp, nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body, _ := json.Marshal(input)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.RegisterUser(c)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid JSON", func(t *testing.T) {
		mockService := new(MockService)
		log := setupLogger()
		handler := handler.NewUserHandler(mockService, log)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("invalid json"))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.RegisterUser(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error when email already exists", func(t *testing.T) {
		mockService := new(MockService)
		log := setupLogger()
		handler := handler.NewUserHandler(mockService, log)

		input := model.RegisterRequest{
			Email:    "existing@example.com",
			Password: "password123",
		}

		mockService.On("RegisterUser", mock.Anything, input).Return(nil, errors.New(ErrEmailAlreadyExists))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body, _ := json.Marshal(input)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.RegisterUser(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should login user successfully", func(t *testing.T) {
		mockService := new(MockService)
		log := setupLogger()
		handler := handler.NewUserHandler(mockService, log)

		input := model.LoginRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		authResp := &model.AuthResponse{
			User: &model.User{
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

		handler.LoginUser(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Login successful", response["message"])
		assert.NotNil(t, response["data"])

		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid credentials", func(t *testing.T) {
		mockService := new(MockService)
		log := setupLogger()
		handler := handler.NewUserHandler(mockService, log)

		input := model.LoginRequest{
			Email:    "test@example.com",
			Password: "wrong-password",
		}

		mockService.On("LoginUser", mock.Anything, input).Return(nil, errors.New(ErrInvalidCredentials))

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body, _ := json.Marshal(input)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.LoginUser(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("should logout user successfully", func(t *testing.T) {
		mockService := new(MockService)
		log := setupLogger()
		handler := handler.NewUserHandler(mockService, log)

		logoutReq := map[string]interface{}{
			"user_id":    1,
			"session_id": "session-123",
		}

		mockService.On("LogoutUser", mock.Anything, uint(1), "session-123").Return(nil)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		body, _ := json.Marshal(logoutReq)
		c.Request = httptest.NewRequest(http.MethodPost, "/auth/logout", bytes.NewBuffer(body))
		c.Request.Header.Set("Content-Type", "application/json")

		handler.LogoutUser(c)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "Logged out successfully", response["message"])

		mockService.AssertExpectations(t)
	})

	t.Run("should return error when session_id cookie is missing", func(t *testing.T) {
		mockService := new(MockService)
		log := setupLogger()
		handler := handler.NewUserHandler(mockService, log)

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		c.Request = httptest.NewRequest(http.MethodPost, "/auth/logout", nil)

		handler.LogoutUser(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}