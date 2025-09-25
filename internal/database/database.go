package database

import (
	"context"
	"log"
	"mini-e-commerce/internal/auth"
	"mini-e-commerce/internal/order"
	"mini-e-commerce/internal/product"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect database: ", err)
	}
	return db
}

func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(&auth.User{}, &product.Product{}, &order.Order{})
}

func ConnectRedis(addr, password string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	if _, err := rdb.Ping(context.Background()).Result(); err != nil {
		log.Fatal("Failed to connect redis: ", err)
	}

	return rdb
}
