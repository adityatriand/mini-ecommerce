package main

import (
	"mini-e-commerce/services/order/internal/handler"
	"mini-e-commerce/services/order/internal/model"
	"mini-e-commerce/services/order/internal/repository"
	"mini-e-commerce/services/order/internal/service"
	"mini-e-commerce/shared/cache"
	"mini-e-commerce/shared/config"
	"mini-e-commerce/shared/database"
	"mini-e-commerce/shared/logger"
	"mini-e-commerce/shared/metrics"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	// Swagger imports
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Initialize logger
	loggerConfig := logger.NewConfig()
	log, err := logger.NewLogger(loggerConfig, "order-service")
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	// Load configuration
	cfg, err := config.LoadConfig("order-service")
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	// Connect to database
	db, err := database.Connect(database.Config{
		URL:             cfg.Database.URL,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}, log.GetZapLogger())
	if err != nil {
		log.Fatal("Failed to connect to database", zap.Error(err))
	}

	// Run migrations
	models := []interface{}{&model.Order{}, &model.OrderItem{}}
	if err := database.Migrate(db, models, log.GetZapLogger()); err != nil {
		log.Fatal("Failed to migrate database", zap.Error(err))
	}

	// Connect to Redis
	redisCache, err := cache.NewRedisCache(cache.Config{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	}, log.GetZapLogger())
	if err != nil {
		log.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	// Initialize components
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, redisCache, log.GetZapLogger(), cfg)
	orderHandler := handler.NewOrderHandler(orderService, log.GetZapLogger())

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Setup router
	r := gin.Default()
	
	// Add middleware
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(metrics.PrometheusMiddleware("order-service"))

	// Health check endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "order-service"})
	})
	r.GET("/health/ready", func(c *gin.Context) {
		// Check database and Redis connectivity
		if err := database.HealthCheck(db); err != nil {
			c.JSON(503, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		if err := redisCache.Ping(c.Request.Context()); err != nil {
			c.JSON(503, gin.H{"status": "unhealthy", "error": err.Error()})
			return
		}
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger UI endpoints
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := r.Group("/api/v1")
	orderHandler.RegisterRoutes(api)

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8003"
	}

	log.Info("Starting order service", zap.String("port", port))

	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	<-quit
	log.Info("Order service shutting down gracefully...")
}
