package tests

import (
	"context"
	"regexp"
	"testing"
	"time"

	"mini-e-commerce/services/product/internal/model"
	"mini-e-commerce/services/product/internal/repository"

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

func TestNewProductRepository(t *testing.T) {
	t.Run("should create repository successfully", func(t *testing.T) {
		gormDB, _ := setupMockDB(t)
		repo := repository.NewProductRepository(gormDB)
		assert.NotNil(t, repo)
	})
}

func TestRepository_Create(t *testing.T) {
	ctx := context.Background()

	t.Run("should create product successfully", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewProductRepository(gormDB)

		product := &model.Product{
			Name:  "Test model.Product",
			Price: 10000,
			Stock: 50,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "products" ("name","description","price","stock","category","sku","is_active","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING "id"`)).
			WithArgs(product.Name, product.Description, product.Price, product.Stock, product.Category, product.SKU, true, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.Create(ctx, product)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), product.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when creation fails", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewProductRepository(gormDB)

		product := &model.Product{
			Name:  "Test model.Product",
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
		repo := repository.NewProductRepository(gormDB)

		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
			WillReturnRows(countRows)

		rows := sqlmock.NewRows([]string{"id", "name", "description", "price", "stock", "category", "sku", "is_active", "created_at", "updated_at"}).
			AddRow(1, "model.Product 1", "", 10000, 50, "", "", true, time.Now(), time.Now()).
			AddRow(2, "model.Product 2", "", 20000, 30, "", "", true, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" ORDER BY created_at DESC LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(rows)

		query := model.ProductQuery{Page: 1, PageSize: 10}
		products, total, err := repo.FindAll(ctx, query)

		assert.NoError(t, err)
		assert.Len(t, products, 2)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, "model.Product 1", products[0].Name)
		assert.Equal(t, "model.Product 2", products[1].Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return empty slice when no products found", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewProductRepository(gormDB)

		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
			WillReturnRows(countRows)

		rows := sqlmock.NewRows([]string{"id", "name", "description", "price", "stock", "category", "sku", "is_active", "created_at", "updated_at"})

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" ORDER BY created_at DESC LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(rows)

		query := model.ProductQuery{Page: 1, PageSize: 10}
		products, total, err := repo.FindAll(ctx, query)

		assert.NoError(t, err)
		assert.Len(t, products, 0)
		assert.Equal(t, int64(0), total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_FindByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should find product by ID successfully", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewProductRepository(gormDB)

		rows := sqlmock.NewRows([]string{"id", "name", "price", "stock", "created_at", "updated_at"}).
			AddRow(1, "Test model.Product", 10000, 50, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" WHERE "products"."id" = $1`)).
			WithArgs(1, 1).
			WillReturnRows(rows)

		product, err := repo.FindByID(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), product.ID)
		assert.Equal(t, "Test model.Product", product.Name)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when product not found", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewProductRepository(gormDB)

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
		repo := repository.NewProductRepository(gormDB)

		product := &model.Product{
			ID:    1,
			Name:  "Updated model.Product",
			Price: 15000,
			Stock: 100,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "products" SET "name"=$1,"description"=$2,"price"=$3,"stock"=$4,"category"=$5,"sku"=$6,"is_active"=$7,"created_at"=$8,"updated_at"=$9 WHERE "id" = $10`)).
			WithArgs(product.Name, product.Description, product.Price, product.Stock, product.Category, product.SKU, product.IsActive, product.CreatedAt, sqlmock.AnyArg(), product.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, product)

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when update fails", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewProductRepository(gormDB)

		product := &model.Product{
			ID:    1,
			Name:  "Updated model.Product",
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
		repo := repository.NewProductRepository(gormDB)

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
		repo := repository.NewProductRepository(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "products" WHERE "products"."id" = $1`)).
			WillReturnError(gorm.ErrInvalidData)
		mock.ExpectRollback()

		err := repo.Delete(ctx, 1)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_FindAllWithQuery(t *testing.T) {
	ctx := context.Background()

	t.Run("should find products with pagination successfully", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewProductRepository(gormDB)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		rows := sqlmock.NewRows([]string{"id", "name", "price", "stock", "created_at", "updated_at"}).
			AddRow(1, "model.Product 1", 10000, 50, time.Now(), time.Now()).
			AddRow(2, "model.Product 2", 20000, 30, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" ORDER BY name ASC LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(rows)

		query := model.ProductQuery{Page: 1, PageSize: 10, SortBy: "name", Order: "asc"}
		products, total, err := repo.FindAll(ctx, query)

		assert.NoError(t, err)
		assert.Len(t, products, 2)
		assert.Equal(t, int64(2), total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should use default sorting when sortBy is empty", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := repository.NewProductRepository(gormDB)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "products"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		rows := sqlmock.NewRows([]string{"id", "name", "price", "stock", "created_at", "updated_at"}).
			AddRow(1, "model.Product 1", 10000, 50, time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "products" ORDER BY created_at DESC LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(rows)

		query := model.ProductQuery{Page: 1, PageSize: 10}
		products, total, err := repo.FindAll(ctx, query)

		assert.NoError(t, err)
		assert.Len(t, products, 1)
		assert.Equal(t, int64(1), total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}