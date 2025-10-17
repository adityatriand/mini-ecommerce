package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"mini-e-commerce/services/product/internal/handler"
	"mini-e-commerce/services/product/internal/model"
	"mini-e-commerce/shared/logger"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) CreateProduct(ctx context.Context, input model.CreateProductRequest) (*model.Product, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Product), args.Error(1)
}

func (m *MockService) GetAllProducts(ctx context.Context, query model.ProductQuery) (*model.ProductListResponse, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.ProductListResponse), args.Error(1)
}

func (m *MockService) GetProductByID(ctx context.Context, id uint) (*model.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Product), args.Error(1)
}

func (m *MockService) UpdateProduct(ctx context.Context, id uint, input model.UpdateProductRequest) (*model.Product, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Product), args.Error(1)
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

func (m *MockService) GetProductsByIDs(ctx context.Context, ids []uint) ([]model.Product, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Product), args.Error(1)
}

func (m *MockService) GetProductBySKU(ctx context.Context, sku string) (*model.Product, error) {
	args := m.Called(ctx, sku)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Product), args.Error(1)
}

func setupHandlerTest() (*gin.Engine, *MockService, *handler.ProductHandler) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	mockService := new(MockService)
	logConfig := logger.Config{
		Level:      "info",
		Format:     "console",
		OutputPath: "",
	}
	log, _ := logger.NewLogger(logConfig, "test")
	productHandler := handler.NewProductHandler(mockService, log.GetZapLogger())
	productHandler.RegisterRoutes(router.Group("/"))
	return router, mockService, productHandler
}

func TestHandler_CreateProduct(t *testing.T) {
	t.Run("should create product successfully", func(t *testing.T) {
		router, mockService, _ := setupHandlerTest()

		input := model.CreateProductRequest{
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		expectedProduct := &model.Product{
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

	t.Run("should return error when service fails", func(t *testing.T) {
		router, mockService, _ := setupHandlerTest()

		input := model.CreateProductRequest{
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

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_GetAllProducts(t *testing.T) {
	t.Run("should get all products successfully", func(t *testing.T) {
		router, mockService, _ := setupHandlerTest()

		query := model.ProductQuery{
			Page:     1,
			PageSize: 10,
		}

		expectedResponse := &model.ProductListResponse{
			Data: []model.Product{
				{ID: 1, Name: "Product 1", Price: 10000, Stock: 50},
				{ID: 2, Name: "Product 2", Price: 20000, Stock: 30},
			},
			Pagination: model.PaginationMetadata{
				Page:       1,
				PageSize:   10,
				Total:      2,
				TotalPages: 1,
			},
		}

		mockService.On("GetAllProducts", mock.Anything, query).Return(expectedResponse, nil)

		req := httptest.NewRequest(http.MethodGet, "/products?page=1&page_size=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when service fails", func(t *testing.T) {
		router, mockService, _ := setupHandlerTest()

		query := model.ProductQuery{
			Page:     1,
			PageSize: 10,
		}

		mockService.On("GetAllProducts", mock.Anything, query).Return(nil, errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/products?page=1&page_size=10", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_GetProductByID(t *testing.T) {
	t.Run("should get product by ID successfully", func(t *testing.T) {
		router, mockService, _ := setupHandlerTest()

		expectedProduct := &model.Product{
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
		router, _, _ := setupHandlerTest()

		req := httptest.NewRequest(http.MethodGet, "/products/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error when product not found", func(t *testing.T) {
		router, mockService, _ := setupHandlerTest()

		mockService.On("GetProductByID", mock.Anything, uint(1)).Return(nil, model.ErrProductNotFound)

		req := httptest.NewRequest(http.MethodGet, "/products/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_UpdateProduct(t *testing.T) {
	t.Run("should update product successfully", func(t *testing.T) {
		router, mockService, _ := setupHandlerTest()

		input := model.UpdateProductRequest{
			Name:  stringPtrHandlerTest("Updated Product"),
			Price: intPtrHandlerTest(15000),
		}

		expectedProduct := &model.Product{
			ID:    1,
			Name:  "Updated Product",
			Price: 15000,
			Stock: 50,
		}

		mockService.On("UpdateProduct", mock.Anything, uint(1), input).Return(expectedProduct, nil)

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPut, "/products/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid ID", func(t *testing.T) {
		router, _, _ := setupHandlerTest()

		input := model.UpdateProductRequest{
			Name: stringPtrHandlerTest("Updated Product"),
		}

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPut, "/products/invalid", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error when product not found", func(t *testing.T) {
		router, mockService, _ := setupHandlerTest()

		input := model.UpdateProductRequest{
			Name: stringPtrHandlerTest("Updated Product"),
		}

		mockService.On("UpdateProduct", mock.Anything, uint(1), input).Return(nil, model.ErrProductNotFound)

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPut, "/products/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when service fails", func(t *testing.T) {
		router, mockService, _ := setupHandlerTest()

		input := model.UpdateProductRequest{
			Name: stringPtrHandlerTest("Updated Product"),
		}

		mockService.On("UpdateProduct", mock.Anything, uint(1), input).Return(nil, errors.New("database error"))

		body, _ := json.Marshal(input)
		req := httptest.NewRequest(http.MethodPut, "/products/1", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

func TestHandler_DeleteProduct(t *testing.T) {
	t.Run("should delete product successfully", func(t *testing.T) {
		router, mockService, _ := setupHandlerTest()

		mockService.On("DeleteProduct", mock.Anything, uint(1)).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/products/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error for invalid ID", func(t *testing.T) {
		router, _, _ := setupHandlerTest()

		req := httptest.NewRequest(http.MethodDelete, "/products/invalid", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("should return error when product not found", func(t *testing.T) {
		router, mockService, _ := setupHandlerTest()

		mockService.On("DeleteProduct", mock.Anything, uint(1)).Return(model.ErrProductNotFound)

		req := httptest.NewRequest(http.MethodDelete, "/products/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("should return error when service fails", func(t *testing.T) {
		router, mockService, _ := setupHandlerTest()

		mockService.On("DeleteProduct", mock.Anything, uint(1)).Return(errors.New("database error"))

		req := httptest.NewRequest(http.MethodDelete, "/products/1", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		mockService.AssertExpectations(t)
	})
}

// Helper functions
func stringPtrHandlerTest(s string) *string {
	return &s
}

func intPtrHandlerTest(i int) *int {
	return &i
}