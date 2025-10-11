package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewJWTManager(t *testing.T) {
	t.Run("should create JWTManager successfully", func(t *testing.T) {
		secret := "test-secret"
		duration := time.Hour
		logger := zap.NewNop()

		jwtManager := NewJWTManager(secret, duration, logger)

		assert.NotNil(t, jwtManager)
	})
}

func TestJWTManager_Generate(t *testing.T) {
	secret := "test-secret"
	duration := time.Hour
	logger := zap.NewNop()
	jwtManager := NewJWTManager(secret, duration, logger)

	t.Run("should generate token successfully", func(t *testing.T) {
		userID := uint(123)

		token, err := jwtManager.Generate(userID)

		require.NoError(t, err)
		assert.NotEmpty(t, token)
	})

	t.Run("should generate different tokens for same user", func(t *testing.T) {
		userID := uint(123)

		token1, err1 := jwtManager.Generate(userID)
		time.Sleep(time.Second)
		token2, err2 := jwtManager.Generate(userID)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, token1, token2, "tokens should be different due to different issuance times")
	})

	t.Run("should generate token with correct claims", func(t *testing.T) {
		userID := uint(456)

		tokenString, err := jwtManager.Generate(userID)
		require.NoError(t, err)

		token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(token *jwt.Token) (any, error) {
			return []byte(secret), nil
		})

		require.NoError(t, err)
		claims, ok := token.Claims.(*UserClaims)
		require.True(t, ok)
		assert.Equal(t, userID, claims.UserID)
		assert.NotNil(t, claims.ExpiresAt)
		assert.NotNil(t, claims.IssuedAt)
	})
}

func TestJWTManager_Verify(t *testing.T) {
	secret := "test-secret"
	duration := time.Hour
	logger := zap.NewNop()
	jwtManager := NewJWTManager(secret, duration, logger)

	t.Run("should verify valid token successfully", func(t *testing.T) {
		userID := uint(123)
		token, err := jwtManager.Generate(userID)
		require.NoError(t, err)

		claims, err := jwtManager.Verify(token)

		require.NoError(t, err)
		assert.NotNil(t, claims)
		assert.Equal(t, userID, claims.UserID)
	})

	t.Run("should return error for invalid token format", func(t *testing.T) {
		invalidToken := "invalid.token.format"

		claims, err := jwtManager.Verify(invalidToken)

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Equal(t, ErrInvalidToken, err)
	})

	t.Run("should return error for expired token", func(t *testing.T) {
		shortDuration := time.Millisecond
		shortJWTManager := NewJWTManager(secret, shortDuration, logger)

		userID := uint(123)
		token, err := shortJWTManager.Generate(userID)
		require.NoError(t, err)

		time.Sleep(10 * time.Millisecond)

		claims, err := shortJWTManager.Verify(token)

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Equal(t, ErrExpiredToken, err)
	})

	t.Run("should return error for token with wrong secret", func(t *testing.T) {
		userID := uint(123)
		token, err := jwtManager.Generate(userID)
		require.NoError(t, err)

		differentJWTManager := NewJWTManager("different-secret", duration, logger)

		claims, err := differentJWTManager.Verify(token)

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Equal(t, ErrInvalidToken, err)
	})

	t.Run("should return error for malformed token", func(t *testing.T) {
		malformedToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.malformed.signature"

		claims, err := jwtManager.Verify(malformedToken)

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Equal(t, ErrInvalidToken, err)
	})

	t.Run("should return error for empty token", func(t *testing.T) {
		claims, err := jwtManager.Verify("")

		assert.Error(t, err)
		assert.Nil(t, claims)
		assert.Equal(t, ErrInvalidToken, err)
	})

	t.Run("should verify token with different signing method returns error", func(t *testing.T) {
		claims := UserClaims{
			UserID: uint(123),
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		tokenString, _ := token.SignedString(jwt.UnsafeAllowNoneSignatureType)

		verifiedClaims, err := jwtManager.Verify(tokenString)

		assert.Error(t, err)
		assert.Nil(t, verifiedClaims)
		assert.Equal(t, ErrInvalidToken, err)
	})
}
