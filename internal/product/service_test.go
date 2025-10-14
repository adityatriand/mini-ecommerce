package product

import (
	"context"
	"errors"
	"testing"
	"time"

	"mini-e-commerce/internal/dto"

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

func (m *MockRepository) Create(ctx context.Context, product *Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockRepository) FindAll(ctx context.Context) ([]Product, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Product), args.Error(1)
}

func (m *MockRepository) FindAllWithPagination(ctx context.Context, offset, limit int, sortBy, order string) ([]Product, int64, error) {
	args := m.Called(ctx, offset, limit, sortBy, order)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]Product), args.Get(1).(int64), args.Error(2)
}

func (m *MockRepository) FindByID(ctx context.Context, id uint) (Product, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(Product), args.Error(1)
}

func (m *MockRepository) Update(ctx context.Context, product *Product) error {
	args := m.Called(ctx, product)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRepository) UpdateStockOptimizedWithTx(tx *gorm.DB, id uint, stockDelta int) error {
	args := m.Called(tx, id, stockDelta)
	return args.Error(0)
}

func (m *MockRepository) FindMultipleByIDs(ctx context.Context, ids []uint) ([]Product, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Product), args.Error(1)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) Get(ctx context.Context, key string, dest interface{}) error {
	args := m.Called(ctx, key, dest)
	return args.Error(0)
}

