package response

import (
	"mini-e-commerce/internal/constants"
	"mini-e-commerce/internal/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SuccessResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type ErrorResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Error   ErrorInfo `json:"error"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Details string `json:"details"`
}

func Success(c *gin.Context, statusCode int, message string, data any) {
	response := &SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	// Get request ID safely
	requestID := getRequestID(c)
	userID := getUserID(c)

	logger.Info("API Success Response",
		zap.Int("status_code", statusCode),
		zap.String("message", message),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("user_id", userID),
		zap.String("request_id", requestID),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	c.JSON(statusCode, response)
}

func SuccessOK(c *gin.Context, message string, data any) {
	Success(c, constants.StatusOK, message, data)
}

func SuccessCreated(c *gin.Context, message string, data any) {
	Success(c, constants.StatusCreated, message, data)
}

func Error(c *gin.Context, statusCode int, message string, errorCode string, details string) {
	response := &ErrorResponse{
		Success: false,
		Message: message,
		Error: ErrorInfo{
			Code:    errorCode,
			Details: details,
		},
	}

	// Get request ID and user ID safely
	requestID := getRequestID(c)
	userID := getUserID(c)

	logger.Error("API Error Response",
		zap.Int("status_code", statusCode),
		zap.String("message", message),
		zap.String("error_code", errorCode),
		zap.String("error_details", details),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("request_id", requestID),
		zap.String("user_id", userID),
		zap.String("remote_addr", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	c.JSON(statusCode, response)
}

func BadRequest(c *gin.Context, message string, details string) {
	Error(c, constants.StatusBadRequest, message, constants.ErrorCodeValidation, details)
}

func NotFound(c *gin.Context, message string, details string) {
	Error(c, constants.StatusNotFound, message, constants.ErrorCodeNotFound, details)
}

func InternalServerError(c *gin.Context, message string, details string) {
	Error(c, constants.StatusInternalServerError, message, constants.ErrorCodeInternalServer, details)
}

// Helper functions for safe context extraction
func getRequestID(c *gin.Context) string {
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			return id
		}
	}
	return "unknown"
}

func getUserID(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok {
			return id
		}
	}
	return "anonymous"
}
