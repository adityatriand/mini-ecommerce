package auth

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" validate:"required,email"`
	Password string `json:"password" binding:"required" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" validate:"required,email"`
	Password string `json:"password" binding:"required" validate:"required"`
}

type UpdateUserRequest struct {
	Email *string `json:"email" validate:"omitempty,email"`
}

type AuthResponse struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
}
