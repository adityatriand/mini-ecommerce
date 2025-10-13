package order

import (
	"context"
	"errors"
	"testing"

	"mini-e-commerce/internal/dto"
	"mini-e-commerce/internal/logger"
	"mini-e-commerce/internal/product"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"
	"gorm.io/gorm"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, order *Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockRepository) CreateWithTransaction(ctx context.Context, order *Order, txFunc func(*gorm.DB) error) error {
	args := m.Called(ctx, order, txFunc)
	if txFunc != nil {
		if err := txFunc(nil); err != nil {
			return err
		}
	}
	return args.Error(0)
}

func (m *MockRepository) FindAll(ctx context.Context) ([]Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Order), args.Error(1)
}

func (m *MockRepository) FindAllWithPagination(ctx context.Context, offset, limit int, sortBy, order string) ([]Order, int64, error) {
	args := m.Called(ctx, offset, limit, sortBy, order)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) FindByID(ctx context.Context, id uint) (Order, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Order), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, order *Order, updateFn func(*Order)) error {
	args := m.Called(ctx, order, updateFn)
	if updateFn != nil {
		updateFn(order)
	}
	return args.Error(0)
}

func (m *MockRepository) UpdateWithTransaction(ctx context.Context, order *Order, updateFn func(*Order), txFunc func(*gorm.DB) error) error {
	args := m.Called(ctx, order, updateFn, txFunc)
	if txFunc != nil {
		if err := txFunc(nil); err != nil {
			return err
		}
	}
	if updateFn != nil {
		updateFn(order)
	}
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) DeleteWithTransaction(ctx context.Context, id uint, txFunc func(*gorm.DB) error) error {
	args := m.Called(ctx, id, txFunc)
	if txFunc != nil {
		if err := txFunc(nil); err != nil {
			return err
		}
	}
	return args.Error(0)
}

type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) CreateProduct(ctx context.Context, input product.CreateProductRequest) (*product.Product, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

func (m *MockProductService) GetAllProducts(ctx context.Context) ([]product.Product, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]product.Product), args.Error(1)
}

func (m *MockProductService) GetAllProductsWithQuery(ctx context.Context, query product.ProductQuery) (*product.ProductListResponse, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.ProductListResponse), args.Error(1)
}

func (m *MockProductService) GetProductByID(ctx context.Context, id uint) (*product.Product, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

func (m *MockProductService) UpdateProduct(ctx context.Context, id uint, input product.UpdateProductRequest) (*product.Product, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*product.Product), args.Error(1)
}

func (m *MockProductService) DeleteProduct(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProductService) UpdateStock(ctx context.Context, id uint, stockDelta int) error {
	args := m.Called(ctx, id, stockDelta)
	return args.Error(0)
}

func (m *MockProductService) UpdateStockWithTx(tx *gorm.DB, id uint, stockDelta int) error {
	args := m.Called(tx, id, stockDelta)
	return args.Error(0)
}

func (m *MockProductService) GetProductsByIDs(ctx context.Context, ids []uint) ([]product.Product, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]product.Product), args.Error(1)
}

func setupServiceTest() (*MockRepository, *MockProductService, Service) {
	mockRepo := new(MockRepository)
	mockProductService := new(MockProductService)
	log, _ := logger.NewLogger(&logger.Config{
		ServiceName: "test",
		AppVersion:  "1.0",
		LogLevel:    zapcore.InfoLevel,
		Mode:        "development",
	})
	service := NewService(mockRepo, mockProductService, log)
	return mockRepo, mockProductService, service
}

