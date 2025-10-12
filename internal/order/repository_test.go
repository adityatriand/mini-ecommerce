package order

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

	t.Run("should create order successfully", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		order := &Order{
			UserID:     1,
			TotalPrice: 50000,
			Status:     StatusPending,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "orders"`)).
			WithArgs(order.UserID, order.TotalPrice, order.Status, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.Create(ctx, order)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), order.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when creation fails", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		order := &Order{
			UserID:     1,
			TotalPrice: 50000,
			Status:     StatusPending,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "orders"`)).
			WillReturnError(gorm.ErrInvalidData)
		mock.ExpectRollback()

		err := repo.Create(ctx, order)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_FindAll(t *testing.T) {
	ctx := context.Background()

	t.Run("should find all orders successfully", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		orderRows := sqlmock.NewRows([]string{"id", "user_id", "total_price", "status", "created_at", "updated_at"}).
			AddRow(1, 1, 50000, "PENDING", time.Now(), time.Now()).
			AddRow(2, 2, 75000, "PAID", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders"`)).
			WillReturnRows(orderRows)

		itemRows := sqlmock.NewRows([]string{"id", "order_id", "product_id", "quantity", "price", "subtotal", "created_at", "updated_at"})
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "order_items" WHERE "order_items"."order_id"`)).
			WillReturnRows(itemRows)

		orders, err := repo.FindAll(ctx)

		assert.NoError(t, err)
		assert.Len(t, orders, 2)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return empty slice when no orders found", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		orderRows := sqlmock.NewRows([]string{"id", "user_id", "total_price", "status", "created_at", "updated_at"})

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders"`)).
			WillReturnRows(orderRows)

		orders, err := repo.FindAll(ctx)

		assert.NoError(t, err)
		assert.Len(t, orders, 0)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_FindByID(t *testing.T) {
	ctx := context.Background()

	t.Run("should find order by ID successfully", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		orderRows := sqlmock.NewRows([]string{"id", "user_id", "total_price", "status", "created_at", "updated_at"}).
			AddRow(1, 1, 50000, "PENDING", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders" WHERE "orders"."id" = $1`)).
			WithArgs(1, 1).
			WillReturnRows(orderRows)

		itemRows := sqlmock.NewRows([]string{"id", "order_id", "product_id", "quantity", "price", "subtotal", "created_at", "updated_at"})
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "order_items" WHERE "order_items"."order_id"`)).
			WillReturnRows(itemRows)

		order, err := repo.FindByID(ctx, 1)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), order.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when order not found", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders" WHERE "orders"."id" = $1`)).
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

	t.Run("should update order successfully", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		order := &Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 50000,
			Status:     StatusPaid,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "orders"`)).
			WithArgs(order.UserID, order.TotalPrice, order.Status, sqlmock.AnyArg(), sqlmock.AnyArg(), order.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, order, func(o *Order) {
			o.Status = StatusPaid
		})

		assert.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when update fails", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		order := &Order{
			ID:         1,
			UserID:     1,
			TotalPrice: 50000,
			Status:     StatusPaid,
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "orders"`)).
			WillReturnError(gorm.ErrInvalidData)
		mock.ExpectRollback()

		err := repo.Update(ctx, order, func(o *Order) {
			o.Status = StatusPaid
		})

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete order successfully", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "orders" WHERE "orders"."id" = $1`)).
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
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "orders" WHERE "orders"."id" = $1`)).
			WillReturnError(gorm.ErrInvalidData)
		mock.ExpectRollback()

		err := repo.Delete(ctx, 1)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_FindAllWithPagination(t *testing.T) {
	ctx := context.Background()

	t.Run("should find orders with pagination successfully", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "orders"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

		orderRows := sqlmock.NewRows([]string{"id", "user_id", "total_price", "status", "created_at", "updated_at"}).
			AddRow(1, 1, 50000, "PENDING", time.Now(), time.Now()).
			AddRow(2, 2, 75000, "PAID", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders" ORDER BY user_id asc LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(orderRows)

		itemRows := sqlmock.NewRows([]string{"id", "order_id", "product_id", "quantity", "price", "subtotal", "created_at", "updated_at"})
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "order_items" WHERE "order_items"."order_id"`)).
			WillReturnRows(itemRows)

		orders, total, err := repo.FindAllWithPagination(ctx, 0, 10, "user_id", "asc")

		assert.NoError(t, err)
		assert.Len(t, orders, 2)
		assert.Equal(t, int64(2), total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should use default sorting when sortBy is empty", func(t *testing.T) {
		gormDB, mock := setupMockDB(t)
		repo := NewRepository(gormDB)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "orders"`)).
			WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

		orderRows := sqlmock.NewRows([]string{"id", "user_id", "total_price", "status", "created_at", "updated_at"}).
			AddRow(1, 1, 50000, "PENDING", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "orders" ORDER BY created_at desc LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(orderRows)

		itemRows := sqlmock.NewRows([]string{"id", "order_id", "product_id", "quantity", "price", "subtotal", "created_at", "updated_at"})
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "order_items" WHERE "order_items"."order_id"`)).
			WillReturnRows(itemRows)

		orders, total, err := repo.FindAllWithPagination(ctx, 0, 10, "", "")

		assert.NoError(t, err)
		assert.Len(t, orders, 1)
		assert.Equal(t, int64(1), total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}
