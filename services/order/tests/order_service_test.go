package tests

import (
	"context"
	"errors"
	"testing"

	"mini-e-commerce/services/order/internal/model"
	"mini-e-commerce/services/order/internal/service"
	"mini-e-commerce/shared/config"
	"mini-e-commerce/shared/logger"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// Local product model for testing
type Product struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Price int    `json:"price"`
	Stock int    `json:"stock"`
}

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockRepository) FindByID(ctx context.Context, id uint) (model.Order, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Order), args.Error(1)
}

func (m *MockRepository) FindAll(ctx context.Context, query model.OrderQuery) ([]model.Order, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) Update(ctx context.Context, order *model.Order) error {
	args := m.Called(ctx, order)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) FindByUserID(ctx context.Context, userID uint, query model.OrderQuery) ([]model.Order, int64, error) {
	args := m.Called(ctx, userID, query)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Order), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) CreateWithTransaction(ctx context.Context, order *model.Order, fn func(tx *gorm.DB) error) error {
	args := m.Called(ctx, order, fn)
	return args.Error(0)
}

func (m *MockRepository) UpdateWithTransaction(ctx context.Context, order *model.Order, updateFn func(*model.Order), fn func(tx *gorm.DB) error) error {
	args := m.Called(ctx, order, updateFn, fn)
	return args.Error(0)
}

func (m *MockRepository) DeleteWithTransaction(ctx context.Context, id uint, fn func(tx *gorm.DB) error) error {
	args := m.Called(ctx, id, fn)
	return args.Error(0)
}

type MockProductService struct {
	mock.Mock
}

func (m *MockProductService) GetProductsByIDs(ctx context.Context, ids []uint) ([]Product, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Product), args.Error(1)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string, dest any) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	args := m.Called(ctx, key, value, ttl)
	return args.Error(0)
}