func TestService_CreateOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("should create order successfully", func(t *testing.T) {
		mockRepo, mockProductService, service := setupServiceTest()

		input := CreateOrderRequest{
			Items: []OrderItemInput{
				{ProductID: 1, Quantity: 2},
				{ProductID: 2, Quantity: 1},
			},
		}

		mockProduct1 := &product.Product{ID: 1, Name: "Product 1", Price: 10000, Stock: 50}
		mockProduct2 := &product.Product{ID: 2, Name: "Product 2", Price: 20000, Stock: 30}

		mockProductService.On("GetProductByID", ctx, uint(1)).Return(mockProduct1, nil).Once()
		mockProductService.On("GetProductByID", ctx, uint(2)).Return(mockProduct2, nil).Once()
		mockProductService.On("GetProductByID", ctx, uint(1)).Return(mockProduct1, nil).Once()
		mockProductService.On("GetProductByID", ctx, uint(2)).Return(mockProduct2, nil).Once()
		mockProductService.On("UpdateStockWithTx", mock.Anything, uint(1), -2).Return(nil)
		mockProductService.On("UpdateStockWithTx", mock.Anything, uint(2), -1).Return(nil)
		mockRepo.On("CreateWithTransaction", ctx, mock.AnythingOfType("*order.Order"), mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)

		order, err := service.CreateOrder(ctx, input, 1)

		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, uint(1), order.UserID)
		assert.Equal(t, 40000, order.TotalPrice)
		assert.Equal(t, StatusPending, order.Status)
		assert.Len(t, order.OrderItems, 2)
		mockRepo.AssertExpectations(t)
		mockProductService.AssertExpectations(t)
	})

	t.Run("should return error for invalid input", func(t *testing.T) {
		_, _, service := setupServiceTest()

		input := CreateOrderRequest{
			Items: []OrderItemInput{},
		}

		order, err := service.CreateOrder(ctx, input, 1)

		assert.Error(t, err)
		assert.Nil(t, order)
	})

	t.Run("should return error when user ID is zero", func(t *testing.T) {
		_, _, service := setupServiceTest()

		input := CreateOrderRequest{
			Items: []OrderItemInput{
				{ProductID: 1, Quantity: 2},
			},
		}

		order, err := service.CreateOrder(ctx, input, 0)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "user ID is required")
	})

	t.Run("should return error when product not found", func(t *testing.T) {
		_, mockProductService, service := setupServiceTest()

		input := CreateOrderRequest{
			Items: []OrderItemInput{
				{ProductID: 999, Quantity: 1},
			},
		}

		mockProductService.On("GetProductByID", ctx, uint(999)).Return(nil, errors.New("product not found"))

		order, err := service.CreateOrder(ctx, input, 1)

		assert.Error(t, err)
		assert.Nil(t, order)
		mockProductService.AssertExpectations(t)
	})

	t.Run("should return error when insufficient stock", func(t *testing.T) {
		_, mockProductService, service := setupServiceTest()

		input := CreateOrderRequest{
			Items: []OrderItemInput{
				{ProductID: 1, Quantity: 100},
			},
		}

		mockProduct := &product.Product{ID: 1, Name: "Product 1", Price: 10000, Stock: 50}
		mockProductService.On("GetProductByID", ctx, uint(1)).Return(mockProduct, nil).Twice()

		order, err := service.CreateOrder(ctx, input, 1)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "insufficient stock")
		mockProductService.AssertExpectations(t)
	})
}

func TestService_GetAllOrders(t *testing.T) {
	ctx := context.Background()

	t.Run("should get all orders successfully", func(t *testing.T) {
		mockRepo, _, service := setupServiceTest()

		expectedOrders := []Order{
			{ID: 1, UserID: 1, TotalPrice: 50000, Status: StatusPending},
			{ID: 2, UserID: 2, TotalPrice: 75000, Status: StatusPaid},
		}

		mockRepo.On("FindAll", ctx).Return(expectedOrders, nil)

		orders, err := service.GetAllOrders(ctx)

		require.NoError(t, err)
		assert.Len(t, orders, 2)
		mockRepo.AssertExpectations(t)
	})
}