func (m *MockCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
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

func TestService_CreateProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("should create product successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		input := CreateProductRequest{
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*product.Product")).Return(nil)
		mockCache.On("DeletePattern", ctx, "product:list:*").Return(nil)

		product, err := service.CreateProduct(ctx, input)

		require.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, input.Name, product.Name)
		assert.Equal(t, input.Price, product.Price)
		assert.Equal(t, input.Stock, product.Stock)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error for invalid input", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		input := CreateProductRequest{
			Name:  "",
			Price: -100,
			Stock: -10,
		}

		product, err := service.CreateProduct(ctx, input)

		assert.Error(t, err)
		assert.Nil(t, product)
	})

	t.Run("should return error when repository create fails", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		input := CreateProductRequest{
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mockRepo.On("Create", ctx, mock.AnythingOfType("*product.Product")).Return(errors.New("database error"))

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

		service := NewService(mockRepo, mockCache, logger)

		mockCache.On("Get", ctx, "product:id:1", mock.Anything).Return(nil)

		product, err := service.GetProductByID(ctx, 1)

		require.NoError(t, err)
		assert.NotNil(t, product)
		mockCache.AssertExpectations(t)
	})

	t.Run("should get product from database when cache misses", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		expectedProduct := Product{
			ID:    1,
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mockCache.On("Get", ctx, "product:id:1", mock.Anything).Return(redis.Nil)
		mockRepo.On("FindByID", ctx, uint(1)).Return(expectedProduct, nil)
		mockCache.On("Set", ctx, "product:id:1", expectedProduct, CacheTTLProduct).Return(nil)

		product, err := service.GetProductByID(ctx, 1)

		require.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, uint(1), product.ID)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("should return error when product not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		mockCache.On("Get", ctx, "product:id:999", mock.Anything).Return(redis.Nil)
		mockRepo.On("FindByID", ctx, uint(999)).Return(Product{}, gorm.ErrRecordNotFound)

		product, err := service.GetProductByID(ctx, 999)

		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Equal(t, ErrProductNotFound, err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestService_GetAllProducts(t *testing.T) {
	ctx := context.Background()

	t.Run("should get all products successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		expectedProducts := []Product{
			{ID: 1, Name: "Product 1", Price: 10000, Stock: 50},
			{ID: 2, Name: "Product 2", Price: 20000, Stock: 30},
		}

		mockRepo.On("FindAll", ctx).Return(expectedProducts, nil)

		products, err := service.GetAllProducts(ctx)

		require.NoError(t, err)
		assert.Len(t, products, 2)
		mockRepo.AssertExpectations(t)
	})
}

func TestService_UpdateProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("should update product successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		existingProduct := Product{
			ID:    1,
			Name:  "Old Product",
			Price: 10000,
			Stock: 50,
		}

		newName := "Updated Product"
		newPrice := 15000

		input := UpdateProductRequest{
			Name:  &newName,
			Price: &newPrice,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(existingProduct, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*product.Product")).Return(nil)
		mockCache.On("Delete", ctx, []string{"product:id:1"}).Return(nil)
		mockCache.On("DeletePattern", ctx, "product:list:*").Return(nil)

		product, err := service.UpdateProduct(ctx, 1, input)

		require.NoError(t, err)
		assert.NotNil(t, product)
		assert.Equal(t, newName, product.Name)
		assert.Equal(t, newPrice, product.Price)
		assert.Equal(t, 50, product.Stock)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when product not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		newName := "Updated Product"
		input := UpdateProductRequest{
			Name: &newName,
		}

		mockRepo.On("FindByID", ctx, uint(999)).Return(Product{}, gorm.ErrRecordNotFound)

		product, err := service.UpdateProduct(ctx, 999, input)

		assert.Error(t, err)
		assert.Nil(t, product)
		assert.Equal(t, ErrProductNotFound, err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestService_DeleteProduct(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete product successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		existingProduct := Product{
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
	})

	t.Run("should return error when product not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		mockRepo.On("FindByID", ctx, uint(999)).Return(Product{}, gorm.ErrRecordNotFound)

		err := service.DeleteProduct(ctx, 999)

		assert.Error(t, err)
		assert.Equal(t, ErrProductNotFound, err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestService_UpdateStock(t *testing.T) {
	ctx := context.Background()

	t.Run("should update stock successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		existingProduct := Product{
			ID:    1,
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(existingProduct, nil)
		mockRepo.On("Update", ctx, mock.AnythingOfType("*product.Product")).Return(nil)
		mockCache.On("Delete", ctx, []string{"product:id:1"}).Return(nil)
		mockCache.On("DeletePattern", ctx, "product:list:*").Return(nil)

		err := service.UpdateStock(ctx, 1, -10)

		require.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error for insufficient stock", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		existingProduct := Product{
			ID:    1,
			Name:  "Test Product",
			Price: 10000,
			Stock: 5,
		}

		mockRepo.On("FindByID", ctx, uint(1)).Return(existingProduct, nil)

		err := service.UpdateStock(ctx, 1, -10)

		assert.Error(t, err)
		assert.Equal(t, "insufficient stock", err.Error())
		mockRepo.AssertExpectations(t)
	})

	t.Run("should return error when product not found", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		mockRepo.On("FindByID", ctx, uint(999)).Return(Product{}, gorm.ErrRecordNotFound)

		err := service.UpdateStock(ctx, 999, -10)

		assert.Error(t, err)
		assert.Equal(t, ErrProductNotFound, err.Error())
		mockRepo.AssertExpectations(t)
	})
}

func TestService_GetAllProductsWithQuery(t *testing.T) {
	ctx := context.Background()

	t.Run("should get products with pagination successfully", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		expectedProducts := []Product{
			{ID: 1, Name: "Product 1", Price: 10000, Stock: 50},
			{ID: 2, Name: "Product 2", Price: 20000, Stock: 30},
		}

		query := ProductQuery{
			PaginationQuery: dto.PaginationQuery{
				Page:     1,
				PageSize: 10,
				Order:    "asc",
			},
			SortBy: "name",
		}

		mockCache.On("Get", ctx, "product:list:1:10:name:asc", mock.Anything).Return(redis.Nil)
		mockRepo.On("FindAllWithPagination", ctx, 0, 10, "name", "asc").Return(expectedProducts, int64(2), nil)
		mockCache.On("Set", ctx, "product:list:1:10:name:asc", mock.Anything, CacheTTLProductList).Return(nil)

		result, err := service.GetAllProductsWithQuery(ctx, query)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result.Data, 2)
		assert.Equal(t, 1, result.Pagination.Page)
		assert.Equal(t, 10, result.Pagination.PageSize)
		assert.Equal(t, int64(2), result.Pagination.Total)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should use default values for pagination", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		expectedProducts := []Product{
			{ID: 1, Name: "Product 1", Price: 10000, Stock: 50},
		}

		query := ProductQuery{}

		mockCache.On("Get", ctx, "product:list:1:10::desc", mock.Anything).Return(redis.Nil)
		mockRepo.On("FindAllWithPagination", ctx, 0, 10, "", "desc").Return(expectedProducts, int64(1), nil)
		mockCache.On("Set", ctx, "product:list:1:10::desc", mock.Anything, CacheTTLProductList).Return(nil)

		result, err := service.GetAllProductsWithQuery(ctx, query)

		require.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, result.Pagination.Page)
		assert.Equal(t, 10, result.Pagination.PageSize)
		mockRepo.AssertExpectations(t)
	})

	t.Run("should limit page size to 100", func(t *testing.T) {
		mockRepo := new(MockRepository)
		mockCache := new(MockCache)
		logger := zap.NewNop()

		service := NewService(mockRepo, mockCache, logger)

		query := ProductQuery{
			PaginationQuery: dto.PaginationQuery{
				Page:     1,
				PageSize: 200,
			},
		}

		mockCache.On("Get", ctx, "product:list:1:100::desc", mock.Anything).Return(redis.Nil)
		mockRepo.On("FindAllWithPagination", ctx, 0, 100, "", "desc").Return([]Product{}, int64(0), nil)
		mockCache.On("Set", ctx, "product:list:1:100::desc", mock.Anything, CacheTTLProductList).Return(nil)

		result, err := service.GetAllProductsWithQuery(ctx, query)

		require.NoError(t, err)
		assert.Equal(t, 100, result.Pagination.PageSize)
		mockRepo.AssertExpectations(t)
	})
}

