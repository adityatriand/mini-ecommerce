package database

import (
	"context"
	"mini-e-commerce/internal/auth"
	"mini-e-commerce/internal/logger"
	"mini-e-commerce/internal/order"
	"mini-e-commerce/internal/product"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(dsn string, log logger.Logger) *gorm.DB {
	log.Info("Connecting to database...")

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database: ", zap.Error(err))
	}

	log.Info("Database connection established successfully")
	return db
}

func Migrate(db *gorm.DB, log logger.Logger) error {
	log.Info("Starting database migration...")

	if err := db.AutoMigrate(&auth.User{}, &product.Product{}, &order.Order{}); err != nil {
		log.Error("Database migration failed", zap.Error(err))
		return err
	}

	log.Info("Database migration completed successfully")
	return nil
}

func ConnectRedis(addr, password string, log logger.Logger) *redis.Client {
	log.Info("Connecting to Redis...", zap.String("addr", addr))

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatal("Failed to connect redis: ", zap.Error(err), zap.String("addr", addr))
	}

	log.Info("Redis connection established successfully")
	return rdb
}
