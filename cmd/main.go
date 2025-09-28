package main

import (
	"mini-e-commerce/internal/config"
	"mini-e-commerce/internal/database"
	"mini-e-commerce/internal/logger"
	"mini-e-commerce/internal/middleware"
	"mini-e-commerce/internal/swagger"
	"mini-e-commerce/routes"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	if err := logger.Init(); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	swagger.SetupSwaggerInfo()

	cfg := config.Load()
	db := database.Connect(cfg.DatabaseUrl)
	if err := database.Migrate(db); err != nil {
		logger.Fatal("Failed to migrate database: ", zap.Error(err))
	}
	rdb := database.ConnectRedis(cfg.RedisAddr, cfg.RedisPassword)

	r := gin.Default()
	r.Use(middleware.RequestLogger())
	r.Use(middleware.ErrorLogger())

	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		logger.Fatal("Failed to set trusted proxies: ", zap.Error(err))
	}

	routes.RegisterRoutes(r, db, rdb)

	port := cfg.Port
	if port == "" {
		port = "8080"
	}

	logger.Info("Starting server", zap.String("port", port))

	go func() {
		if err := r.Run(":" + port); err != nil {
			logger.Fatal("Failed to run server: ", zap.Error(err))
		}
	}()

	<-quit
	logger.Info("Server shutting down gracefully...")

}
