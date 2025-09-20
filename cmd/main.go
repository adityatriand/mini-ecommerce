package main

import (
	"context"
	"fmt"
	"log"
	"mini-e-commerce/internal/auth"
	"mini-e-commerce/internal/middleware"
	"mini-e-commerce/internal/product"
	"mini-e-commerce/internal/response"
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

	err = db.AutoMigrate(&product.Product{})
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
			response.Error(
				c,
				http.StatusBadRequest,
				"Failed to register user",
				response.ErrCodeValidationError,
				err.Error(),
			)
			return
		}

		hashed, err := auth.HashPassword(input.Password)
		if err != nil {
			response.Error(
				c,
				http.StatusInternalServerError,
				"Failed to hash password",
				response.ErrCodeInternalServer,
				err.Error(),
			)
			return
		}

		user := auth.User{
			Email:    input.Email,
			Password: hashed,
		}

		if err := db.Create(&user).Error; err != nil {
			response.Error(
				c,
				http.StatusInternalServerError,
				"Failed to register user",
				response.ErrCodeInternalServer,
				err.Error(),
			)
			return
		}

		response.Success(
			c,
			"User registered",
			gin.H{
				"user_id": user.ID,
			},
		)

	})

	// User Login
	r.POST("/auth/login", func(c *gin.Context) {
		var input struct {
			Email    string `json:"email" binding:"required,email"`
			Password string `json:"password" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			response.Error(
				c,
				http.StatusBadRequest,
				"Failed to register user",
				response.ErrCodeValidationError,
				err.Error(),
			)
			return
		}

		var user auth.User
		if err := db.Where("email = ?", input.Email).First(&user).Error; err != nil {
			response.Error(
				c,
				http.StatusInternalServerError,
				"Invalid Credentials",
				response.ErrCodeInvalidCredentials,
				err.Error(),
			)
			return
		}

		if !auth.CheckPassword(user.Password, input.Password) {
			response.Error(
				c,
				http.StatusInternalServerError,
				"Invalid Credentials",
				response.ErrCodeInvalidCredentials,
				err.Error(),
			)
			return
		}

		// generate session_id
		sessionID := uuid.New().String()
		if err := rdb.Set(ctx, "session:"+sessionID, user.ID, 3600*time.Second).Err(); err != nil {
			response.Error(
				c,
				http.StatusInternalServerError,
				"Failed to create session",
				response.ErrCodeInternalServer,
				err.Error(),
			)
			return
		}

		// set cookie
		c.SetCookie("session_id", sessionID, 3600, "/", "localhost", false, true)

		response.Success(
			c,
			"Login successful",
			nil,
		)
	})

	// Check
	r.GET("/me", middleware.AuthMiddleware(rdb, ctx), func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(http.StatusOK, gin.H{
			"message": "you are logged in",
			"user_id": userID,
		})
	})

	// Create Product
	r.POST("/products", middleware.AuthMiddleware(rdb, ctx), func(c *gin.Context) {
		var input struct {
			Name  string `json:"name" binding:"required"`
			Price int    `json:"price" binding:"required"`
			Stock int    `json:"stock" binding:"required"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			response.Error(
				c,
				http.StatusBadRequest,
				"Invalid input request",
				response.ErrCodeValidationError,
				err.Error(),
			)
			return
		}

		product := product.Product{
			Name:  input.Name,
			Price: input.Price,
			Stock: input.Stock,
		}

		if err := db.Create(&product).Error; err != nil {
			response.Error(
				c,
				http.StatusInternalServerError,
				"Failed to create product",
				response.ErrCodeInternalServer,
				err.Error(),
			)
			return
		}

		response.Success(
			c,
			"Product created successfully",
			product,
		)
	})

	// Get Products
	r.GET("/products", func(c *gin.Context) {
		var products []product.Product
		if err := db.Find(&products).Error; err != nil {
			response.Error(
				c,
				http.StatusInternalServerError,
				"Failed to fetch products",
				response.ErrCodeInternalServer,
				err.Error(),
			)
			return
		}
		response.Success(
			c,
			"Products fetched successfully",
			products,
		)
	})

	// Get By Id Products
	r.GET("/products/:id", func(c *gin.Context) {
		var product product.Product
		if err := db.First(&product, c.Param("id")).Error; err != nil {
			response.Error(
				c,
				http.StatusNotFound,
				"Failed to fetch product",
				response.ErrCodeDataNotFound,
				err.Error(),
			)
			return
		}
		response.Success(
			c,
			"Product fetched successfully",
			product,
		)
	})

	// Delete Product
	r.DELETE("/products/:id", middleware.AuthMiddleware(rdb, ctx), func(c *gin.Context) {
		var prod product.Product
		if err := db.First(&prod, c.Param("id")).Error; err != nil {
			response.Error(
				c,
				http.StatusNotFound,
				"Failed to fetch product",
				response.ErrCodeDataNotFound,
				err.Error(),
			)
			return
		}

		if err := db.Delete(&prod).Error; err != nil {
			response.Error(
				c,
				http.StatusInternalServerError,
				"Failed to delete product",
				response.ErrCodeInternalServer,
				err.Error(),
			)
			return
		}

		response.Success(
			c,
			"Product deleted successfully",
			nil,
		)
	})

	// Update Product
	r.PATCH("/products/:id", middleware.AuthMiddleware(rdb, ctx), func(c *gin.Context) {
		var input struct {
			Name  *string `json:"name"`
			Price *int    `json:"price"`
			Stock *int    `json:"stock"`
		}

		if err := c.ShouldBindJSON(&input); err != nil {
			response.Error(
				c,
				http.StatusBadRequest,
				"Invalid input request",
				response.ErrCodeValidationError,
				err.Error(),
			)
			return
		}

		var product product.Product
		if err := db.First(&product, c.Param("id")).Error; err != nil {
			response.Error(
				c,
				http.StatusNotFound,
				"Failed to fetch product",
				response.ErrCodeDataNotFound,
				err.Error(),
			)
			return
		}

		// update hanya field yang dikirim
		if input.Name != nil {
			product.Name = *input.Name
		}
		if input.Price != nil {
			product.Price = *input.Price
		}
		if input.Stock != nil {
			product.Stock = *input.Stock
		}

		if err := db.Save(&product).Error; err != nil {
			response.Error(
				c,
				http.StatusInternalServerError,
				"Failed to update product",
				response.ErrCodeInternalServer,
				err.Error(),
			)
			return
		}

		response.Success(
			c,
			"Product updated successfully",
			product,
		)
	})

	fmt.Println(("Server running at http://localhost:8080"))
	if err := r.Run(":8080"); err != nil {
		log.Fatal((err))
	}

}
