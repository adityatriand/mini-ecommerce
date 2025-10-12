package middleware

import (
	"mini-e-commerce/internal/logger"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Info(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Debug(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...zap.Field) {
	m.Called(msg, fields)
}

func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockLogger) WithContext(c any) logger.ContextLogger {
	args := m.Called(c)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(logger.ContextLogger)
}

func (m *MockLogger) GetZapLogger() *zap.Logger {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*zap.Logger)
}

func setupLoggingTest() (*gin.Engine, *MockLogger) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockLogger := new(MockLogger)
	return router, mockLogger
}

func TestRequestLogger(t *testing.T) {
	t.Run("should set request_id and log HTTP request", func(t *testing.T) {
		router, mockLogger := setupLoggingTest()

		mockLogger.On("Info", "HTTP Request", mock.Anything).Return()

		router.Use(RequestLogger(mockLogger))
		router.GET("/test", func(c *gin.Context) {
			requestID, exists := c.Get("request_id")
			assert.True(t, exists)
			assert.NotEmpty(t, requestID)
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockLogger.AssertCalled(t, "Info", "HTTP Request", mock.Anything)
	})

	t.Run("should log request with query parameters", func(t *testing.T) {
		router, mockLogger := setupLoggingTest()

		mockLogger.On("Info", "HTTP Request", mock.Anything).Return()

		router.Use(RequestLogger(mockLogger))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req := httptest.NewRequest(http.MethodGet, "/test?foo=bar&baz=qux", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockLogger.AssertCalled(t, "Info", "HTTP Request", mock.Anything)
	})

	t.Run("should log POST request", func(t *testing.T) {
		router, mockLogger := setupLoggingTest()

		mockLogger.On("Info", "HTTP Request", mock.Anything).Return()

		router.Use(RequestLogger(mockLogger))
		router.POST("/test", func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"created": true})
		})

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockLogger.AssertCalled(t, "Info", "HTTP Request", mock.Anything)
	})

	t.Run("should log error status codes", func(t *testing.T) {
		router, mockLogger := setupLoggingTest()

		mockLogger.On("Info", "HTTP Request", mock.Anything).Return()

		router.Use(RequestLogger(mockLogger))
		router.GET("/error", func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		})

		req := httptest.NewRequest(http.MethodGet, "/error", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockLogger.AssertCalled(t, "Info", "HTTP Request", mock.Anything)
	})

	t.Run("should log with user agent", func(t *testing.T) {
		router, mockLogger := setupLoggingTest()

		mockLogger.On("Info", "HTTP Request", mock.Anything).Return()

		router.Use(RequestLogger(mockLogger))
		router.GET("/test", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("User-Agent", "TestAgent/1.0")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockLogger.AssertCalled(t, "Info", "HTTP Request", mock.Anything)
	})
}

func TestErrorLogger(t *testing.T) {
	t.Run("should recover from panic and log error", func(t *testing.T) {
		router, mockLogger := setupLoggingTest()

		mockLogger.On("Error", "Panic recovered", mock.Anything).Return()

		router.Use(ErrorLogger(mockLogger))
		router.GET("/panic", func(c *gin.Context) {
			panic("something went wrong")
		})

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.Contains(t, w.Body.String(), "Internal server error")
		assert.Contains(t, w.Body.String(), "INTERNAL_SERVER_ERROR")
		mockLogger.AssertCalled(t, "Error", "Panic recovered", mock.Anything)
	})

	t.Run("should recover from panic with nil value", func(t *testing.T) {
		router, mockLogger := setupLoggingTest()

		mockLogger.On("Error", "Panic recovered", mock.Anything).Return()

		router.Use(ErrorLogger(mockLogger))
		router.GET("/panic-nil", func(c *gin.Context) {
			panic(nil)
		})

		req := httptest.NewRequest(http.MethodGet, "/panic-nil", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockLogger.AssertCalled(t, "Error", "Panic recovered", mock.Anything)
	})

	t.Run("should recover from panic and include request_id if present", func(t *testing.T) {
		router, mockLogger := setupLoggingTest()

		mockLogger.On("Error", "Panic recovered", mock.Anything).Return()

		router.Use(func(c *gin.Context) {
			c.Set("request_id", "test-request-id-123")
			c.Next()
		})
		router.Use(ErrorLogger(mockLogger))
		router.GET("/panic", func(c *gin.Context) {
			panic("test panic with request_id")
		})

		req := httptest.NewRequest(http.MethodGet, "/panic", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockLogger.AssertCalled(t, "Error", "Panic recovered", mock.Anything)
	})

	t.Run("should not panic when no panic occurs", func(t *testing.T) {
		router, mockLogger := setupLoggingTest()

		router.Use(ErrorLogger(mockLogger))
		router.GET("/ok", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"success": true})
		})

		req := httptest.NewRequest(http.MethodGet, "/ok", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockLogger.AssertNotCalled(t, "Error", "Panic recovered", mock.Anything)
	})
}

func TestExtractRequestIDSafely(t *testing.T) {
	t.Run("should return request_id when present", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("request_id", "test-id-123")

		result := extractRequestIDSafely(c)

		assert.Equal(t, "test-id-123", result)
	})

	t.Run("should return missing_context when context is nil", func(t *testing.T) {
		result := extractRequestIDSafely(nil)

		assert.Equal(t, "missing_context", result)
	})

	t.Run("should return missing_request_id when request_id not set", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())

		result := extractRequestIDSafely(c)

		assert.Equal(t, "missing_request_id", result)
	})

	t.Run("should return empty_request_id when request_id is empty string", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("request_id", "")

		result := extractRequestIDSafely(c)

		assert.Equal(t, "empty_request_id", result)
	})

	t.Run("should return invalid_request_id_type when request_id is not a string", func(t *testing.T) {
		gin.SetMode(gin.TestMode)
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("request_id", 12345)

		result := extractRequestIDSafely(c)

		assert.Equal(t, "invalid_request_id_type", result)
	})
}
