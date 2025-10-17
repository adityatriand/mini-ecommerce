package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	"mini-e-commerce/services/product/internal/model"
	"mini-e-commerce/services/product/internal/service"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, product *model.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockRepository) FindByID(ctx context.Context, id uint) (model.Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(model.Product), args.Error(1)
}

func (m *MockRepository) FindBySKU(ctx context.Context, sku string) (model.Product, error) {
	args := m.Called(ctx, sku)
	return args.Get(0).(model.Product), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, product *model.Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) FindAll(ctx context.Context, query model.ProductQuery) ([]model.Product, int64, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) FindMultipleByIDs(ctx context.Context, ids []uint) ([]model.Product, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Product), args.Error(1)
}

func (m *MockRepository) UpdateStock(ctx context.Context, id uint, stockDelta int) error {
	args := m.Called(ctx, id, stockDelta)
	return args.Error(0)
}

func (m *MockRepository) UpdateStockOptimizedWithTx(tx *gorm.DB, id uint, stockDelta int) error {
	args := m.Called(tx, id, stockDelta)
	return args.Error(0)
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

func TestService_CreateProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("should create product successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := service.NewProductService(mockRepo, mockCache, logger)

		input := model.CreateProductRequest{
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Product")).Return(nil)
		mockCache.On("DeletePattern", ctx, "product:list:*").Return(nil)

		product, err := service.CreateProduct(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, input.Name, product.Name)
		assert.Equal(t, input.Price, product.Price)
		assert.Equal(t, input.Stock, product.Stock)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := service.NewProductService(mockRepo, mockCache, logger)

		input := model.CreateProductRequest{
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*model.Product")).Return(errors.New("database error"))

		product, err := service.CreateProduct(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, product)
		mockRepo.AssertExpectations(t)
	})
}

func TestService_GetProductByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should get product from cache", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := service.NewProductService(mockRepo, mockCache, logger)

		expectedProduct := &model.Product{
			ID:    1,
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mockCache.On("Get", ctx, "product:id:1", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
			dest := args.Get(2).(*model.Product)
			*dest = *expectedProduct
		})

		product, err := service.GetProductByID(ctx, 1)

		require.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, expectedProduct.Name, product.Name)
		mockCache.AssertExpectations(t)
	})

	t.Run("should get product from database when not in cache", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := service.NewProductService(mockRepo, mockCache, logger)

		expectedProduct := model.Product{
			ID:    1,
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mockCache.On("Get", ctx, "product:id:1", mock.Anything).Return(redis.Nil)
		mockRepo.On("FindByID", ctx, uint(1)).Return(expectedProduct, nil)
		mockCache.On("Set", ctx, "product:id:1", expectedProduct, 5*time.Minute).Return(nil)

		product, err := service.GetProductByID(ctx, 1)

		require.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, expectedProduct.Name, product.Name)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("should return error when product not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := service.NewProductService(mockRepo, mockCache, logger)

		mockCache.On("Get", ctx, "product:id:1", mock.Anything).Return(redis.Nil)
		mockRepo.On("FindByID", ctx, uint(1)).Return(model.Product{}, model.ErrProductNotFound)

		product, err := service.GetProductByID(ctx, 1)

		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Equal(t, model.ErrProductNotFound.Error(), err.Error())
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}

func TestService_GetAllProducts(t *testing.T) {
	ctx := context.Background()

	t.Run("should get all products successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := service.NewProductService(mockRepo, mockCache, logger)

		query := model.ProductQuery{
			Page:     1,
			PageSize: 10,
		}

		expectedProducts := []model.Product{
			{ID: 1, Name: "Product 1", Price: 10000, Stock: 50},
			{ID: 2, Name: "Product 2", Price: 20000, Stock: 30},
		}

		mockCache.On("Get", ctx, "product:list:1:10:::", mock.Anything).Return(redis.Nil)
		mockRepo.On("FindAll", ctx, query).Return(expectedProducts, int64(2), nil)
		mockCache.On("Set", ctx, "product:list:1:10:::", mock.AnythingOfType("model.ProductListResponse"), 2*time.Minute).Return(nil)

		result, err := service.GetAllProducts(ctx, query)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, int64(2), result.Pagination.Total)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := service.NewProductService(mockRepo, mockCache, logger)

		query := model.ProductQuery{
			Page:     1,
			PageSize: 10,
		}

		mockCache.On("Get", ctx, "product:list:1:10:::", mock.Anything).Return(redis.Nil)
		mockRepo.On("FindAll", ctx, query).Return(nil, int64(0), errors.New("database error"))

		result, err := service.GetAllProducts(ctx, query)

		assert.Error(t, err)
		assert.Nil(t, result)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})
}

