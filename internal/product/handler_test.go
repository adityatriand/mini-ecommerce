package product

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
	"gorm.io/gorm"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) CreateProduct(ctx context.Context, input CreateProductRequest) (*Product, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Product), args.Error(1)
}

func (m *MockService) GetAllProducts(ctx context.Context) ([]Product, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Product), args.Error(1)
}

func (m *MockService) GetAllProductsWithQuery(ctx context.Context, query ProductQuery) (*ProductListResponse, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ProductListResponse), args.Error(1)
}

func (m *MockService) GetProductByID(ctx context.Context, id uint) (*Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Product), args.Error(1)
}

func (m *MockService) UpdateProduct(ctx context.Context, id uint, input UpdateProductRequest) (*Product, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Product), args.Error(1)
}

func (m *MockService) DeleteProduct(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockService) UpdateStock(ctx context.Context, id uint, stockDelta int) error {
	args := m.Called(ctx, id, stockDelta)
	return args.Error(0)
}

func (m *MockService) UpdateStockWithTx(tx *gorm.DB, id uint, stockDelta int) error {
	args := m.Called(tx, id, stockDelta)
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

func TestHandler_CreateProduct(t *testing.T) {
	t.Run("should create product successfully", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.POST("/products", handler.CreateProduct)

		input := CreateProductRequest{
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		expectedProduct := &Product{
			ID:    1,
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mockService.On("CreateProduct", mock.Anything, input).Return(expectedProduct, nil)

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid input", func(t *testing.T) {
		router, _, handler := setupHandlerTest()
		router.POST("/products", handler.CreateProduct)

		invalidInput := map[string]interface{}{
			"name": 12345,
		}

		body, _ := json.Marshal(invalidInput)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error when service fails", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.POST("/products", handler.CreateProduct)

		input := CreateProductRequest{
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mockService.On("CreateProduct", mock.Anything, input).Return(nil, errors.New("database error"))

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_GetAllProducts(t *testing.T) {
	t.Run("should get all products successfully", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.GET("/products", handler.GetAllProducts)

		expectedResponse := &ProductListResponse{
			Data: []Product{
				{ID: 1, Name: "Product 1", Price: 10000, Stock: 50},
				{ID: 2, Name: "Product 2", Price: 20000, Stock: 30},
			},
			Pagination: dto.PaginationMetadata{
				Page:       1,
				PageSize:   10,
				Total:      2,
				TotalPages: 1,
			},
		}

		mockService.On("GetAllProductsWithQuery", mock.Anything, mock.AnythingOfType("product.ProductQuery")).
			Return(expectedResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/products", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should get products with query parameters", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.GET("/products", handler.GetAllProducts)

		expectedResponse := &ProductListResponse{
			Data: []Product{
				{ID: 1, Name: "Product 1", Price: 10000, Stock: 50},
			},
			Pagination: dto.PaginationMetadata{
				Page:       2,
				PageSize:   5,
				Total:      10,
				TotalPages: 2,
			},
		}

		mockService.On("GetAllProductsWithQuery", mock.Anything, mock.AnythingOfType("product.ProductQuery")).
			Return(expectedResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/products?page=2&page_size=5&order=asc&sort_by=name", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when service fails", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.GET("/products", handler.GetAllProducts)

		mockService.On("GetAllProductsWithQuery", mock.Anything, mock.AnythingOfType("product.ProductQuery")).
			Return(nil, errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/products", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_GetProductByID(t *testing.T) {
	t.Run("should get product by ID successfully", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.GET("/products/:id", handler.GetProductByID)

		expectedProduct := &Product{
			ID:    1,
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mockService.On("GetProductByID", mock.Anything, uint(1)).Return(expectedProduct, nil)

		req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid ID", func(t *testing.T) {
		router, _, handler := setupHandlerTest()
		router.GET("/products/:id", handler.GetProductByID)

		req := httptest.NewRequest(http.MethodGet, "/products/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return not found when product does not exist", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.GET("/products/:id", handler.GetProductByID)

		mockService.On("GetProductByID", mock.Anything, uint(999)).
			Return(nil, errors.New("product not found"))

		req := httptest.NewRequest(http.MethodGet, "/products/999", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_UpdateProduct(t *testing.T) {
	t.Run("should update product successfully", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.PATCH("/products/:id", handler.UpdateProduct)

		newName := "Updated Product"
		newPrice := 15000
		input := UpdateProductRequest{
			Name:  &newName,
			Price: &newPrice,
		}

		expectedProduct := &Product{
			ID:    1,
			Name:  "Updated Product",
			Price: 15000,
			Stock: 50,
		}

		mockService.On("UpdateProduct", mock.Anything, uint(1), input).
			Return(expectedProduct, nil)

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPatch, "/products/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid ID", func(t *testing.T) {
		router, _, handler := setupHandlerTest()
		router.PATCH("/products/:id", handler.UpdateProduct)

		newName := "Updated Product"
		input := UpdateProductRequest{
			Name: &newName,
		}

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPatch, "/products/invalid", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error for invalid input", func(t *testing.T) {
		router, _, handler := setupHandlerTest()
		router.PATCH("/products/:id", handler.UpdateProduct)

		invalidInput := map[string]interface{}{
			"price": "invalid",
		}

		body, _ := json.Marshal(invalidInput)
		req := httptest.NewRequest(http.MethodPatch, "/products/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error when service fails", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.PATCH("/products/:id", handler.UpdateProduct)

		newName := "Updated Product"
		input := UpdateProductRequest{
			Name: &newName,
		}

		mockService.On("UpdateProduct", mock.Anything, uint(1), input).
			Return(nil, errors.New("database error"))

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPatch, "/products/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_DeleteProduct(t *testing.T) {
	t.Run("should delete product successfully", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.DELETE("/products/:id", handler.DeleteProduct)

		mockService.On("DeleteProduct", mock.Anything, uint(1)).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/products/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid ID", func(t *testing.T) {
		router, _, handler := setupHandlerTest()
		router.DELETE("/products/:id", handler.DeleteProduct)

		req := httptest.NewRequest(http.MethodDelete, "/products/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error when service fails", func(t *testing.T) {
		router, mockService, handler := setupHandlerTest()
		router.DELETE("/products/:id", handler.DeleteProduct)

		mockService.On("DeleteProduct", mock.Anything, uint(1)).
			Return(errors.New("database error"))

		req := httptest.NewRequest(http.MethodDelete, "/products/1", nil)
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
