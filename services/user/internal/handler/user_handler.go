package handler

import (
	"strconv"

	"mini-e-commerce/services/user/internal/model"
	"mini-e-commerce/services/user/internal/service"
	"mini-e-commerce/shared/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type UserHandler struct {
	service        service.UserService
	logger         *zap.Logger
	responseHelper *response.ResponseHelper
}

func NewUserHandler(service service.UserService, logger *zap.Logger) *UserHandler {
	return &UserHandler{
		service:        service,
		logger:         logger,
		responseHelper: response.NewResponseHelper(logger),
	}
}

func (h *UserHandler) RegisterRoutes(r *gin.RouterGroup) {
	users := r.Group("/users")
	{
		users.POST("/register", h.RegisterUser)
		users.POST("/login", h.LoginUser)
		users.POST("/refresh", h.RefreshToken)
		users.POST("/logout", h.LogoutUser)
		
		// Protected routes (require authentication)
		users.GET("/:id", h.GetUserByID)
		users.PUT("/:id", h.UpdateUser)
		users.DELETE("/:id", h.DeleteUser)
		users.GET("", h.GetAllUsers)
	}
}

// RegisterUser handles user registration
func (h *UserHandler) RegisterUser(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHelper.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	authResponse, err := h.service.RegisterUser(c.Request.Context(), req)
	if err != nil {
		h.responseHelper.BadRequest(c, "Registration failed", err.Error())
		return
	}

	h.responseHelper.Created(c, authResponse, "User registered successfully")
}

// LoginUser handles user login
func (h *UserHandler) LoginUser(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHelper.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	authResponse, err := h.service.LoginUser(c.Request.Context(), req)
	if err != nil {
		h.responseHelper.Unauthorized(c, "Login failed: "+err.Error())
		return
	}

	h.responseHelper.Success(c, authResponse, "Login successful")
}

// GetUserByID handles getting user by ID
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.responseHelper.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	user, err := h.service.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		if err.Error() == "user not found" {
			h.responseHelper.NotFound(c, "User not found")
		} else {
			h.responseHelper.InternalError(c, "Failed to get user", err.Error())
		}
		return
	}

	h.responseHelper.Success(c, user)
}

// UpdateUser handles user update
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.responseHelper.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHelper.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	user, err := h.service.UpdateUser(c.Request.Context(), uint(id), req)
	if err != nil {
		if err.Error() == "user not found" {
			h.responseHelper.NotFound(c, "User not found")
		} else if err.Error() == "email already exists" {
			h.responseHelper.Conflict(c, "Email already exists", err.Error())
		} else {
			h.responseHelper.InternalError(c, "Failed to update user", err.Error())
		}
		return
	}

	h.responseHelper.Success(c, user, "User updated successfully")
}

// DeleteUser handles user deletion
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		h.responseHelper.BadRequest(c, "Invalid user ID", err.Error())
		return
	}

	err = h.service.DeleteUser(c.Request.Context(), uint(id))
	if err != nil {
		if err.Error() == "user not found" {
			h.responseHelper.NotFound(c, "User not found")
		} else {
			h.responseHelper.InternalError(c, "Failed to delete user", err.Error())
		}
		return
	}

	h.responseHelper.Success(c, nil, "User deleted successfully")
}

// GetAllUsers handles getting all users with pagination
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	users, total, err := h.service.GetAllUsers(c.Request.Context(), limit, offset)
	if err != nil {
		h.responseHelper.InternalError(c, "Failed to get users", err.Error())
		return
	}

	// Calculate pagination metadata
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	currentPage := (offset / limit) + 1

	meta := response.PaginationMeta{
		Page:       currentPage,
		PageSize:   limit,
		Total:      total,
		TotalPages: totalPages,
	}

	h.responseHelper.SuccessWithMeta(c, users, meta)
}

// RefreshToken handles token refresh
func (h *UserHandler) RefreshToken(c *gin.Context) {
	var req struct {
		UserID       uint   `json:"user_id" binding:"required"`
		SessionID    string `json:"session_id" binding:"required"`
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHelper.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	authResponse, err := h.service.RefreshToken(c.Request.Context(), req.UserID, req.SessionID, req.RefreshToken)
	if err != nil {
		h.responseHelper.Unauthorized(c, "Token refresh failed: "+err.Error())
		return
	}

	h.responseHelper.Success(c, authResponse, "Token refreshed successfully")
}

// LogoutUser handles user logout
func (h *UserHandler) LogoutUser(c *gin.Context) {
	var req struct {
		UserID    uint   `json:"user_id" binding:"required"`
		SessionID string `json:"session_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.responseHelper.BadRequest(c, "Invalid request body", err.Error())
		return
	}

	err := h.service.LogoutUser(c.Request.Context(), req.UserID, req.SessionID)
	if err != nil {
		h.responseHelper.InternalError(c, "Logout failed", err.Error())
		return
	}

	h.responseHelper.Success(c, nil, "Logged out successfully")
}
