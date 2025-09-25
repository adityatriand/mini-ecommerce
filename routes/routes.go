package routes

import (
	"mini-e-commerce/internal/auth"
	"mini-e-commerce/internal/order"
	"mini-e-commerce/internal/product"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client) {
	api := r.Group("/api")

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, rdb)
	authHandler := auth.NewHandler(authService)
	authHandler.RegisterRoutes(api)

	productRepo := product.NewRepository(db)
	productService := product.NewService(productRepo)
	productHandler := product.NewHandler(productService)
	productHandler.RegisterRoutes(api, rdb)

	orderRepo := order.NewRepository(db)
	orderService := order.NewService(orderRepo, productService)
	orderHandler := order.NewHandler(orderService)
	orderHandler.RegisterRoutes(api, rdb)

}
