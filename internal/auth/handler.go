package auth

import (
	"errors"
	"net/http"

	"mini-e-commerce/internal/response"

	"github.com/gin-gonic/gin"
)

const (
	ErrMsgInvalidInput       = "Invalid input"
	ErrMsgFailedToRegister   = "Failed to register user"
	ErrMsgFailedToLogin      = "Failed to login user"
	ErrMsgInvalidCredentials = "Invalid credentials"
	ErrMsgFailedToLogout     = "Failed to logout"
	ErrMsgUnauthorized       = "Unauthorized"

	// Success messages
	MsgUserRegistered   = "User registered successfully"
	MsgLoginSuccessful  = "Login successful"
	MsgLogoutSuccessful = "Logout successful"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/auth")
	{
		group.POST("/register", h.Register)
		group.POST("/login", h.Login)
		group.POST("/logout", h.Logout)
	}
}

func (h *Handler) Register(c *gin.Context) {
	var input RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, ErrMsgInvalidInput, response.ErrCodeValidationError, err.Error())
		return
	}

	user, err := h.service.RegisterUser(c.Request.Context(), input)
	if err != nil {
		if err.Error() == ErrEmailAlreadyExists {
			response.Error(c, http.StatusConflict, "Email already exists", response.ErrCodeDataAlreadyExists, err.Error())
			return
		}
		if err.Error() == ErrWeakPassword {
			response.Error(c, http.StatusBadRequest, "Password too weak", response.ErrCodeValidationError, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, ErrMsgFailedToRegister, response.ErrCodeInternalServer, err.Error())
		return
	}

	response.Success(c, MsgUserRegistered, gin.H{
		"user_id": user.ID,
		"email":   user.Email,
	})
}

func (h *Handler) Login(c *gin.Context) {
	var input LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		response.Error(c, http.StatusBadRequest, ErrMsgInvalidInput, response.ErrCodeValidationError, err.Error())
		return
	}

	user, sessionID, err := h.service.LoginUser(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Error(c, http.StatusUnauthorized, ErrMsgInvalidCredentials, response.ErrCodeInvalidCredentials, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, ErrMsgFailedToLogin, response.ErrCodeInternalServer, err.Error())
		return
	}

	// Set cookie with secure settings
	c.SetCookie("session_id", sessionID, int(SessionTimeout.Seconds()), "/", "", false, true)

	response.Success(c, MsgLoginSuccessful, gin.H{
		"user_id": user.ID,
		"email":   user.Email,
	})
}

func (h *Handler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		response.Error(c, http.StatusBadRequest, ErrMsgUnauthorized, response.ErrCodeUnauthorized, "No session found")
		return
	}

	if err := h.service.LogoutUser(c.Request.Context(), sessionID); err != nil {
		response.Error(c, http.StatusInternalServerError, ErrMsgFailedToLogout, response.ErrCodeInternalServer, err.Error())
		return
	}

	// Clear cookie
	c.SetCookie("session_id", "", -1, "/", "", false, true)

	response.Success(c, MsgLogoutSuccessful, nil)
}
