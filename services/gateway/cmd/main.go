package main

import (
	"mini-e-commerce/services/gateway/internal/middleware"
	"mini-e-commerce/services/gateway/internal/service"
	"mini-e-commerce/shared/config"
	"mini-e-commerce/shared/logger"
	"mini-e-commerce/shared/metrics"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	// Swagger imports
	_ "mini-e-commerce/services/gateway/docs"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func main() {
	// Initialize logger
	loggerConfig := logger.NewConfig()
	log, err := logger.NewLogger(loggerConfig, "gateway-service")
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer log.Sync()

	// Load configuration
	cfg, err := config.LoadConfig("gateway-service")
	if err != nil {
		log.Fatal("Failed to load config", zap.Error(err))
	}

	// Initialize services
	proxyService := service.NewProxyService(cfg, log.GetZapLogger())
	authMiddleware := middleware.NewAuthMiddleware(cfg, log.GetZapLogger())

	// Setup graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Setup router
	r := gin.Default()
	
	// Add middleware
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.RequestLogger(log.GetZapLogger()))
	r.Use(metrics.PrometheusMiddleware("gateway-service"))

	// Health check endpoints
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "gateway-service"})
	})
	r.GET("/health/ready", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Swagger UI endpoints
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := r.Group("/api/v1")
	{
		// Public routes (no authentication required)
		api.POST("/users/register", proxyService.ProxyToUserService)
		api.POST("/users/login", proxyService.ProxyToUserService)
		
		// Protected routes (authentication required)
		protected := api.Group("")
		protected.Use(authMiddleware.RequireAuth())
		{
			// User service routes
			protected.GET("/users/:id", proxyService.ProxyToUserService)
			protected.PUT("/users/:id", proxyService.ProxyToUserService)
			protected.DELETE("/users/:id", proxyService.ProxyToUserService)
			protected.GET("/users", proxyService.ProxyToUserService)
			protected.POST("/users/refresh", proxyService.ProxyToUserService)
			protected.POST("/users/logout", proxyService.ProxyToUserService)
			
			// Product service routes
			protected.GET("/products", proxyService.ProxyToProductService)
			protected.GET("/products/:id", proxyService.ProxyToProductService)
			protected.POST("/products", proxyService.ProxyToProductService)
			protected.PUT("/products/:id", proxyService.ProxyToProductService)
			protected.DELETE("/products/:id", proxyService.ProxyToProductService)
			
			// Order service routes
			protected.GET("/orders", proxyService.ProxyToOrderService)
			protected.GET("/orders/:id", proxyService.ProxyToOrderService)
			protected.POST("/orders", proxyService.ProxyToOrderService)
			protected.PUT("/orders/:id", proxyService.ProxyToOrderService)
			protected.DELETE("/orders/:id", proxyService.ProxyToOrderService)
		}
	}

	// Start server
	port := cfg.Port
	if port == "" {
		port = "8000"
	}

	log.Info("Starting gateway service", zap.String("port", port))

	go func() {
		if err := r.Run(":" + port); err != nil {
			log.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	<-quit
	log.Info("Gateway service shutting down gracefully...")
}
