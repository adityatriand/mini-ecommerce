package middleware

import (
	"fmt"
	"strings"
	"time"

	"mini-e-commerce/shared/config"
	"mini-e-commerce/shared/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthMiddleware handles authentication for the gateway
type AuthMiddleware struct {
	config *config.ServiceConfig
	logger *zap.Logger
}

func NewAuthMiddleware(config *config.ServiceConfig, logger *zap.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		config: config,
		logger: logger,
	}
}

// RequireAuth returns a middleware that requires authentication
func (a *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.NewResponseHelper(a.logger).Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			response.NewResponseHelper(a.logger).Unauthorized(c, "Invalid authorization format")
			c.Abort()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Validate token
		token, err := a.validateToken(tokenString)
		if err != nil {
			a.logger.Warn("Invalid token", zap.Error(err))
			response.NewResponseHelper(a.logger).Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.NewResponseHelper(a.logger).Unauthorized(c, "Invalid token claims")
			c.Abort()
			return
		}

		// Get user ID from claims
		userID, ok := claims["user_id"].(float64)
		if !ok {
			response.NewResponseHelper(a.logger).Unauthorized(c, "Invalid user ID in token")
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("user_id", uint(userID))
		c.Set("user_email", claims["email"])

		a.logger.Debug("User authenticated", 
			zap.Uint("user_id", uint(userID)),
			zap.String("email", claims["email"].(string)))

		c.Next()
	}
}

// validateToken validates a JWT token
func (a *AuthMiddleware) validateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Check signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// Return secret key
		return []byte(a.config.JWT.Secret), nil
	})
}

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestLogger logs HTTP requests
func RequestLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		if raw != "" {
			path = path + "?" + raw
		}

		logger.Info("HTTP Request",
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.String("client_ip", clientIP),
			zap.Int("body_size", bodySize),
		)
	}
}