func (m *MockCache) Delete(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func (m *MockCache) DeletePattern(ctx context.Context, pattern string) error {
	args := m.Called(ctx, pattern)
	return args.Error(0)
}

func (m *MockCache) Ping(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func setupOrderServiceTest() (*MockRepository, *MockProductService, service.OrderService) {
	mockRepo := new(MockRepository)
	mockProductService := new(MockProductService)
	mockCache := new(MockCache)
	logConfig := logger.Config{
		Level:      "info",
		Format:     "console",
		OutputPath: "",
	}
	log, _ := logger.NewLogger(logConfig, "test")
	serviceConfig := &config.ServiceConfig{
		Services: config.ServicesConfig{
			ProductService: config.ServiceEndpoint{
				URL: "http://localhost:8081",
			},
		},
	}
	service := service.NewOrderService(mockRepo, mockCache, log.GetZapLogger(), serviceConfig)
	return mockRepo, mockProductService, service
}

func TestOrderService_CreateOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("should return error when product service unavailable", func(t *testing.T) {
		_, _, service := setupOrderServiceTest()

		input := model.CreateOrderRequest{
			Items: []model.OrderItemRequest{
				{ProductID: 1, Quantity: 2},
			},
		}

		order, err := service.CreateOrder(ctx, input, 1)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Equal(t, model.ErrProductServiceUnavailable.Error(), err.Error())
	})

	t.Run("should return error when product service unavailable for non-existent product", func(t *testing.T) {
		_, _, service := setupOrderServiceTest()

		input := model.CreateOrderRequest{
			Items: []model.OrderItemRequest{
				{ProductID: 999, Quantity: 1},
			},
		}

		order, err := service.CreateOrder(ctx, input, 1)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Equal(t, model.ErrProductServiceUnavailable.Error(), err.Error())
	})

	t.Run("should return error when product service unavailable for insufficient stock test", func(t *testing.T) {
		_, _, service := setupOrderServiceTest()

		input := model.CreateOrderRequest{
			Items: []model.OrderItemRequest{
				{ProductID: 1, Quantity: 100},
			},
		}

		order, err := service.CreateOrder(ctx, input, 1)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Equal(t, model.ErrProductServiceUnavailable.Error(), err.Error())
	})
}

func TestOrderService_GetOrderByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get order by ID successfully", func(t *testing.T) {
		mockRepo, _, service := setupOrderServiceTest()

		expectedOrder := model.Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 20000,
			Status:     model.StatusPending,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(expectedOrder, nil)

		order, err := service.GetOrderByID(ctx, 1, 1)

		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, uint(1), order.ID)
		assert.Equal(t, uint(1), order.UserID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when order not found", func(t *testing.T) {
		mockRepo, _, service := setupOrderServiceTest()

		mockRepo.On("FindByID", ctx, uint(999)).Return(model.Order{}, model.ErrOrderNotFound)

		order, err := service.GetOrderByID(ctx, 999, 1)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Equal(t, model.ErrOrderNotFound.Error(), err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestOrderService_GetAllOrders(t *testing.T) {
	ctx := context.Background()

	t.Run("should get all orders successfully", func(t *testing.T) {
		mockRepo, _, service := setupOrderServiceTest()

		query := model.OrderQuery{
			Page:     1,
			PageSize: 10,
		}

		expectedOrders := []model.Order{
			{ID: 1, UserID: 1, TotalPrice: 20000, Status: model.StatusPending},
			{ID: 2, UserID: 2, TotalPrice: 30000, Status: model.StatusPaid},
		}


		mockRepo.On("FindAll", ctx, query).Return(expectedOrders, int64(2), nil)

		result, err := service.GetAllOrders(ctx, query, 1)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(2), result.Pagination.Total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		mockRepo, _, service := setupOrderServiceTest()

		query := model.OrderQuery{
			Page:     1,
			PageSize: 10,
		}

		mockRepo.On("FindAll", ctx, query).Return(nil, int64(0), errors.New("database error"))

		result, err := service.GetAllOrders(ctx, query, 1)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}

func TestOrderService_UpdateOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("should update order successfully", func(t *testing.T) {
		mockRepo, _, service := setupOrderServiceTest()

		input := model.UpdateOrderRequest{
			Status: &[]model.OrderStatus{model.StatusPaid}[0],
		}

		existingOrder := model.Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 20000,
			Status:     model.StatusPending,
		}


		mockRepo.On("FindByID", ctx, uint(1)).Return(existingOrder, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*model.Order")).Return(nil)

		order, err := service.UpdateOrder(ctx, 1, input, 1)

		require.NoError(t, err)
		assert.NotNil(t, order)
		assert.Equal(t, model.StatusPaid, order.Status)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when order not found", func(t *testing.T) {
		mockRepo, _, service := setupOrderServiceTest()

		input := model.UpdateOrderRequest{
			Status: &[]model.OrderStatus{model.StatusPaid}[0],
		}

		mockRepo.On("FindByID", ctx, uint(999)).Return(model.Order{}, model.ErrOrderNotFound)

		order, err := service.UpdateOrder(ctx, 999, input, 1)

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Equal(t, model.ErrOrderNotFound.Error(), err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when not authorized", func(t *testing.T) {
		mockRepo, _, service := setupOrderServiceTest()

		input := model.UpdateOrderRequest{
			Status: &[]model.OrderStatus{model.StatusPaid}[0],
		}

		existingOrder := model.Order{
			ID:         1,
			UserID:     2, // Different user
			TotalPrice: 20000,
			Status:     model.StatusPending,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(existingOrder, nil)

		order, err := service.UpdateOrder(ctx, 1, input, 1) // User 1 trying to update User 2's order

		assert.Error(t, err)
		assert.Nil(t, order)
		assert.Equal(t, model.ErrNotAuthorizedToUpdate.Error(), err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestOrderService_DeleteOrder(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete order successfully", func(t *testing.T) {
		mockRepo, _, service := setupOrderServiceTest()

		existingOrder := model.Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 20000,
			Status:     model.StatusPending,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(existingOrder, nil)
		mockRepo.On("DeleteWithTransaction", ctx, uint(1), mock.AnythingOfType("func(*gorm.DB) error")).Return(nil)

		err := service.DeleteOrder(ctx, 1, 1)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when order not found", func(t *testing.T) {
		mockRepo, _, service := setupOrderServiceTest()

		mockRepo.On("FindByID", ctx, uint(999)).Return(model.Order{}, model.ErrOrderNotFound)

		err := service.DeleteOrder(ctx, 999, 1)

		assert.Error(t, err)
		assert.Equal(t, model.ErrOrderNotFound.Error(), err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when not authorized", func(t *testing.T) {
		mockRepo, _, service := setupOrderServiceTest()

		existingOrder := model.Order{
			ID:         1,
			UserID:     2, // Different user
			TotalPrice: 20000,
			Status:     model.StatusPending,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(existingOrder, nil)

		err := service.DeleteOrder(ctx, 1, 1) // User 1 trying to delete User 2's order

		assert.Error(t, err)
		assert.Equal(t, model.ErrNotAuthorizedToUpdate.Error(), err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestOrderService_GetOrdersByUserID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get orders by user ID successfully", func(t *testing.T) {
		mockRepo, _, service := setupOrderServiceTest()

		query := model.OrderQuery{
			Page:     1,
			PageSize: 10,
		}

		expectedOrders := []model.Order{
			{ID: 1, UserID: 1, TotalPrice: 20000, Status: model.StatusPending},
			{ID: 2, UserID: 1, TotalPrice: 30000, Status: model.StatusPaid},
		}


		mockRepo.On("FindByUserID", ctx, uint(1), query).Return(expectedOrders, int64(2), nil)

		result, err := service.GetOrdersByUserID(ctx, 1, query)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(2), result.Pagination.Total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		mockRepo, _, service := setupOrderServiceTest()

		query := model.OrderQuery{
			Page:     1,
			PageSize: 10,
		}

		mockRepo.On("FindByUserID", ctx, uint(1), query).Return(nil, int64(0), errors.New("database error"))

		result, err := service.GetOrdersByUserID(ctx, 1, query)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
	})
}