func TestService_GetOrderByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get order by ID successfully", func(t *testing.T) {
		mockRepo, _, service := setupServiceTest()

		expectedOrder := Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 50000,
			Status:     StatusPending,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(expectedOrder, nil)

		order, err := service.GetOrderByID(ctx, 1)

		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, uint(1), order.ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when order not found", func(t *testing.T) {
		mockRepo, _, service := setupServiceTest()

		mockRepo.On("FindByID", ctx, uint(999)).Return(Order{}, gorm.ErrRecordNotFound)

		order, err := service.GetOrderByID(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "order not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestService_UpdateOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("should update order status to PAID successfully", func(t *testing.T) {
		mockRepo, _, service := setupServiceTest()

		existingOrder := Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 50000,
			Status:     StatusPending,
		}

		newStatus := StatusPaid
		input := UpdateOrderRequest{
			Status: &newStatus,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(existingOrder, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*order.Order"), mock.AnythingOfType("func(*order.Order)")).Return(nil)

		order, err := service.UpdateOrder(ctx, 1, input, 1)

		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, StatusPaid, order.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should cancel order and restore stock successfully", func(t *testing.T) {
		mockRepo, mockProductService, service := setupServiceTest()

		existingOrder := Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 50000,
			Status:     StatusPending,
			OrderItems: []OrderItem{
				{ID: 1, ProductID: 1, Quantity: 2, Price: 10000, Subtotal: 20000},
				{ID: 2, ProductID: 2, Quantity: 1, Price: 30000, Subtotal: 30000},
			},
		}

		newStatus := StatusCancelled
		input := UpdateOrderRequest{
			Status: &newStatus,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(existingOrder, nil)
		mockProductService.On("UpdateStockWithTx", mock.Anything, uint(1), 2).Return(nil)
		mockProductService.On("UpdateStockWithTx", mock.Anything, uint(2), 1).Return(nil)
		mockRepo.On("UpdateWithTransaction", ctx, mock.AnythingOfType("*order.Order"), mock.AnythingOfType("func(*order.Order)"), mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)

		order, err := service.UpdateOrder(ctx, 1, input, 1)

		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, StatusCancelled, order.Status)
		mockRepo.AssertExpectations(t)
		mockProductService.AssertExpectations(t)
	})

	t.Run("should return error when order not found", func(t *testing.T) {
		mockRepo, _, service := setupServiceTest()

		newStatus := StatusPaid
		input := UpdateOrderRequest{
			Status: &newStatus,
		}

		mockRepo.On("FindByID", ctx, uint(999)).Return(Order{}, gorm.ErrRecordNotFound)

		order, err := service.UpdateOrder(ctx, 999, input, 1)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "order not found")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when user not authorized", func(t *testing.T) {
		mockRepo, _, service := setupServiceTest()

		existingOrder := Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 50000,
			Status:     StatusPending,
		}

		newStatus := StatusPaid
		input := UpdateOrderRequest{
			Status: &newStatus,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(existingOrder, nil)

		order, err := service.UpdateOrder(ctx, 1, input, 999)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "not authorized")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when trying to change paid order to pending", func(t *testing.T) {
		mockRepo, _, service := setupServiceTest()

		existingOrder := Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 50000,
			Status:     StatusPaid,
		}

		newStatus := StatusPending
		input := UpdateOrderRequest{
			Status: &newStatus,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(existingOrder, nil)

		order, err := service.UpdateOrder(ctx, 1, input, 1)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "cannot change paid order back to pending")
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when trying to change cancelled order", func(t *testing.T) {
		mockRepo, _, service := setupServiceTest()

		existingOrder := Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 50000,
			Status:     StatusCancelled,
		}

		newStatus := StatusPaid
		input := UpdateOrderRequest{
			Status: &newStatus,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(existingOrder, nil)

		order, err := service.UpdateOrder(ctx, 1, input, 1)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "cannot change cancelled order status")
		mockRepo.AssertExpectations(t)
	})
}

func TestService_DeleteOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete order and restore stock successfully", func(t *testing.T) {
		mockRepo, mockProductService, service := setupServiceTest()

		existingOrder := Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 50000,
			Status:     StatusPending,
			OrderItems: []OrderItem{
				{ID: 1, ProductID: 1, Quantity: 2, Price: 10000, Subtotal: 20000},
				{ID: 2, ProductID: 2, Quantity: 1, Price: 30000, Subtotal: 30000},
			},
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(existingOrder, nil)
		mockProductService.On("UpdateStockWithTx", mock.Anything, uint(1), 2).Return(nil)
		mockProductService.On("UpdateStockWithTx", mock.Anything, uint(2), 1).Return(nil)
		mockRepo.On("DeleteWithTransaction", ctx, uint(1), mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)

		err := service.DeleteOrder(ctx, 1)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockProductService.AssertExpectations(t)
	})

	t.Run("should return error when order not found", func(t *testing.T) {
		mockRepo, _, service := setupServiceTest()

		mockRepo.On("FindByID", ctx, uint(999)).Return(Order{}, gorm.ErrRecordNotFound)

		err := service.DeleteOrder(ctx, 999)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "order not found")
		mockRepo.AssertExpectations(t)
	})
}

