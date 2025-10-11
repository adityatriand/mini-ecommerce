package middleware

import (
	"errors"
	"mini-e-commerce/internal/auth"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AuthMiddleware(jwtManager auth.JWTManagerInterface, sessionManager auth.SessionManagerInterface, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		authHeader := c.GetHeader("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwtManager.Verify(token)
			if err == nil {
				c.Set("user_id", claims.UserID)
				logger.Debug("User authenticated via JWT", zap.Uint("user_id", claims.UserID))
				c.Next()
				return
			}

			if errors.Is(err, auth.ErrExpiredToken) {
				logger.Debug("JWT token expired", zap.String("token", token[:10]+"..."))
			} else {
				logger.Warn("Invalid JWT token", zap.Error(err))
			}
		}

		sessionID, err := c.Cookie("session_id")
		if err != nil {
			logger.Debug("No session cookie found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			c.Abort()
			return
		}

		userIDStr, err := c.Cookie("user_id")
		if err != nil {
			logger.Debug("No user_id cookie found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			c.Abort()
			return
		}

		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			logger.Warn("Invalid user_id cookie format", zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			c.Abort()
			return
		}

		refreshToken, err := c.Cookie("refresh_token")
		if err != nil {
			logger.Debug("No refresh_token cookie found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			c.Abort()
			return
		}

		if err := sessionManager.ValidateRefreshToken(ctx, uint(userID), sessionID, refreshToken); err != nil {
			if errors.Is(err, auth.ErrSessionNotFound) {
				logger.Debug("Session not found", zap.Uint("user_id", uint(userID)))
			} else if errors.Is(err, auth.ErrInvalidRefreshToken) {
				logger.Warn("Invalid refresh token", zap.Uint("user_id", uint(userID)))
			} else {
				logger.Error("Session validation error", zap.Error(err), zap.Uint("user_id", uint(userID)))
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			c.Abort()
			return
		}

		c.Set("user_id", uint(userID))
		logger.Debug("User authenticated via session", zap.Uint("user_id", uint(userID)))
		c.Next()
	}
}
