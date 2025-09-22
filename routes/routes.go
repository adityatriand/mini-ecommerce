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

	auth.RegisterRoutes(api, db, rdb)
	product.RegisterRoutes(api, db, rdb)
	order.RegisterRoutes(api, db, rdb)
}