func TestService_GetAllOrdersWithQuery(t *testing.T) {
	ctx := context.Background()

	t.Run("should get orders with pagination successfully", func(t *testing.T) {
		mockRepo, _, service := setupServiceTest()

		expectedOrders := []Order{
			{ID: 1, UserID: 1, TotalPrice: 50000, Status: StatusPending},
			{ID: 2, UserID: 2, TotalPrice: 75000, Status: StatusPaid},
		}

		query := OrderQuery{
			PaginationQuery: dto.PaginationQuery{
				Page:     1,
				PageSize: 10,
				Order:    "asc",
			},
			SortBy: "user_id",
		}

		mockRepo.On("FindAllWithPagination", ctx, 0, 10, "user_id", "asc").Return(expectedOrders, int64(2), nil)

		result, err := service.GetAllOrdersWithQuery(ctx, query)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, 1, result.Pagination.Page)
		assert.Equal(t, 10, result.Pagination.PageSize)
		assert.Equal(t, int64(2), result.Pagination.Total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should use default values for pagination", func(t *testing.T) {
		mockRepo, _, service := setupServiceTest()

		expectedOrders := []Order{
			{ID: 1, UserID: 1, TotalPrice: 50000, Status: StatusPending},
		}

		query := OrderQuery{}

		mockRepo.On("FindAllWithPagination", ctx, 0, 10, "", "desc").Return(expectedOrders, int64(1), nil)

		result, err := service.GetAllOrdersWithQuery(ctx, query)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Pagination.Page)
		assert.Equal(t, 10, result.Pagination.PageSize)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should limit page size to 100", func(t *testing.T) {
		mockRepo, _, service := setupServiceTest()

		query := OrderQuery{
			PaginationQuery: dto.PaginationQuery{
				Page:     1,
				PageSize: 200,
			},
		}

		mockRepo.On("FindAllWithPagination", ctx, 0, 100, "", "desc").Return([]Order{}, int64(0), nil)

		result, err := service.GetAllOrdersWithQuery(ctx, query)

		require.NoError(t, err)
		assert.Equal(t, 100, result.Pagination.PageSize)
		mockRepo.AssertExpectations(t)
	})
}

func TestService_CreateOrderOptimized(t *testing.T) {
	tests := []struct {
		name                string
		input               CreateOrderRequest
		userID              uint
		mockProductsReturn  []product.Product
		mockProductsError   error
		mockRepoError       error
		expectedError       bool
		expectedErrorMsg    string
	}{
		{
			name: "should create order successfully with optimized batch product fetching",
			input: CreateOrderRequest{
				Items: []OrderItemInput{
					{ProductID: 1, Quantity: 2},
					{ProductID: 2, Quantity: 1},
				},
			},
			userID: 1,
			mockProductsReturn: []product.Product{
				{ID: 1, Name: "Product 1", Price: 100, Stock: 10},
				{ID: 2, Name: "Product 2", Price: 200, Stock: 5},
			},
			mockProductsError: nil,
			mockRepoError:     nil,
			expectedError:     false,
		},
		{
			name: "should return error when product not found",
			input: CreateOrderRequest{
				Items: []OrderItemInput{
					{ProductID: 999, Quantity: 1},
				},
			},
			userID:             1,
			mockProductsReturn: []product.Product{},
			mockProductsError:  nil,
			mockRepoError:      nil,
			expectedError:      true,
			expectedErrorMsg:   "product not found",
		},
		{
			name: "should return error when insufficient stock",
			input: CreateOrderRequest{
				Items: []OrderItemInput{
					{ProductID: 1, Quantity: 100},
				},
			},
			userID: 1,
			mockProductsReturn: []product.Product{
				{ID: 1, Name: "Product 1", Price: 100, Stock: 10},
			},
			mockProductsError: nil,
			mockRepoError:     nil,
			expectedError:     true,
			expectedErrorMsg:   "insufficient stock",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo, mockProductService, service := setupServiceTest()

			mockProductService.On("GetProductsByIDs", mock.Anything, mock.AnythingOfType("[]uint")).Return(tt.mockProductsReturn, tt.mockProductsError)

			if !tt.expectedError {
				mockRepo.On("CreateWithTransaction", mock.Anything, mock.AnythingOfType("*order.Order"), mock.AnythingOfType("func(*gorm.DB) error")).Return(tt.mockRepoError)
				mockProductService.On("UpdateStockWithTx", mock.Anything, mock.AnythingOfType("uint"), mock.AnythingOfType("int")).Return(nil)
			}

			result, err := service.CreateOrderOptimized(context.Background(), tt.input, tt.userID)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
				if tt.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), tt.expectedErrorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.userID, result.UserID)
				assert.Equal(t, StatusPending, result.Status)
			}

			mockRepo.AssertExpectations(t)
			mockProductService.AssertExpectations(t)
		})
	}
}
