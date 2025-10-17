package response

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestResponseHelper_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	logger := zaptest.NewLogger(t)
	helper := NewResponseHelper(logger)
	helper.Success(c, gin.H{"message": "test"})

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "success")
	assert.Contains(t, w.Body.String(), "test")
}

func TestResponseHelper_BadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	logger := zaptest.NewLogger(t)
	helper := NewResponseHelper(logger)
	helper.BadRequest(c, "Invalid input", "validation error")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "error")
	assert.Contains(t, w.Body.String(), "Invalid input")
}

func TestResponseHelper_InternalError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	logger := zaptest.NewLogger(t)
	helper := NewResponseHelper(logger)
	helper.InternalError(c, "Internal server error", "database error")

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "error")
	assert.Contains(t, w.Body.String(), "Internal server error")
}