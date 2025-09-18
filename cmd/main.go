package main

import (
	"context"
	"fmt"
	"log"
	"mini-e-commerce/internal/auth"
	"mini-e-commerce/internal/middleware"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	rdb *redis.Client
	ctx = context.Background()
)

func main() {

	// Connect Database
	dsn := "host=localhost user=postgres password=postgres dbname=mini-ecommerce port=5432 sslmode=disable"
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database: ", err)
	}

	err = db.AutoMigrate(&auth.User{})
	if err != nil {
		log.Fatal("failed to migrate database: ", err)
	}

	// Connect Redis
	rdb = redis.NewClient((&redis.Options{
		Addr: "localhost:6379",
	}))
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatal("failed to connect redis: ", err)
	}

	// Setup Gin
	r := gin.Default()

	// Test Endpoint
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// User Register
	r.POST("/auth/register", func(c *gin.Context) {
		var input struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		hashed, err := auth.HashPassword(input.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "failed to hash password",
			})
			return
		}

		user := auth.User{
			Email:    input.Email,
			Password: hashed,
		}

		if err := db.Create(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create user"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "user registered",
			"user_id": user.ID,
		})

	})

	//User Login
	r.POST("/auth/login", func(c *gin.Context) {
		var input struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var user auth.User
		if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid credentials"})
			return
		}

		if !auth.CheckPassword(user.Password, input.Password) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid credentials"})
			return
		}

		// generate session_id
		sessionID := uuid.New().String()
		if err := rdb.Set(ctx, "session:"+sessionID, user.ID, 3600*time.Second).Err(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
			return
		}

		// set cookie
		c.SetCookie("session_id", sessionID, 3600, "/", "localhost", false, true)

		c.JSON(http.StatusOK, gin.H{"message": "login successful"})
	})

	r.GET("/me", middleware.AuthMiddleware(rdb, ctx), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{
			"message": "you are logged in",
			"user_id": userID,
		})
	})

	fmt.Println(("Server running at http://localhost:8080"))
	if err := r.Run(":8080"); err != nil {
		log.Fatal((err))
	}

}
