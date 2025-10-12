package order

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"mini-e-commerce/internal/dto"
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

func (m *MockService) CreateOrder(ctx context.Context, input CreateOrderRequest, userID uint) (*Order, error) {
	args := m.Called(ctx, input, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Order), args.Error(1)
}

func (m *MockService) GetAllOrders(ctx context.Context) ([]Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Order), args.Error(1)
}

func (m *MockService) GetAllOrdersWithQuery(ctx context.Context, query OrderQuery) (*OrderListResponse, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*OrderListResponse), args.Error(1)
}

func (m *MockService) GetOrderByID(ctx context.Context, id uint) (*Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Order), args.Error(1)
}

func (m *MockService) UpdateOrder(ctx context.Context, id uint, input UpdateOrderRequest, userID uint) (*Order, error) {
	args := m.Called(ctx, id, input, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Order), args.Error(1)
}

func (m *MockService) DeleteOrder(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupHandlerTest() (*gin.Engine, *MockService, *Handler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockService := new(MockService)
	log, _ := logger.NewLogger(&logger.Config{
		ServiceName: "test",
		AppVersion:  "1.0",
		LogLevel:    zapcore.InfoLevel,
		Mode:        "development",
	})
	handler := NewHandler(mockService, log)
	return router, mockService, handler
}

func TestHandler_CreateOrder(t *testing.T) {
	t.Run("should create order successfully", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.POST("/orders", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.CreateOrder(c)
		})

		input := CreateOrderRequest{
			Items: []OrderItemInput{
				{ProductID: 1, Quantity: 2},
			},
		}

		expectedOrder := &Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 20000,
			Status:     StatusPending,
		}

		mockService.On("CreateOrder", mock.Anything, input, uint(1)).Return(expectedOrder, nil)

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid input", func(t *testing.T) {
		router, _, handler := setupHandlerTest()
		router.POST("/orders", handler.CreateOrder)

		invalidInput := map[string]any{
			"items": "invalid",
		}

		body, _ := json.Marshal(invalidInput)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error when user ID missing from context", func(t *testing.T) {
		router, _, handler := setupHandlerTest()
		router.POST("/orders", handler.CreateOrder)

		input := CreateOrderRequest{
			Items: []OrderItemInput{
				{ProductID: 1, Quantity: 2},
			},
		}

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return error when product not found", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.POST("/orders", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.CreateOrder(c)
		})

		input := CreateOrderRequest{
			Items: []OrderItemInput{
				{ProductID: 999, Quantity: 1},
			},
		}

		mockService.On("CreateOrder", mock.Anything, input, uint(1)).
			Return(nil, errors.New(ErrProductNotFound))

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when insufficient stock", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.POST("/orders", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.CreateOrder(c)
		})

		input := CreateOrderRequest{
			Items: []OrderItemInput{
				{ProductID: 1, Quantity: 100},
			},
		}

		mockService.On("CreateOrder", mock.Anything, input, uint(1)).
			Return(nil, errors.New(ErrInsufficientStock))

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when service fails", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.POST("/orders", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.CreateOrder(c)
		})

		input := CreateOrderRequest{
			Items: []OrderItemInput{
				{ProductID: 1, Quantity: 2},
			},
		}

		mockService.On("CreateOrder", mock.Anything, input, uint(1)).
			Return(nil, errors.New("database error"))

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_GetOrders(t *testing.T) {
	t.Run("should get all orders successfully", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.GET("/orders", handler.GetOrders)

		expectedResponse := &OrderListResponse{
			Data: []Order{
				{ID: 1, UserID: 1, TotalPrice: 50000, Status: StatusPending},
				{ID: 2, UserID: 2, TotalPrice: 75000, Status: StatusPaid},
			},
			Pagination: dto.PaginationMetadata{
				Page:       1,
				PageSize:   10,
				Total:      2,
				TotalPages: 1,
			},
		}

		mockService.On("GetAllOrdersWithQuery", mock.Anything, mock.AnythingOfType("order.OrderQuery")).
			Return(expectedResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should get orders with query parameters", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.GET("/orders", handler.GetOrders)

		expectedResponse := &OrderListResponse{
			Data: []Order{
				{ID: 1, UserID: 1, TotalPrice: 50000, Status: StatusPending},
			},
			Pagination: dto.PaginationMetadata{
				Page:       2,
				PageSize:   5,
				Total:      10,
				TotalPages: 2,
			},
		}

		mockService.On("GetAllOrdersWithQuery", mock.Anything, mock.AnythingOfType("order.OrderQuery")).
			Return(expectedResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders?page=2&page_size=5&order=asc&sort_by=user_id", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when service fails", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.GET("/orders", handler.GetOrders)

		mockService.On("GetAllOrdersWithQuery", mock.Anything, mock.AnythingOfType("order.OrderQuery")).
			Return(nil, errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_GetOrderByID(t *testing.T) {
	t.Run("should get order by ID successfully", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.GET("/orders/:id", handler.GetOrderByID)

		expectedOrder := &Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 50000,
			Status:     StatusPending,
		}

		mockService.On("GetOrderByID", mock.Anything, uint(1)).Return(expectedOrder, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid ID", func(t *testing.T) {
		router, _, handler := setupHandlerTest()
		router.GET("/orders/:id", handler.GetOrderByID)

		req := httptest.NewRequest(http.MethodGet, "/orders/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return not found when order does not exist", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.GET("/orders/:id", handler.GetOrderByID)

		mockService.On("GetOrderByID", mock.Anything, uint(999)).
			Return(nil, errors.New(ErrOrderNotFound))

		req := httptest.NewRequest(http.MethodGet, "/orders/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_UpdateOrder(t *testing.T) {
	t.Run("should update order successfully", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.PATCH("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.UpdateOrder(c)
		})

		newStatus := StatusPaid
		input := UpdateOrderRequest{
			Status: &newStatus,
		}

		expectedOrder := &Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 50000,
			Status:     StatusPaid,
		}

		mockService.On("UpdateOrder", mock.Anything, uint(1), input, uint(1)).
			Return(expectedOrder, nil)

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPatch, "/orders/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid ID", func(t *testing.T) {
		router, _, handler := setupHandlerTest()
		router.PATCH("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.UpdateOrder(c)
		})

		newStatus := StatusPaid
		input := UpdateOrderRequest{
			Status: &newStatus,
		}

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPatch, "/orders/invalid", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error when user ID missing from context", func(t *testing.T) {
		router, _, handler := setupHandlerTest()
		router.PATCH("/orders/:id", handler.UpdateOrder)

		newStatus := StatusPaid
		input := UpdateOrderRequest{
			Status: &newStatus,
		}

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPatch, "/orders/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("should return error when order not found", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.PATCH("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.UpdateOrder(c)
		})

		newStatus := StatusPaid
		input := UpdateOrderRequest{
			Status: &newStatus,
		}

		mockService.On("UpdateOrder", mock.Anything, uint(999), input, uint(1)).
			Return(nil, errors.New(ErrOrderNotFound))

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPatch, "/orders/999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when user not authorized", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.PATCH("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(999))
			handler.UpdateOrder(c)
		})

		newStatus := StatusPaid
		input := UpdateOrderRequest{
			Status: &newStatus,
		}

		mockService.On("UpdateOrder", mock.Anything, uint(1), input, uint(999)).
			Return(nil, errors.New(ErrNotAuthorizedToUpdate))

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPatch, "/orders/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid status", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.PATCH("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.UpdateOrder(c)
		})

		newStatus := StatusPaid
		input := UpdateOrderRequest{
			Status: &newStatus,
		}

		mockService.On("UpdateOrder", mock.Anything, uint(1), input, uint(1)).
			Return(nil, errors.New(ErrInvalidStatusValue))

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPatch, "/orders/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_DeleteOrder(t *testing.T) {
	t.Run("should delete order successfully", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.DELETE("/orders/:id", handler.DeleteOrder)

		mockService.On("DeleteOrder", mock.Anything, uint(1)).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/orders/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid ID", func(t *testing.T) {
		router, _, handler := setupHandlerTest()
		router.DELETE("/orders/:id", handler.DeleteOrder)

		req := httptest.NewRequest(http.MethodDelete, "/orders/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error when order not found", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.DELETE("/orders/:id", handler.DeleteOrder)

		mockService.On("DeleteOrder", mock.Anything, uint(999)).
			Return(errors.New(ErrOrderNotFound))

		req := httptest.NewRequest(http.MethodDelete, "/orders/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when service fails", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.DELETE("/orders/:id", handler.DeleteOrder)

		mockService.On("DeleteOrder", mock.Anything, uint(1)).
			Return(errors.New("database error"))

		req := httptest.NewRequest(http.MethodDelete, "/orders/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestNewHandler(t *testing.T) {
	t.Run("should create handler successfully", func(t *testing.T) {
		mockService := new(MockService)
		log, _ := logger.NewLogger(&logger.Config{
			ServiceName: "test",
			AppVersion:  "1.0",
			LogLevel:    zapcore.InfoLevel,
			Mode:        "development",
		})
		handler := NewHandler(mockService, log)

		require.NotNil(t, handler)
		assert.NotNil(t, handler.service)
		assert.NotNil(t, handler.logger)
		assert.NotNil(t, handler.responseHelper)
	})
}
