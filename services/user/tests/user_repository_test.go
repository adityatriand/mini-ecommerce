package tests

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"mini-e-commerce/services/user/internal/model"
	"mini-e-commerce/services/user/internal/repository"
)

func setupTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: db,
	}), &gorm.Config{})
	require.NoError(t, err)

	return gormDB, mock
}

func TestNewUserRepository(t *testing.T) {
	t.Run("should create repository successfully", func(t *testing.T) {
		db, _ := setupTestDB(t)

		repo := repository.NewUserRepository(db)

		assert.NotNil(t, repo)
	})
}

func TestRepository_Create(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	t.Run("should create user successfully", func(t *testing.T) {
		user := &model.User{
			Email:    "test@example.com",
			Password: "hashed-password",
			IsActive: true,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users" ("email","password","first_name","last_name","is_active","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
			WithArgs(user.Email, user.Password, user.FirstName, user.LastName, user.IsActive, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
		mock.ExpectCommit()

		err := repo.Create(ctx, user)

		require.NoError(t, err)
		assert.Equal(t, uint(1), user.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when creation fails", func(t *testing.T) {
		user := &model.User{
			Email:    "test@example.com",
			Password: "hashed-password",
			IsActive: true,
		}

		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users" ("email","password","first_name","last_name","is_active","created_at","updated_at") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
			WithArgs(user.Email, user.Password, user.FirstName, user.LastName, user.IsActive, sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		err := repo.Create(ctx, user)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_FindByEmail(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	t.Run("should find user by email successfully", func(t *testing.T) {
		email := "test@example.com"
		expectedUser := model.User{
			ID:        1,
			Email:     email,
			Password:  "hashed-password",
			CreatedAt: time.Now(),
		}

		rows := sqlmock.NewRows([]string{"id", "email", "password", "created_at"}).
			AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Password, expectedUser.CreatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1`)).
			WithArgs(email, 1).
			WillReturnRows(rows)

		user, err := repo.FindByEmail(ctx, email)

		require.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		email := "notfound@example.com"

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1`)).
			WithArgs(email, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.FindByEmail(ctx, email)

		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		assert.Equal(t, uint(0), user.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_FindByID(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	t.Run("should find user by ID successfully", func(t *testing.T) {
		userID := uint(1)
		expectedUser := model.User{
			ID:        userID,
			Email:     "test@example.com",
			Password:  "hashed-password",
			CreatedAt: time.Now(),
		}

		rows := sqlmock.NewRows([]string{"id", "email", "password", "created_at"}).
			AddRow(expectedUser.ID, expectedUser.Email, expectedUser.Password, expectedUser.CreatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
			WithArgs(userID, 1).
			WillReturnRows(rows)

		user, err := repo.FindByID(ctx, userID)

		require.NoError(t, err)
		assert.Equal(t, expectedUser.ID, user.ID)
		assert.Equal(t, expectedUser.Email, user.Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when user not found", func(t *testing.T) {
		userID := uint(999)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1`)).
			WithArgs(userID, 1).
			WillReturnError(gorm.ErrRecordNotFound)

		user, err := repo.FindByID(ctx, userID)

		assert.Error(t, err)
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		assert.Equal(t, uint(0), user.ID)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_Update(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	t.Run("should update user successfully", func(t *testing.T) {
		user := &model.User{
			ID:        1,
			Email:     "updated@example.com",
			Password:  "new-hashed-password",
			CreatedAt: time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET "email"=$1,"password"=$2,"first_name"=$3,"last_name"=$4,"is_active"=$5,"created_at"=$6,"updated_at"=$7 WHERE "id" = $8`)).
			WithArgs(user.Email, user.Password, user.FirstName, user.LastName, user.IsActive, user.CreatedAt, sqlmock.AnyArg(), user.ID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Update(ctx, user)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when update fails", func(t *testing.T) {
		user := &model.User{
			ID:        1,
			Email:     "updated@example.com",
			Password:  "new-hashed-password",
			CreatedAt: time.Now(),
		}

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "users" SET "email"=$1,"password"=$2,"first_name"=$3,"last_name"=$4,"is_active"=$5,"created_at"=$6,"updated_at"=$7 WHERE "id" = $8`)).
			WithArgs(user.Email, user.Password, user.FirstName, user.LastName, user.IsActive, user.CreatedAt, sqlmock.AnyArg(), user.ID).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		err := repo.Update(ctx, user)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_Delete(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	t.Run("should delete user successfully", func(t *testing.T) {
		userID := uint(1)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "users" WHERE "users"."id" = $1`)).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.Delete(ctx, userID)

		require.NoError(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when delete fails", func(t *testing.T) {
		userID := uint(1)

		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`DELETE FROM "users"`)).
			WithArgs(userID).
			WillReturnError(errors.New("database error"))
		mock.ExpectRollback()

		err := repo.Delete(ctx, userID)

		assert.Error(t, err)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestRepository_FindAll(t *testing.T) {
	db, mock := setupTestDB(t)
	repo := repository.NewUserRepository(db)
	ctx := context.Background()

	t.Run("should find all users successfully", func(t *testing.T) {
		expectedUsers := []model.User{
			{
				ID:        1,
				Email:     "user1@example.com",
				Password:  "hashed-password-1",
				CreatedAt: time.Now(),
			},
			{
				ID:        2,
				Email:     "user2@example.com",
				Password:  "hashed-password-2",
				CreatedAt: time.Now(),
			},
		}

		countRows := sqlmock.NewRows([]string{"count"}).AddRow(2)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "users"`)).
			WillReturnRows(countRows)

		rows := sqlmock.NewRows([]string{"id", "email", "password", "first_name", "last_name", "is_active", "created_at", "updated_at"}).
			AddRow(expectedUsers[0].ID, expectedUsers[0].Email, expectedUsers[0].Password, expectedUsers[0].FirstName, expectedUsers[0].LastName, expectedUsers[0].IsActive, expectedUsers[0].CreatedAt, expectedUsers[0].UpdatedAt).
			AddRow(expectedUsers[1].ID, expectedUsers[1].Email, expectedUsers[1].Password, expectedUsers[1].FirstName, expectedUsers[1].LastName, expectedUsers[1].IsActive, expectedUsers[1].CreatedAt, expectedUsers[1].UpdatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" ORDER BY created_at DESC LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(rows)

		users, total, err := repo.FindAll(ctx, 10, 0)

		require.NoError(t, err)
		assert.Len(t, users, 2)
		assert.Equal(t, int64(2), total)
		assert.Equal(t, expectedUsers[0].Email, users[0].Email)
		assert.Equal(t, expectedUsers[1].Email, users[1].Email)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return empty slice when no users found", func(t *testing.T) {
		countRows := sqlmock.NewRows([]string{"count"}).AddRow(0)
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "users"`)).
			WillReturnRows(countRows)

		rows := sqlmock.NewRows([]string{"id", "email", "password", "first_name", "last_name", "is_active", "created_at", "updated_at"})

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" ORDER BY created_at DESC LIMIT $1`)).
			WithArgs(10).
			WillReturnRows(rows)

		users, total, err := repo.FindAll(ctx, 10, 0)

		require.NoError(t, err)
		assert.Empty(t, users)
		assert.Equal(t, int64(0), total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("should return error when query fails", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(`SELECT count(*) FROM "users"`)).
			WillReturnError(errors.New("database error"))

		users, total, err := repo.FindAll(ctx, 10, 0)

		assert.Error(t, err)
		assert.Nil(t, users)
		assert.Equal(t, int64(0), total)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}