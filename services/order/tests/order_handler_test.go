package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"mini-e-commerce/services/order/internal/handler"
	"mini-e-commerce/services/order/internal/model"
	"mini-e-commerce/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) CreateOrder(ctx context.Context, input model.CreateOrderRequest, userID uint) (*model.Order, error) {
	args := m.Called(ctx, input, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockService) GetAllOrders(ctx context.Context, query model.OrderQuery, userID uint) (*model.OrderListResponse, error) {
	args := m.Called(ctx, query, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OrderListResponse), args.Error(1)
}

func (m *MockService) GetAllOrdersWithQuery(ctx context.Context, query model.OrderQuery) (*model.OrderListResponse, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OrderListResponse), args.Error(1)
}

func (m *MockService) GetOrderByID(ctx context.Context, id uint, userID uint) (*model.Order, error) {
	args := m.Called(ctx, id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockService) UpdateOrder(ctx context.Context, id uint, input model.UpdateOrderRequest, userID uint) (*model.Order, error) {
	args := m.Called(ctx, id, input, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func (m *MockService) DeleteOrder(ctx context.Context, id uint, userID uint) error {
	args := m.Called(ctx, id, userID)
	return args.Error(0)
}

func (m *MockService) GetOrdersByUserID(ctx context.Context, userID uint, query model.OrderQuery) (*model.OrderListResponse, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.OrderListResponse), args.Error(1)
}

func (m *MockService) CreateOrderOptimized(ctx context.Context, input model.CreateOrderRequest, userID uint) (*model.Order, error) {
	args := m.Called(ctx, input, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Order), args.Error(1)
}

func setupOrderHandlerTest() (*gin.Engine, *MockService, *handler.OrderHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockService := new(MockService)
	logConfig := logger.Config{
		Level:      "info",
		Format:     "console",
		OutputPath: "",
	}
	log, _ := logger.NewLogger(logConfig, "test")
	handler := handler.NewOrderHandler(mockService, log.GetZapLogger())
	return router, mockService, handler
}

func TestOrderHandler_CreateOrder(t *testing.T) {
	t.Run("should create order successfully", func(t *testing.T) {
		router, mockService, handler := setupOrderHandlerTest()
		router.POST("/orders", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.CreateOrder(c)
		})

		input := model.CreateOrderRequest{
			Items: []model.OrderItemRequest{
				{ProductID: 1, Quantity: 2},
			},
		}

		expectedOrder := &model.Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 20000,
			Status:     model.StatusPending,
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
		router, _, handler := setupOrderHandlerTest()
		router.POST("/orders", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.CreateOrder(c)
		})

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
		router, _, handler := setupOrderHandlerTest()
		router.POST("/orders", handler.CreateOrder)

		input := model.CreateOrderRequest{
			Items: []model.OrderItemRequest{
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
		router, mockService, handler := setupOrderHandlerTest()
		router.POST("/orders", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.CreateOrder(c)
		})

		input := model.CreateOrderRequest{
			Items: []model.OrderItemRequest{
				{ProductID: 999, Quantity: 1},
			},
		}

		mockService.On("CreateOrder", mock.Anything, input, uint(1)).
			Return(nil, model.ErrProductNotFound)

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when insufficient stock", func(t *testing.T) {
		router, mockService, handler := setupOrderHandlerTest()
		router.POST("/orders", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.CreateOrder(c)
		})

		input := model.CreateOrderRequest{
			Items: []model.OrderItemRequest{
				{ProductID: 1, Quantity: 100},
			},
		}

		mockService.On("CreateOrder", mock.Anything, input, uint(1)).
			Return(nil, model.ErrInsufficientStock)

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when service fails", func(t *testing.T) {
		router, mockService, handler := setupOrderHandlerTest()
		router.POST("/orders", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.CreateOrder(c)
		})

		input := model.CreateOrderRequest{
			Items: []model.OrderItemRequest{
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

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestOrderHandler_GetAllOrders(t *testing.T) {
	t.Run("should get all orders successfully", func(t *testing.T) {
		router, mockService, handler := setupOrderHandlerTest()
		router.GET("/orders", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.GetOrders(c)
		})

		expectedResponse := &model.OrderListResponse{
			Data: []model.Order{
				{ID: 1, UserID: 1, TotalPrice: 50000, Status: model.StatusPending},
				{ID: 2, UserID: 2, TotalPrice: 75000, Status: model.StatusPaid},
			},
			Pagination: model.PaginationMetadata{
				Page:       1,
				PageSize:   10,
				Total:      2,
				TotalPages: 1,
			},
		}

		mockService.On("GetOrdersByUserID", mock.Anything, uint(1), mock.AnythingOfType("model.OrderQuery")).
			Return(expectedResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should get orders with query parameters", func(t *testing.T) {
		router, mockService, handler := setupOrderHandlerTest()
		router.GET("/orders", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.GetOrders(c)
		})

		expectedResponse := &model.OrderListResponse{
			Data: []model.Order{
				{ID: 1, UserID: 1, TotalPrice: 50000, Status: model.StatusPending},
			},
			Pagination: model.PaginationMetadata{
				Page:       2,
				PageSize:   5,
				Total:      10,
				TotalPages: 2,
			},
		}

		mockService.On("GetOrdersByUserID", mock.Anything, uint(1), mock.AnythingOfType("model.OrderQuery")).
			Return(expectedResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders?page=2&page_size=5&order=asc&sort_by=user_id", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when service fails", func(t *testing.T) {
		router, mockService, handler := setupOrderHandlerTest()
		router.GET("/orders", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.GetOrders(c)
		})

		mockService.On("GetOrdersByUserID", mock.Anything, uint(1), mock.AnythingOfType("model.OrderQuery")).
			Return(nil, errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/orders", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestOrderHandler_GetOrderByID(t *testing.T) {
	t.Run("should get order by ID successfully", func(t *testing.T) {
		router, mockService, handler := setupOrderHandlerTest()
		router.GET("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.GetOrderByID(c)
		})

		expectedOrder := &model.Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 50000,
			Status:     model.StatusPending,
		}

		mockService.On("GetOrderByID", mock.Anything, uint(1), uint(1)).Return(expectedOrder, nil)

		req := httptest.NewRequest(http.MethodGet, "/orders/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid ID", func(t *testing.T) {
		router, _, handler := setupOrderHandlerTest()
		router.GET("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.GetOrderByID(c)
		})

		req := httptest.NewRequest(http.MethodGet, "/orders/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return not found when order does not exist", func(t *testing.T) {
		router, mockService, handler := setupOrderHandlerTest()
		router.GET("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.GetOrderByID(c)
		})

		mockService.On("GetOrderByID", mock.Anything, uint(999), uint(1)).
			Return(nil, model.ErrOrderNotFound)

		req := httptest.NewRequest(http.MethodGet, "/orders/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestOrderHandler_UpdateOrder(t *testing.T) {
	t.Run("should update order successfully", func(t *testing.T) {
		router, mockService, handler := setupOrderHandlerTest()
		router.PATCH("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.UpdateOrder(c)
		})

		newStatus := model.StatusPaid
		input := model.UpdateOrderRequest{
			Status: &newStatus,
		}

		expectedOrder := &model.Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 50000,
			Status:     model.StatusPaid,
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
		router, _, handler := setupOrderHandlerTest()
		router.PATCH("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.UpdateOrder(c)
		})

		newStatus := model.StatusPaid
		input := model.UpdateOrderRequest{
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
		router, _, handler := setupOrderHandlerTest()
		router.PATCH("/orders/:id", handler.UpdateOrder)

		newStatus := model.StatusPaid
		input := model.UpdateOrderRequest{
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
		router, mockService, handler := setupOrderHandlerTest()
		router.PATCH("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.UpdateOrder(c)
		})

		newStatus := model.StatusPaid
		input := model.UpdateOrderRequest{
			Status: &newStatus,
		}

		mockService.On("UpdateOrder", mock.Anything, uint(999), input, uint(1)).
			Return(nil, model.ErrOrderNotFound)

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPatch, "/orders/999", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when user not authorized", func(t *testing.T) {
		router, mockService, handler := setupOrderHandlerTest()
		router.PATCH("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(999))
			handler.UpdateOrder(c)
		})

		newStatus := model.StatusPaid
		input := model.UpdateOrderRequest{
			Status: &newStatus,
		}

		mockService.On("UpdateOrder", mock.Anything, uint(1), input, uint(999)).
			Return(nil, model.ErrNotAuthorizedToUpdate)

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPatch, "/orders/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid status", func(t *testing.T) {
		router, mockService, handler := setupOrderHandlerTest()
		router.PATCH("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.UpdateOrder(c)
		})

		newStatus := model.StatusPaid
		input := model.UpdateOrderRequest{
			Status: &newStatus,
		}

		mockService.On("UpdateOrder", mock.Anything, uint(1), input, uint(1)).
			Return(nil, model.ErrInvalidStatusValue)

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPatch, "/orders/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestOrderHandler_DeleteOrder(t *testing.T) {
	t.Run("should delete order successfully", func(t *testing.T) {
		router, mockService, handler := setupOrderHandlerTest()
		router.DELETE("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.DeleteOrder(c)
		})

		mockService.On("DeleteOrder", mock.Anything, uint(1), uint(1)).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/orders/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid ID", func(t *testing.T) {
		router, _, handler := setupOrderHandlerTest()
		router.DELETE("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.DeleteOrder(c)
		})

		req := httptest.NewRequest(http.MethodDelete, "/orders/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error when order not found", func(t *testing.T) {
		router, mockService, handler := setupOrderHandlerTest()
		router.DELETE("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.DeleteOrder(c)
		})

		mockService.On("DeleteOrder", mock.Anything, uint(999), uint(1)).
			Return(model.ErrOrderNotFound)

		req := httptest.NewRequest(http.MethodDelete, "/orders/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when service fails", func(t *testing.T) {
		router, mockService, handler := setupOrderHandlerTest()
		router.DELETE("/orders/:id", func(c *gin.Context) {
			c.Set("user_id", uint(1))
			handler.DeleteOrder(c)
		})

		mockService.On("DeleteOrder", mock.Anything, uint(1), uint(1)).
			Return(errors.New("database error"))

		req := httptest.NewRequest(http.MethodDelete, "/orders/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestNewOrderHandler(t *testing.T) {
	t.Run("should create handler successfully", func(t *testing.T) {
		mockService := new(MockService)
		logConfig := logger.Config{
			Level:      "info",
			Format:     "console",
			OutputPath: "",
		}
		log, _ := logger.NewLogger(logConfig, "test")
		handler := handler.NewOrderHandler(mockService, log.GetZapLogger())

		require.NotNil(t, handler)
	})
}