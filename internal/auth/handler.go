package auth

import (
	"errors"

	"mini-e-commerce/internal/constants"
	"mini-e-commerce/internal/logger"
	"mini-e-commerce/internal/response"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const (
	// Domain-specific error messages (keep local)
	ErrMsgFailedToRegister   = "Failed to register user"
	ErrMsgFailedToLogin      = "Failed to login user"
	ErrMsgInvalidCredentials = "Invalid credentials"
	ErrMsgFailedToLogout     = "Failed to logout"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
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
		response.BadRequest(c, constants.InvalidInputMessage, err.Error())
		return
	}

	user, err := h.service.RegisterUser(c.Request.Context(), input)
	if err != nil {
		if err.Error() == ErrEmailAlreadyExists {
			response.Error(c, constants.StatusBadRequest, "Email already exists", constants.ErrorCodeValidation, err.Error())
			return
		}
		if err.Error() == ErrWeakPassword {
			response.BadRequest(c, "Password too weak", err.Error())
			return
		}
		response.InternalServerError(c, ErrMsgFailedToRegister, err.Error())
		return
	}

	logger.Info("User registered successfully",
		zap.Uint("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("request_id", c.GetString("request_id")),
	)

	response.SuccessCreated(c, constants.UserRegisteredMessage, gin.H{
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
		response.BadRequest(c, constants.InvalidInputMessage, err.Error())
		return
	}

	user, sessionID, err := h.service.LoginUser(c.Request.Context(), input)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			response.Error(c, constants.StatusUnauthorized, ErrMsgInvalidCredentials, constants.ErrorCodeValidation, err.Error())
			return
		}
		response.InternalServerError(c, ErrMsgFailedToLogin, err.Error())
		return
	}

	// Set cookie with secure settings
	c.SetCookie("session_id", sessionID, int(SessionTimeout.Seconds()), "/", "", false, true)

	logger.Info("User logged in successfully",
		zap.Uint("user_id", user.ID),
		zap.String("email", user.Email),
		zap.String("request_id", c.GetString("request_id")),
	)

	response.SuccessOK(c, constants.LoginSuccessfulMessage, gin.H{
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
		response.BadRequest(c, constants.UnauthorizedMessage, "No session found")
		return
	}

	if err := h.service.LogoutUser(c.Request.Context(), sessionID); err != nil {
		response.InternalServerError(c, ErrMsgFailedToLogout, err.Error())
		return
	}

	// Clear cookie
	c.SetCookie("session_id", "", -1, "/", "", false, true)

	logger.Info("User logged out successfully",
		zap.String("session_id", sessionID),
		zap.String("request_id", c.GetString("request_id")),
	)

	response.SuccessOK(c, constants.LogoutSuccessfulMessage, nil)
}
