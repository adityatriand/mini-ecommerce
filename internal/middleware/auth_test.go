package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"mini-e-commerce/internal/auth"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) Generate(userID uint) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockJWTManager) Verify(token string) (*auth.UserClaims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.UserClaims), args.Error(1)
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

func setupAuthTest() (*gin.Engine, *MockJWTManager, *MockSessionManager) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockJWT := new(MockJWTManager)
	mockSession := new(MockSessionManager)
	return router, mockJWT, mockSession
}

func TestAuthMiddleware_JWTAuth(t *testing.T) {
	t.Run("should authenticate with valid JWT token", func(t *testing.T) {
		router, mockJWT, mockSession := setupAuthTest()
		logger := zap.NewNop()

		mockJWT.On("Verify", "valid-token").Return(&auth.UserClaims{UserID: 123}, nil)

		router.GET("/protected", AuthMiddleware(mockJWT, mockSession, logger), func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockJWT.AssertExpectations(t)
	})

	t.Run("should reject expired JWT token and fallback to session", func(t *testing.T) {
		router, mockJWT, mockSession := setupAuthTest()
		logger := zap.NewNop()

		mockJWT.On("Verify", "expired-token").Return(nil, auth.ErrExpiredToken)

		router.GET("/protected", AuthMiddleware(mockJWT, mockSession, logger), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer expired-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockJWT.AssertExpectations(t)
	})

	t.Run("should reject invalid JWT token and fallback to session", func(t *testing.T) {
		router, mockJWT, mockSession := setupAuthTest()
		logger := zap.NewNop()

		mockJWT.On("Verify", "invalid-token").Return(nil, errors.New("invalid token"))

		router.GET("/protected", AuthMiddleware(mockJWT, mockSession, logger), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockJWT.AssertExpectations(t)
	})
}

func TestAuthMiddleware_SessionAuth(t *testing.T) {
	t.Run("should authenticate with valid session", func(t *testing.T) {
		router, mockJWT, mockSession := setupAuthTest()
		logger := zap.NewNop()

		mockSession.On("ValidateRefreshToken", mock.Anything, uint(456), "session-123", "refresh-token-456").Return(nil)

		router.GET("/protected", AuthMiddleware(mockJWT, mockSession, logger), func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "session-123"})
		req.AddCookie(&http.Cookie{Name: "user_id", Value: "456"})
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh-token-456"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockSession.AssertExpectations(t)
	})

	t.Run("should reject when session_id cookie is missing", func(t *testing.T) {
		router, mockJWT, mockSession := setupAuthTest()
		logger := zap.NewNop()

		router.GET("/protected", AuthMiddleware(mockJWT, mockSession, logger), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "authentication required")
	})

	t.Run("should reject when user_id cookie is missing", func(t *testing.T) {
		router, mockJWT, mockSession := setupAuthTest()
		logger := zap.NewNop()

		router.GET("/protected", AuthMiddleware(mockJWT, mockSession, logger), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "session-123"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid session")
	})

	t.Run("should reject when user_id cookie has invalid format", func(t *testing.T) {
		router, mockJWT, mockSession := setupAuthTest()
		logger := zap.NewNop()

		router.GET("/protected", AuthMiddleware(mockJWT, mockSession, logger), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "session-123"})
		req.AddCookie(&http.Cookie{Name: "user_id", Value: "invalid"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid session")
	})

	t.Run("should reject when refresh_token cookie is missing", func(t *testing.T) {
		router, mockJWT, mockSession := setupAuthTest()
		logger := zap.NewNop()

		router.GET("/protected", AuthMiddleware(mockJWT, mockSession, logger), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "session-123"})
		req.AddCookie(&http.Cookie{Name: "user_id", Value: "456"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid session")
	})

	t.Run("should reject when session is not found", func(t *testing.T) {
		router, mockJWT, mockSession := setupAuthTest()
		logger := zap.NewNop()

		mockSession.On("ValidateRefreshToken", mock.Anything, uint(456), "session-999", "refresh-token-456").
			Return(auth.ErrSessionNotFound)

		router.GET("/protected", AuthMiddleware(mockJWT, mockSession, logger), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "session-999"})
		req.AddCookie(&http.Cookie{Name: "user_id", Value: "456"})
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh-token-456"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid session")
		mockSession.AssertExpectations(t)
	})

	t.Run("should reject when refresh token is invalid", func(t *testing.T) {
		router, mockJWT, mockSession := setupAuthTest()
		logger := zap.NewNop()

		mockSession.On("ValidateRefreshToken", mock.Anything, uint(456), "session-123", "wrong-token").
			Return(auth.ErrInvalidRefreshToken)

		router.GET("/protected", AuthMiddleware(mockJWT, mockSession, logger), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "session-123"})
		req.AddCookie(&http.Cookie{Name: "user_id", Value: "456"})
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "wrong-token"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid session")
		mockSession.AssertExpectations(t)
	})

	t.Run("should reject when session validation fails with unknown error", func(t *testing.T) {
		router, mockJWT, mockSession := setupAuthTest()
		logger := zap.NewNop()

		mockSession.On("ValidateRefreshToken", mock.Anything, uint(456), "session-123", "refresh-token-456").
			Return(errors.New("database error"))

		router.GET("/protected", AuthMiddleware(mockJWT, mockSession, logger), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "session-123"})
		req.AddCookie(&http.Cookie{Name: "user_id", Value: "456"})
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh-token-456"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		assert.Contains(t, w.Body.String(), "invalid session")
		mockSession.AssertExpectations(t)
	})
}

func TestAuthMiddleware_Priority(t *testing.T) {
	t.Run("should prefer JWT over session when both are present", func(t *testing.T) {
		router, mockJWT, mockSession := setupAuthTest()
		logger := zap.NewNop()

		mockJWT.On("Verify", "valid-jwt").Return(&auth.UserClaims{UserID: 100}, nil)

		router.GET("/protected", AuthMiddleware(mockJWT, mockSession, logger), func(c *gin.Context) {
			userID, _ := c.Get("user_id")
			c.JSON(http.StatusOK, gin.H{"user_id": userID})
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", "Bearer valid-jwt")
		req.AddCookie(&http.Cookie{Name: "session_id", Value: "session-123"})
		req.AddCookie(&http.Cookie{Name: "user_id", Value: "456"})
		req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "refresh-token-456"})
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "100")
		mockJWT.AssertExpectations(t)
		mockSession.AssertNotCalled(t, "ValidateRefreshToken")
	})
}
