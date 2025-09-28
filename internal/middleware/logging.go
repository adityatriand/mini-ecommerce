package middleware

import (
	"mini-e-commerce/internal/logger"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func RequestLogger(log logger.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		requestID := uuid.New().String()
		ctx.Set("request_id", requestID)

		start := time.Now()
		ctx.Next()
		duration := time.Since(start)

		log.Info("HTTP Request",
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

func ErrorLogger(log logger.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		requestID := extractRequestIDSafely(c)

		log.Error("Panic recovered",
			zap.Any("panic", recovered),
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Stack("stacktrace"),
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

func extractRequestIDSafely(c *gin.Context) string {
	if c == nil {
		return "missing_context"
	}

	requestID, exists := c.Get("request_id")
	if !exists {
		return "missing_request_id"
	}

	if id, ok := requestID.(string); ok {
		if id == "" {
			return "empty_request_id"
		}
		return id
	}

	return "invalid_request_id_type"
}
