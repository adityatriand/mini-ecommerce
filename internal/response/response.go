package response

import (
	"mini-e-commerce/internal/constants"

	"github.com/gin-gonic/gin"
)

type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Success bool       `json:"success"`
	Message string     `json:"message"`
	Error   ErrorInfo  `json:"error"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Details string `json:"details"`
}

func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	response := &SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	c.JSON(statusCode, response)
}

func SuccessOK(c *gin.Context, message string, data interface{}) {
	Success(c, constants.StatusOK, message, data)
}

func SuccessCreated(c *gin.Context, message string, data interface{}) {
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