func TestService_UpdateProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("should update product successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := service.NewProductService(mockRepo, mockCache, logger)

		input := model.UpdateProductRequest{
			Name:  stringPtrServiceTest("Updated Product"),
			Price: intPtrServiceTest(15000),
		}

		existingProduct := model.Product{
			ID:    1,
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(existingProduct, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*model.Product")).Return(nil)
		mockCache.On("Delete", ctx, []string{"product:id:1"}).Return(nil)
		mockCache.On("DeletePattern", ctx, "product:list:*").Return(nil)

		product, err := service.UpdateProduct(ctx, 1, input)

		require.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, "Updated Product", product.Name)
		assert.Equal(t, 15000, product.Price)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("should return error when product not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := service.NewProductService(mockRepo, mockCache, logger)

		input := model.UpdateProductRequest{
			Name: stringPtrServiceTest("Updated Product"),
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(model.Product{}, model.ErrProductNotFound)

		product, err := service.UpdateProduct(ctx, 1, input)

		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Equal(t, model.ErrProductNotFound.Error(), err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestService_DeleteProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete product successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := service.NewProductService(mockRepo, mockCache, logger)

		existingProduct := model.Product{
			ID:    1,
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(existingProduct, nil)
		mockRepo.On("Delete", ctx, uint(1)).Return(nil)
		mockCache.On("Delete", ctx, []string{"product:id:1"}).Return(nil)
		mockCache.On("DeletePattern", ctx, "product:list:*").Return(nil)

		err := service.DeleteProduct(ctx, 1)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := service.NewProductService(mockRepo, mockCache, logger)

		existingProduct := model.Product{
			ID:    1,
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(existingProduct, nil)
		mockRepo.On("Delete", ctx, uint(1)).Return(errors.New("database error"))

		err := service.DeleteProduct(ctx, 1)

		assert.Error(t, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestService_GetProductsByIDs(t *testing.T) {
	ctx := context.Background()

	t.Run("should get products by IDs successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := service.NewProductService(mockRepo, mockCache, logger)

		ids := []uint{1, 2}
		expectedProducts := []model.Product{
			{ID: 1, Name: "Product 1", Price: 10000, Stock: 50},
			{ID: 2, Name: "Product 2", Price: 20000, Stock: 30},
		}

		// Mock cache misses for both products
		mockCache.On("Get", ctx, "product:id:1", mock.Anything).Return(redis.Nil)
		mockCache.On("Get", ctx, "product:id:2", mock.Anything).Return(redis.Nil)
		mockRepo.On("FindMultipleByIDs", ctx, ids).Return(expectedProducts, nil)
		// Mock cache sets for both products
		mockCache.On("Set", ctx, "product:id:1", expectedProducts[0], 5*time.Minute).Return(nil)
		mockCache.On("Set", ctx, "product:id:2", expectedProducts[1], 5*time.Minute).Return(nil)

		products, err := service.GetProductsByIDs(ctx, ids)

		require.NoError(t, err)
		assert.Len(t, products, 2)
		assert.Equal(t, "Product 1", products[0].Name)
		assert.Equal(t, "Product 2", products[1].Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when repository fails", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := service.NewProductService(mockRepo, mockCache, logger)

		ids := []uint{1, 2}

		// Mock cache misses for both products
		mockCache.On("Get", ctx, "product:id:1", mock.Anything).Return(redis.Nil)
		mockCache.On("Get", ctx, "product:id:2", mock.Anything).Return(redis.Nil)
		mockRepo.On("FindMultipleByIDs", ctx, ids).Return(nil, errors.New("database error"))

		products, err := service.GetProductsByIDs(ctx, ids)

		assert.Error(t, err)
		assert.Nil(t, products)
		mockRepo.AssertExpectations(t)
	})
}

// Helper functions
func stringPtrServiceTest(s string) *string {
	return &s
}

func intPtrServiceTest(i int) *int {
	return &i
}