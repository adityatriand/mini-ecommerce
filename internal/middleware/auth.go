package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func AuthMiddleware(rdb *redis.Client, ctx context.Context) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID, err := c.Cookie("session_id")
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no session"})
			c.Abort()
			return
		}

		userID, err := rdb.Get(ctx, "session:"+sessionID).Result()
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
			c.Abort()
			return
		}

		c.Set("user_id", userID)
		c.Next()
	}
}
