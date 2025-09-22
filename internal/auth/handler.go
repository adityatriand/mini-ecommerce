package auth

import (
	"errors"
	"net/http"

	"mini-e-commerce/internal/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

func RegisterRoutes(r *gin.RouterGroup, db *gorm.DB, rdb *redis.Client) {
	repo := NewRepository(db)
	service := NewService(repo, rdb)

	group := r.Group("/auth")
	{
		group.POST("/register", func(c *gin.Context) {
			var input RegisterInput

			if err := c.ShouldBindJSON(&input); err != nil {
				response.Error(c, http.StatusBadRequest, "Failed to register user", response.ErrCodeValidationError, err.Error())
				return
			}

			user, err := service.RegisterUser(c.Request.Context(), input)
			if err != nil {
				response.Error(c, http.StatusInternalServerError, "Failed to register user", response.ErrCodeInternalServer, err.Error())
				return
			}

			response.Success(c, "User registered", gin.H{
				"user_id": user.ID,
			})
		})
		group.POST("/login", func(c *gin.Context) {
			var input LoginInput

			if err := c.ShouldBindJSON(&input); err != nil {
				response.Error(c, http.StatusBadRequest, "Failed to login user", response.ErrCodeValidationError, err.Error())
				return
			}

			user, sessionID, err := service.LoginUser(c.Request.Context(), input)
			if err != nil {
				if errors.Is(err, ErrInvalidCredentials) {
					response.Error(c, http.StatusUnauthorized, "Invalid credentials", response.ErrCodeInvalidCredentials, "wrong email or password")
					return
				}
				if errors.Is(err, gorm.ErrRecordNotFound) {
					response.Error(c, http.StatusUnauthorized, "Invalid credentials", response.ErrCodeInvalidCredentials, "user not found")
					return
				}
				response.Error(c, http.StatusInternalServerError, "Failed to login", response.ErrCodeInternalServer, err.Error())
				return
			}

			c.SetCookie("session_id", sessionID, 3600, "/", "localhost", false, true)

			response.Success(c, "Login successful", gin.H{
				"user_id":    user.ID,
				"session_id": sessionID,
			})
		})
	}
}
