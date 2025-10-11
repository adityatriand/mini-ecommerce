package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type JWTManagerInterface interface {
	Generate(userID uint) (string, error)
	Verify(tokenStr string) (*UserClaims, error)
}

type JWTManager struct {
	SecretKey     string
	TokenDuration time.Duration
	logger        *zap.Logger
}

type UserClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func NewJWTManager(secret string, duration time.Duration, logger *zap.Logger) JWTManagerInterface {
	return &JWTManager{
		SecretKey:     secret,
		TokenDuration: duration,
		logger:        logger,
	}
}

func (j *JWTManager) Generate(userID uint) (string, error) {
	claims := UserClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.TokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(j.SecretKey))
	if err != nil {
		j.logger.Error("Failed to generate JWT token", zap.Error(err), zap.Uint("user_id", userID))
		return "", err
	}

	j.logger.Debug("JWT token generated successfully", zap.Uint("user_id", userID))
	return signedToken, nil
}

func (j *JWTManager) Verify(tokenStr string) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &UserClaims{}, func(token *jwt.Token) (any, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(j.SecretKey), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			j.logger.Warn("Token expired", zap.Error(err))
			return nil, ErrExpiredToken
		}
		j.logger.Error("Failed to parse JWT token", zap.Error(err))
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*UserClaims)
	if !ok || !token.Valid {
		j.logger.Error("Invalid token claims")
		return nil, ErrInvalidToken
	}

	j.logger.Debug("JWT token verified successfully", zap.Uint("user_id", claims.UserID))
	return claims, nil
}
