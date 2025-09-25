package main

import (
	"log"
	"mini-e-commerce/internal/config"
	"mini-e-commerce/internal/database"
	"mini-e-commerce/routes"

	"github.com/gin-gonic/gin"
)

func main() {

	cfg := config.Load()
	db := database.Connect(cfg.DatabaseUrl)
	if err := database.Migrate(db); err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}
	rdb := database.ConnectRedis(cfg.RedisAddr, cfg.RedisPassword)

	r := gin.Default()

	// Set trusted proxies for security
	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		log.Fatal("Failed to set trusted proxies: ", err)
	}

	routes.RegisterRoutes(r, db, rdb)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to run server: ", err)
	}

}