func TestService_GetProductsByIDs(t *testing.T) {
	tests := []struct {
		name           string
		ids            []uint
		mockRepoReturn []Product
		mockRepoError  error
		expectedResult []Product
		expectedError  bool
	}{
		{
			name:           "should return empty slice when no IDs provided",
			ids:            []uint{},
			mockRepoReturn: []Product{},
			mockRepoError:  nil,
			expectedResult: []Product{},
			expectedError:  false,
		},
		{
			name: "should return products from cache and database",
			ids:  []uint{1, 2},
			mockRepoReturn: []Product{
				{ID: 2, Name: "Product 2", Price: 200, Stock: 20},
			},
			mockRepoError: nil,
			expectedResult: []Product{
				{ID: 1, Name: "Product 1", Price: 100, Stock: 10},
				{ID: 2, Name: "Product 2", Price: 200, Stock: 20},
			},
			expectedError: false,
		},
		{
			name:           "should return error when repository fails",
			ids:            []uint{1, 2},
			mockRepoReturn: nil,
			mockRepoError:  errors.New("database error"),
			expectedResult: nil,
			expectedError:  true,
		},
		{
			name:           "should return error when too many IDs requested",
			ids:            make([]uint, 1001), // Create slice with 1001 elements
			mockRepoReturn: nil,
			mockRepoError:  errors.New("too many IDs requested"),
			expectedResult: nil,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockRepository)
			mockCache := new(MockCache)
			logger := zap.NewNop()

			service := NewService(mockRepo, mockCache, logger)

			if len(tt.ids) > 0 && len(tt.ids) <= 1000 {
				mockCache.On("Get", mock.Anything, "product:id:1", mock.AnythingOfType("*product.Product")).Return(nil).Run(func(args mock.Arguments) {
					product := args.Get(2).(*Product)
					*product = Product{ID: 1, Name: "Product 1", Price: 100, Stock: 10}
				})
				mockCache.On("Get", mock.Anything, "product:id:2", mock.AnythingOfType("*product.Product")).Return(redis.Nil)
				mockRepo.On("FindMultipleByIDs", mock.Anything, []uint{2}).Return(tt.mockRepoReturn, tt.mockRepoError)
				if tt.mockRepoError == nil {
					mockCache.On("Set", mock.Anything, "product:id:2", mock.AnythingOfType("product.Product"), mock.AnythingOfType("time.Duration")).Return(nil)
				}
			}

			result, err := service.GetProductsByIDs(context.Background(), tt.ids)

			if tt.expectedError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockRepo.AssertExpectations(t)
			mockCache.AssertExpectations(t)
		})
	}
}
