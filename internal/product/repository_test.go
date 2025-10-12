package product

import (
	"context"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

func TestNewRepository(t *testing.T) {
	t.Run("should create repository successfully", func(t *testing.T) {
		gormDB, _ := setupMockDB(t)
		repo := NewRepository(gormDB)
		assert.NotNil(t, repo)
	})
}

func TestRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("should create product successfully", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		product := &Product{
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "products"`)).
			WithArgs(product.Name, product.Price, product.Stock, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.Create(ctx, product)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), product.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when creation fails", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		product := &Product{
			Name:  "Test Product",
			Price: 10000,
			Stock: 50,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "products"`)).
			WillReturnError(gorm.ErrInvalidData)
		mock.ExpectRollback()

		err := repo.Create(ctx, product)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_FindAll(t *testing.T) {
	ctx := context.Background()

	t.Run("should find all products successfully", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		rows := sqlmock.NewRows([]string{"id", "name", "price", "stock", "created_at", "updated_at"}).
			AddRow(1, "Product 1", 10000, 50, time.Now(), time.Now()).
			AddRow(2, "Product 2", 20000, 30, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products"`)).
			WillReturnRows(rows)

		products, err := repo.FindAll(ctx)

		assert.NoError(t, err)
		assert.Len(t, products, 2)
		assert.Equal(t, "Product 1", products[0].Name)
		assert.Equal(t, "Product 2", products[1].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return empty slice when no products found", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		rows := sqlmock.NewRows([]string{"id", "name", "price", "stock", "created_at", "updated_at"})

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products"`)).
			WillReturnRows(rows)

		products, err := repo.FindAll(ctx)

		assert.NoError(t, err)
		assert.Len(t, products, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_FindByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should find product by ID successfully", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		rows := sqlmock.NewRows([]string{"id", "name", "price", "stock", "created_at", "updated_at"}).
			AddRow(1, "Test Product", 10000, 50, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE "products"."id" = $1`)).
			WithArgs(1, 1).
			WillReturnRows(rows)

		product, err := repo.FindByID(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), product.ID)
		assert.Equal(t, "Test Product", product.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when product not found", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE "products"."id" = $1`)).
			WithArgs(999, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		_, err := repo.FindByID(ctx, 999)

		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_Update(t *testing.T) {
	ctx := context.Background()

	t.Run("should update product successfully", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		product := &Product{
			ID:    1,
			Name:  "Updated Product",
			Price: 15000,
			Stock: 100,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "products"`)).
			WithArgs(product.Name, product.Price, product.Stock, sqlmock.AnyArg(), sqlmock.AnyArg(), product.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, product)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when update fails", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		product := &Product{
			ID:    1,
			Name:  "Updated Product",
			Price: 15000,
			Stock: 100,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "products"`)).
			WillReturnError(gorm.ErrInvalidData)
		mock.ExpectRollback()

		err := repo.Update(ctx, product)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete product successfully", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "products" WHERE "products"."id" = $1`)).
			WithArgs(1).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectCommit()

		err := repo.Delete(ctx, 1)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when delete fails", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "products" WHERE "products"."id" = $1`)).
			WillReturnError(gorm.ErrInvalidData)
		mock.ExpectRollback()

		err := repo.Delete(ctx, 1)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_FindAllWithPagination(t *testing.T) {
	ctx := context.Background()

	t.Run("should find products with pagination successfully", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		rows := sqlmock.NewRows([]string{"id", "name", "price", "stock", "created_at", "updated_at"}).
			AddRow(1, "Product 1", 10000, 50, time.Now(), time.Now()).
			AddRow(2, "Product 2", 20000, 30, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" ORDER BY name asc LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(rows)

		products, total, err := repo.FindAllWithPagination(ctx, 0, 10, "name", "asc")

		assert.NoError(t, err)
		assert.Len(t, products, 2)
		assert.Equal(t, int64(2), total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should use default sorting when sortBy is empty", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "name", "price", "stock", "created_at", "updated_at"}).
			AddRow(1, "Product 1", 10000, 50, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" ORDER BY created_at desc LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(rows)

		products, total, err := repo.FindAllWithPagination(ctx, 0, 10, "", "")

		assert.NoError(t, err)
		assert.Len(t, products, 1)
		assert.Equal(t, int64(1), total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
