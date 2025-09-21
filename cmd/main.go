package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"mini-e-commerce/internal/auth"
	"mini-e-commerce/internal/middleware"
	"mini-e-commerce/internal/order"
	"mini-e-commerce/internal/product"
	"mini-e-commerce/internal/response"
	"net/http"
	"strconv"
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

	// Run migrations (sementara di sini, untuk production pakai migration tool)
	if err := db.AutoMigrate(&auth.User{}, &product.Product{}, &order.Order{}); err != nil {
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

	// USER
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
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Error(c, http.StatusUnauthorized, "Invalid credentials", response.ErrCodeInvalidCredentials, "user not found")
				return
			}
			response.Error(c, http.StatusInternalServerError, "Failed to query user", response.ErrCodeInternalServer, err.Error())
			return
		}

		if !auth.CheckPassword(user.Password, input.Password) {
			response.Error(c, http.StatusUnauthorized, "Invalid credentials", response.ErrCodeInvalidCredentials, "Invalid password")
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

		response.Success(c, "Login successful", gin.H{
			"user_id":    user.ID,
			"session_id": sessionID,
		})
	})

	// CRUD: Products
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

	// CRUD: Orders
	r.POST("/orders", middleware.AuthMiddleware(rdb, ctx), func(c *gin.Context) {
		var input struct {
			ProductID uint `json:"product_id" binding:"required"`
			Quantity  int  `json:"quantity" binding:"required,gt=0"`
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
		if err := db.First(&product, input.ProductID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				response.Error(
					c,
					http.StatusNotFound,
					"Product not found",
					response.ErrCodeDataNotFound,
					"No product with given ID",
				)
				return
			}

			response.Error(
				c,
				http.StatusInternalServerError,
				"Failed to query product",
				response.ErrCodeInternalServer,
				err.Error(),
			)
			return
		}

		if input.Quantity > product.Stock {
			response.Error(
				c,
				http.StatusBadRequest,
				"Stock product not available",
				response.ErrCodeValidationError,
				"Quantity is more than product stock",
			)
			return
		}

		userIDStr, ok := c.Get("user_id")
		if !ok {
			response.Error(
				c,
				http.StatusUnauthorized,
				"Unauthorized",
				response.ErrCodeUnauthorized,
				"Missing user_id in context",
			)
			return
		}
		idUint, err := strconv.ParseUint(userIDStr.(string), 10, 32)
		if err != nil {
			response.Error(
				c,
				http.StatusInternalServerError,
				"Invalid user id in context",
				response.ErrCodeInternalServer,
				err.Error(),
			)
			return
		}
		uid := uint(idUint)

		order := order.Order{
			UserID:     uid,
			ProductID:  input.ProductID,
			Quantity:   input.Quantity,
			TotalPrice: input.Quantity * product.Price,
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			product.Stock -= input.Quantity
			if err := tx.Save(&product).Error; err != nil {
				return err
			}
			if err := tx.Create(&order).Error; err != nil {
				return err
			}

			return nil
		}); err != nil {
			response.Error(
				c,
				http.StatusInternalServerError,
				"Failed to process order",
				response.ErrCodeInternalServer,
				err.Error(),
			)
			return
		}

		response.Success(
			c,
			"Order created successfully",
			order,
		)
	})
	r.GET("/orders", middleware.AuthMiddleware(rdb, ctx), func(c *gin.Context) {
		var orders []order.Order
		if err := db.Find(&orders).Error; err != nil {
			response.Error(
				c,
				http.StatusInternalServerError,
				"Failed to fetch orders",
				response.ErrCodeInternalServer,
				err.Error(),
			)
			return
		}
		response.Success(
			c,
			"Orders fetched successfully",
			orders,
		)
	})
	r.GET("/orders/:id", middleware.AuthMiddleware(rdb, ctx), func(c *gin.Context) {
		var order order.Order
		if err := db.First(&order, c.Param("id")).Error; err != nil {
			response.Error(
				c,
				http.StatusNotFound,
				"Failed to fetch order",
				response.ErrCodeDataNotFound,
				err.Error(),
			)
			return
		}
		response.Success(
			c,
			"Orders fetched successfully",
			order,
		)
	})
	r.DELETE("/orders/:id", middleware.AuthMiddleware(rdb, ctx), func(c *gin.Context) {
		var ord order.Order
		if err := db.First(&ord, c.Param("id")).Error; err != nil {
			response.Error(c, http.StatusNotFound, "Order not found", response.ErrCodeDataNotFound, err.Error())
			return
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			// restore stock
			var prod product.Product
			if err := tx.First(&prod, ord.ProductID).Error; err != nil {
				return err
			}
			prod.Stock += ord.Quantity
			if err := tx.Save(&prod).Error; err != nil {
				return err
			}
			if err := tx.Delete(&ord).Error; err != nil {
				return err
			}
			return nil
		}); err != nil {
			response.Error(c, http.StatusInternalServerError, "Failed to delete order", response.ErrCodeInternalServer, err.Error())
			return
		}

		response.Success(c, "Order deleted successfully", nil)
	})
	r.PATCH("/orders/:id", middleware.AuthMiddleware(rdb, ctx), func(c *gin.Context) {
		var input struct {
			Quantity int `json:"quantity" binding:"required"`
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

		var order order.Order
		if err := db.First(&order, c.Param("id")).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(
				c,
				http.StatusNotFound,
				"Order not found",
				response.ErrCodeDataNotFound,
				"No order with given ID",
			)
			return
		}

		userIDStr, _ := c.Get("user_id")
		uid, _ := strconv.ParseUint(userIDStr.(string), 10, 32)
		if order.UserID != uint(uid) {
			response.Error(
				c,
				http.StatusForbidden,
				"Not allowed to update this order",
				response.ErrCodeForbidden,
				"Order does not belong to user",
			)
			return
		}

		var product product.Product
		if err := db.First(&product, order.ProductID).Error; err != nil {
			response.Error(
				c,
				http.StatusInternalServerError,
				"Failed to fetch product",
				response.ErrCodeInternalServer,
				err.Error(),
			)
			return
		}

		diff := input.Quantity - order.Quantity
		if diff > 0 && diff > product.Stock {
			response.Error(
				c,
				http.StatusBadRequest,
				"Stock product not available",
				response.ErrCodeValidationError,
				"Quantity exceeds product stock",
			)
			return
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			product.Stock -= diff
			if err := tx.Save(&product).Error; err != nil {
				return err
			}

			order.Quantity = input.Quantity
			order.TotalPrice = input.Quantity * product.Price
			if err := tx.Save(&order).Error; err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			response.Error(
				c,
				http.StatusInternalServerError,
				"Failed to update order",
				response.ErrCodeInternalServer,
				err.Error(),
			)
			return
		}

		response.Success(
			c,
			"Order updated successfully",
			order,
		)
	})

	fmt.Println(("Server running at http://localhost:8080"))
	if err := r.Run(":8080"); err != nil {
		log.Fatal((err))
	}

}
