package main

import (
	"mini-e-commerce/internal/auth"
	"mini-e-commerce/internal/cache"
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
	configLog := logger.NewConfig()
	logger, err := logger.NewLogger(configLog)
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	swagger.SetupSwaggerInfo()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config: ", zap.Error(err))
	}
	db := database.Connect(cfg.DatabaseUrl, logger)
	if err := database.Migrate(db, logger); err != nil {
		logger.Fatal("Failed to migrate database: ", zap.Error(err))
	}
	rdb := database.ConnectRedis(cfg.RedisAddr, cfg.RedisPassword, logger)

	redisCache := cache.NewRedisCache(rdb, logger.GetZapLogger())

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpiration, logger.GetZapLogger())
	sessionManager := auth.NewSessionManager(rdb, logger.GetZapLogger())

	logger.Info("Hybrid auth system initialized",
		zap.Duration("jwt_expiration", cfg.JWTExpiration),
		zap.Duration("refresh_expiration", cfg.RefreshExpiration),
	)

	r := gin.Default()
	r.Use(middleware.RequestLogger(logger))
	r.Use(middleware.ErrorLogger(logger))

	if err := r.SetTrustedProxies(cfg.TrustedProxies); err != nil {
		logger.Fatal("Failed to set trusted proxies: ", zap.Error(err))
	}

	routes.RegisterRoutes(r, db, redisCache, logger, jwtManager, sessionManager, &cfg)

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
