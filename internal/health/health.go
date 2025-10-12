package health

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthChecker struct {
	db          *gorm.DB
	redisClient *redis.Client
}

type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
	Version  string            `json:"version,omitempty"`
}

func NewHealthChecker(db *gorm.DB, redisClient *redis.Client) *HealthChecker {
	return &HealthChecker{
		db:          db,
		redisClient: redisClient,
	}
}

// HealthCheck returns a simple health check
func (h *HealthChecker) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// ReadinessCheck checks if all dependencies are ready
func (h *HealthChecker) ReadinessCheck(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	response := HealthResponse{
		Status:   "healthy",
		Services: make(map[string]string),
		Version:  "1.0.0",
	}

	// Check database
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.PingContext(ctx) != nil {
		response.Services["database"] = "unhealthy"
		response.Status = "unhealthy"
	} else {
		response.Services["database"] = "healthy"
	}

	// Check Redis
	if err := h.redisClient.Ping(ctx).Err(); err != nil {
		response.Services["redis"] = "unhealthy"
		response.Status = "unhealthy"
	} else {
		response.Services["redis"] = "healthy"
	}

	statusCode := http.StatusOK
	if response.Status == "unhealthy" {
		statusCode = http.StatusServiceUnavailable
	}

	c.JSON(statusCode, response)
}

// LivenessCheck checks if the application is alive
func (h *HealthChecker) LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "alive",
	})
}
