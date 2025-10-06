package routes

import (
	"mini-e-commerce/internal/auth"
	"mini-e-commerce/internal/config"
	"mini-e-commerce/internal/logger"
	"mini-e-commerce/internal/order"
	"mini-e-commerce/internal/product"

	_ "mini-e-commerce/docs" // generated docs

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.Engine, db *gorm.DB, rdb *redis.Client, log logger.Logger, jwtManager *auth.JWTManager, sessionManager *auth.SessionManager, cfg *config.Config) {
	api := r.Group("/api")

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	authRepo := auth.NewRepository(db)
	authService := auth.NewService(authRepo, jwtManager, sessionManager, log.GetZapLogger(), cfg.JWTExpiration, cfg.RefreshExpiration)
	authHandler := auth.NewHandler(authService, log)
	authHandler.RegisterRoutes(api)

	productRepo := product.NewRepository(db)
	productService := product.NewService(productRepo)
	productHandler := product.NewHandler(productService, log)
	productHandler.RegisterRoutes(api, jwtManager, sessionManager, log.GetZapLogger())

	orderRepo := order.NewRepository(db)
	orderService := order.NewService(orderRepo, productService)
	orderHandler := order.NewHandler(orderService, log)
	orderHandler.RegisterRoutes(api, jwtManager, sessionManager, log.GetZapLogger())

}
