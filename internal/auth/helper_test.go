package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashPassword(t *testing.T) {
	t.Run("should hash password successfully", func(t *testing.T) {
		password := "mySecurePassword123"

		hashed, err := HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hashed)
		assert.NotEqual(t, password, hashed, "hashed password should not equal plain password")
	})

	t.Run("should produce different hashes for same password", func(t *testing.T) {
		password := "mySecurePassword123"

		hash1, err1 := HashPassword(password)
		hash2, err2 := HashPassword(password)

		require.NoError(t, err1)
		require.NoError(t, err2)
		assert.NotEqual(t, hash1, hash2, "bcrypt should produce different salts")
	})

	t.Run("should hash empty password", func(t *testing.T) {
		password := ""

		hashed, err := HashPassword(password)

		require.NoError(t, err)
		assert.NotEmpty(t, hashed)
	})
}

func TestCheckPassword(t *testing.T) {
	t.Run("should return true for correct password", func(t *testing.T) {
		password := "mySecurePassword123"
		hashed, err := HashPassword(password)
		require.NoError(t, err)

		result := CheckPassword(hashed, password)

		assert.True(t, result, "correct password should return true")
	})

	t.Run("should return false for incorrect password", func(t *testing.T) {
		password := "mySecurePassword123"
		wrongPassword := "wrongPassword456"
		hashed, err := HashPassword(password)
		require.NoError(t, err)

		result := CheckPassword(hashed, wrongPassword)

		assert.False(t, result, "incorrect password should return false")
	})

	t.Run("should return false for empty password against hashed", func(t *testing.T) {
		password := "mySecurePassword123"
		hashed, err := HashPassword(password)
		require.NoError(t, err)

		result := CheckPassword(hashed, "")

		assert.False(t, result, "empty password should not match")
	})

	t.Run("should return false for invalid hash format", func(t *testing.T) {
		invalidHash := "not-a-valid-bcrypt-hash"
		password := "mySecurePassword123"

		result := CheckPassword(invalidHash, password)

		assert.False(t, result, "invalid hash should return false")
	})

	t.Run("should handle case sensitivity", func(t *testing.T) {
		password := "MyPassword123"
		hashed, err := HashPassword(password)
		require.NoError(t, err)

		resultLower := CheckPassword(hashed, "mypassword123")
		resultUpper := CheckPassword(hashed, "MYPASSWORD123")

		assert.False(t, resultLower, "passwords should be case-sensitive")
		assert.False(t, resultUpper, "passwords should be case-sensitive")
	})
}
