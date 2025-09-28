package response

import (
	"mini-e-commerce/internal/logger"
	"net/http"

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

type ResponseHelper struct {
	logger logger.Logger
}

func NewResponseHelper(log logger.Logger) *ResponseHelper {
	return &ResponseHelper{logger: log}
}

func (r *ResponseHelper) Success(c *gin.Context, statusCode int, message string, data any) {
	response := &SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	ctxLogger := r.logger.WithContext(c)
	ctxLogger.Info("API Success Response",
		zap.Int("status_code", statusCode),
		zap.String("message", message),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	c.JSON(statusCode, response)
}

func (r *ResponseHelper) SuccessOK(c *gin.Context, message string, data any) {
	r.Success(c, http.StatusOK, message, data)
}

func (r *ResponseHelper) SuccessCreated(c *gin.Context, message string, data any) {
	r.Success(c, http.StatusCreated, message, data)
}

func (r *ResponseHelper) Error(c *gin.Context, statusCode int, message string, errorCode string, details string) {
	response := &ErrorResponse{
		Success: false,
		Message: message,
		Error: ErrorInfo{
			Code:    errorCode,
			Details: details,
		},
	}

	ctxLogger := r.logger.WithContext(c)
	ctxLogger.Error("API Error Response",
		zap.Int("status_code", statusCode),
		zap.String("message", message),
		zap.String("error_code", errorCode),
		zap.String("error_details", details),
		zap.String("method", c.Request.Method),
		zap.String("path", c.Request.URL.Path),
		zap.String("remote_addr", c.ClientIP()),
		zap.String("user_agent", c.Request.UserAgent()),
	)

	c.JSON(statusCode, response)
}

func (r *ResponseHelper) BadRequest(c *gin.Context, message string, details string) {
	r.Error(c, http.StatusBadRequest, message, ErrCodeValidationError, details)
}

func (r *ResponseHelper) NotFound(c *gin.Context, message string, details string) {
	r.Error(c, http.StatusNotFound, message, ErrCodeDataNotFound, details)
}

func (r *ResponseHelper) InternalServerError(c *gin.Context, message string, details string) {
	r.Error(c, http.StatusInternalServerError, message, ErrCodeInternalServer, details)
}
