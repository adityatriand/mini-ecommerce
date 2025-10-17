package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ResponseHelper provides standardized API response methods
type ResponseHelper struct {
	logger *zap.Logger
}

func NewResponseHelper(logger *zap.Logger) *ResponseHelper {
	return &ResponseHelper{
		logger: logger,
	}
}

// SuccessResponse represents a successful API response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// ErrorResponse represents an error API response
type ErrorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Error   interface{} `json:"error,omitempty"`
}

// PaginationMeta represents pagination metadata
type PaginationMeta struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// Success sends a successful response
func (r *ResponseHelper) Success(c *gin.Context, data interface{}, message ...string) {
	msg := "Success"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	response := SuccessResponse{
		Success: true,
		Message: msg,
		Data:    data,
	}

	c.JSON(http.StatusOK, response)
}

// Created sends a created response
func (r *ResponseHelper) Created(c *gin.Context, data interface{}, message ...string) {
	msg := "Created successfully"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	response := SuccessResponse{
		Success: true,
		Message: msg,
		Data:    data,
	}

	c.JSON(http.StatusCreated, response)
}

// SuccessWithMeta sends a successful response with metadata
func (r *ResponseHelper) SuccessWithMeta(c *gin.Context, data interface{}, meta interface{}, message ...string) {
	msg := "Success"
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}

	response := SuccessResponse{
		Success: true,
		Message: msg,
		Data:    data,
		Meta:    meta,
	}

	c.JSON(http.StatusOK, response)
}

// BadRequest sends a bad request error response
func (r *ResponseHelper) BadRequest(c *gin.Context, message string, err interface{}) {
	response := ErrorResponse{
		Success: false,
		Message: message,
		Error:   err,
	}

	r.logger.Warn("Bad request", 
		zap.String("path", c.Request.URL.Path),
		zap.String("message", message),
		zap.Any("error", err))

	c.JSON(http.StatusBadRequest, response)
}

// Unauthorized sends an unauthorized error response
func (r *ResponseHelper) Unauthorized(c *gin.Context, message string) {
	response := ErrorResponse{
		Success: false,
		Message: message,
	}

	r.logger.Warn("Unauthorized access", 
		zap.String("path", c.Request.URL.Path),
		zap.String("message", message))

	c.JSON(http.StatusUnauthorized, response)
}

// Forbidden sends a forbidden error response
func (r *ResponseHelper) Forbidden(c *gin.Context, message string) {
	response := ErrorResponse{
		Success: false,
		Message: message,
	}

	r.logger.Warn("Forbidden access", 
		zap.String("path", c.Request.URL.Path),
		zap.String("message", message))

	c.JSON(http.StatusForbidden, response)
}

// NotFound sends a not found error response
func (r *ResponseHelper) NotFound(c *gin.Context, message string) {
	response := ErrorResponse{
		Success: false,
		Message: message,
	}

	r.logger.Warn("Resource not found", 
		zap.String("path", c.Request.URL.Path),
		zap.String("message", message))

	c.JSON(http.StatusNotFound, response)
}

// Conflict sends a conflict error response
func (r *ResponseHelper) Conflict(c *gin.Context, message string, err interface{}) {
	response := ErrorResponse{
		Success: false,
		Message: message,
		Error:   err,
	}

	r.logger.Warn("Conflict", 
		zap.String("path", c.Request.URL.Path),
		zap.String("message", message),
		zap.Any("error", err))

	c.JSON(http.StatusConflict, response)
}

// InternalError sends an internal server error response
func (r *ResponseHelper) InternalError(c *gin.Context, message string, err interface{}) {
	response := ErrorResponse{
		Success: false,
		Message: message,
		Error:   err,
	}

	r.logger.Error("Internal server error", 
		zap.String("path", c.Request.URL.Path),
		zap.String("message", message),
		zap.Any("error", err))

	c.JSON(http.StatusInternalServerError, response)
}

// ServiceUnavailable sends a service unavailable error response
func (r *ResponseHelper) ServiceUnavailable(c *gin.Context, message string) {
	response := ErrorResponse{
		Success: false,
		Message: message,
	}

	r.logger.Error("Service unavailable", 
		zap.String("path", c.Request.URL.Path),
		zap.String("message", message))

	c.JSON(http.StatusServiceUnavailable, response)
}

// CustomError sends a custom error response
func (r *ResponseHelper) CustomError(c *gin.Context, statusCode int, message string, err interface{}) {
	response := ErrorResponse{
		Success: false,
		Message: message,
		Error:   err,
	}

	r.logger.Error("Custom error", 
		zap.String("path", c.Request.URL.Path),
		zap.Int("status_code", statusCode),
		zap.String("message", message),
		zap.Any("error", err))

	c.JSON(statusCode, response)
}
