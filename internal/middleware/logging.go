package middleware

import (
	"mini-e-commerce/internal/logger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func RequestLogger() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := uuid.New().String()
		ctx.Set("request_id", requestID)

		start := time.Now()
		ctx.Next()
		duration := time.Since(start)

		logger.Info("HTTP Request",
			zap.String("request_id", requestID),
			zap.String("method", ctx.Request.Method),
			zap.String("path", ctx.Request.URL.Path),
			zap.String("query", ctx.Request.URL.RawQuery),
			zap.Int("status", ctx.Writer.Status()),
			zap.Duration("duration", duration),
			zap.String("client_ip", ctx.ClientIP()),
			zap.String("user_agent", ctx.Request.UserAgent()),
			zap.Int("response_size", ctx.Writer.Size()),
		)
	}
}

func ErrorLogger() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		requestID, exists := c.Get("request_id")
		var requestIDStr string
		if exists {
			if id, ok := requestID.(string); ok {
				requestIDStr = id
			} else {
				requestIDStr = "unknown"
			}
		} else {
			requestIDStr = "missing"
		}

		logger.Error("Panic recovered",
			zap.Any("panic", recovered),
			zap.String("request_id", requestIDStr),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
		)

		c.JSON(500, gin.H{
			"success": false,
			"message": "Internal server error",
			"error": gin.H{
				"code":    "INTERNAL_SERVER_ERROR",
				"details": "An unexpected error occurred",
			},
		})

	})
}
