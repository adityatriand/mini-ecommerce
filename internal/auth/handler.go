package auth

import (
	"errors"
	"net/http"

	"mini-e-commerce/internal/logger"
	"mini-e-commerce/internal/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	ErrMsgFailedToRegister   = "Failed to register user"
	ErrMsgFailedToLogin      = "Failed to login user"
	ErrMsgInvalidCredentials = "Invalid credentials"
	ErrMsgFailedToLogout     = "Failed to logout"
)

type Handler struct {
	service        Service
	logger         logger.Logger
	responseHelper *response.ResponseHelper
}

func NewHandler(service Service, log logger.Logger) *Handler {
	return &Handler{
		service:        service,
		logger:         log,
		responseHelper: response.NewResponseHelper(log),
	}
}

func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	group := r.Group("/auth")
	{
		group.POST("/register", h.Register)
		group.POST("/login", h.Login)
		group.POST("/logout", h.Logout)
	}
}

// RegisterUser godoc
// @Summary Create new user
// @Description Create a new user with username and password
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param   request body RegisterRequest true "User request body"
// @Success 201 {object} response.SuccessResponse{data=AuthResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/register [post]
func (h *Handler) Register(c *gin.Context) {
	var input RegisterRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		h.responseHelper.BadRequest(c, response.ErrCodeValidationError, err.Error())
		return
	}

	user, err := h.service.RegisterUser(c.Request.Context(), input)
	if err != nil {
		if err.Error() == ErrEmailAlreadyExists {
			h.responseHelper.BadRequest(c, "Email already exists", err.Error())
			return
		}
		if err.Error() == ErrWeakPassword {
			h.responseHelper.BadRequest(c, "Password too weak", err.Error())
			return
		}
		h.responseHelper.InternalServerError(c, ErrMsgFailedToRegister, err.Error())
		return
	}

	h.logger.Info("User registered",
		zap.Uint("user_id", user.ID),
		zap.String("email", user.Email),
	)

	h.responseHelper.SuccessCreated(c, "User created successfully", gin.H{
		"user_id": user.ID,
		"email":   user.Email,
	})
}

// AuthLogin godoc
// @Summary Auth for login user
// @Description Authentication login user
// @Tags Auth
// @Accept  json
// @Produce  json
// @Param   request body LoginRequest true "Login request body"
// @Success 201 {object} response.SuccessResponse{data=AuthResponse}
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/login [post]
func (h *Handler) Login(c *gin.Context) {
	var input LoginRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		h.responseHelper.BadRequest(c, response.ErrCodeValidationError, err.Error())
		return
	}

	user, sessionID, err := h.service.LoginUser(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			h.responseHelper.Error(c, http.StatusUnauthorized, ErrMsgInvalidCredentials, response.ErrCodeInvalidCredentials, err.Error())
			return
		}
		h.responseHelper.InternalServerError(c, ErrMsgFailedToLogin, err.Error())
		return
	}

	c.SetCookie("session_id", sessionID, int(SessionTimeout.Seconds()), "/", "", false, true)

	h.logger.Info("User session created",
		zap.Uint("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("session_id", sessionID),
	)

	h.responseHelper.SuccessOK(c, "Login successfully", gin.H{
		"user_id": user.ID,
		"email":   user.Email,
	})
}

// AuthLogout godoc
// @Summary Logout user
// @Description Logout the current authenticated user by invalidating their session
// @Tags Auth
// @Accept  json
// @Produce  json
// @Success 200 {object} response.SuccessResponse
// @Failure 400 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /auth/logout [post]
func (h *Handler) Logout(c *gin.Context) {
	sessionID, err := c.Cookie("session_id")
	if err != nil {
		h.responseHelper.BadRequest(c, response.ErrCodeUnauthorized, "No session found")
		return
	}

	if err := h.service.LogoutUser(c.Request.Context(), sessionID); err != nil {
		h.responseHelper.InternalServerError(c, ErrMsgFailedToLogout, err.Error())
		return
	}

	c.SetCookie("session_id", "", -1, "/", "", false, true)

	h.logger.Info("User session terminated",
		zap.String("session_id", sessionID),
	)

	h.responseHelper.SuccessOK(c, "Logout successfully", nil)
}
