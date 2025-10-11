package auth

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

func TestNewSessionManager(t *testing.T) {
	t.Run("should create SessionManager successfully", func(t *testing.T) {
		client, mr := setupTestRedis(t)
		defer mr.Close()
		logger := zap.NewNop()

		sessionManager := NewSessionManager(client, logger)

		assert.NotNil(t, sessionManager)
	})
}

func TestSessionManager_StoreRefreshToken(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	logger := zap.NewNop()
	sessionManager := NewSessionManager(client, logger)
	ctx := context.Background()

	t.Run("should store refresh token successfully", func(t *testing.T) {
		userID := uint(123)
		sessionID := "session-123"
		token := "refresh-token-123"
		ttl := time.Hour

		err := sessionManager.StoreRefreshToken(ctx, userID, sessionID, token, ttl)

		require.NoError(t, err)

		key := sessionManager.GetSessionKey(userID, sessionID)
		storedToken, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		assert.Equal(t, token, storedToken)
	})

	t.Run("should store refresh token with correct TTL", func(t *testing.T) {
		userID := uint(456)
		sessionID := "session-456"
		token := "refresh-token-456"
		ttl := 5 * time.Second

		err := sessionManager.StoreRefreshToken(ctx, userID, sessionID, token, ttl)

		require.NoError(t, err)

		key := sessionManager.GetSessionKey(userID, sessionID)
		actualTTL, err := client.TTL(ctx, key).Result()
		require.NoError(t, err)
		assert.True(t, actualTTL > 0 && actualTTL <= ttl)
	})

	t.Run("should overwrite existing token", func(t *testing.T) {
		userID := uint(789)
		sessionID := "session-789"
		token1 := "refresh-token-old"
		token2 := "refresh-token-new"
		ttl := time.Hour

		err := sessionManager.StoreRefreshToken(ctx, userID, sessionID, token1, ttl)
		require.NoError(t, err)

		err = sessionManager.StoreRefreshToken(ctx, userID, sessionID, token2, ttl)
		require.NoError(t, err)

		key := sessionManager.GetSessionKey(userID, sessionID)
		storedToken, err := client.Get(ctx, key).Result()
		require.NoError(t, err)
		assert.Equal(t, token2, storedToken)
	})
}

func TestSessionManager_ValidateRefreshToken(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	logger := zap.NewNop()
	sessionManager := NewSessionManager(client, logger)
	ctx := context.Background()

	t.Run("should validate correct refresh token", func(t *testing.T) {
		userID := uint(123)
		sessionID := "session-123"
		token := "refresh-token-123"
		ttl := time.Hour

		err := sessionManager.StoreRefreshToken(ctx, userID, sessionID, token, ttl)
		require.NoError(t, err)

		err = sessionManager.ValidateRefreshToken(ctx, userID, sessionID, token)
		assert.NoError(t, err)
	})

	t.Run("should return error for non-existent session", func(t *testing.T) {
		userID := uint(999)
		sessionID := "non-existent-session"
		token := "some-token"

		err := sessionManager.ValidateRefreshToken(ctx, userID, sessionID, token)

		assert.Error(t, err)
		assert.Equal(t, ErrSessionNotFound, err)
	})

	t.Run("should return error for invalid refresh token", func(t *testing.T) {
		userID := uint(123)
		sessionID := "session-123"
		correctToken := "correct-token"
		wrongToken := "wrong-token"
		ttl := time.Hour

		err := sessionManager.StoreRefreshToken(ctx, userID, sessionID, correctToken, ttl)
		require.NoError(t, err)

		err = sessionManager.ValidateRefreshToken(ctx, userID, sessionID, wrongToken)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidRefreshToken, err)
	})

	t.Run("should return error for empty token", func(t *testing.T) {
		userID := uint(123)
		sessionID := "session-123"
		correctToken := "correct-token"
		ttl := time.Hour

		err := sessionManager.StoreRefreshToken(ctx, userID, sessionID, correctToken, ttl)
		require.NoError(t, err)

		err = sessionManager.ValidateRefreshToken(ctx, userID, sessionID, "")

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidRefreshToken, err)
	})
}

func TestSessionManager_DeleteRefreshToken(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	logger := zap.NewNop()
	sessionManager := NewSessionManager(client, logger)
	ctx := context.Background()

	t.Run("should delete refresh token successfully", func(t *testing.T) {
		userID := uint(123)
		sessionID := "session-123"
		token := "refresh-token-123"
		ttl := time.Hour

		err := sessionManager.StoreRefreshToken(ctx, userID, sessionID, token, ttl)
		require.NoError(t, err)

		err = sessionManager.DeleteRefreshToken(ctx, userID, sessionID)
		assert.NoError(t, err)

		key := sessionManager.GetSessionKey(userID, sessionID)
		_, err = client.Get(ctx, key).Result()
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("should not return error when deleting non-existent token", func(t *testing.T) {
		userID := uint(999)
		sessionID := "non-existent-session"

		err := sessionManager.DeleteRefreshToken(ctx, userID, sessionID)

		assert.NoError(t, err)
	})

	t.Run("should delete only specific session", func(t *testing.T) {
		userID := uint(123)
		sessionID1 := "session-1"
		sessionID2 := "session-2"
		token1 := "token-1"
		token2 := "token-2"
		ttl := time.Hour

		err := sessionManager.StoreRefreshToken(ctx, userID, sessionID1, token1, ttl)
		require.NoError(t, err)
		err = sessionManager.StoreRefreshToken(ctx, userID, sessionID2, token2, ttl)
		require.NoError(t, err)

		err = sessionManager.DeleteRefreshToken(ctx, userID, sessionID1)
		assert.NoError(t, err)

		key1 := sessionManager.GetSessionKey(userID, sessionID1)
		_, err = client.Get(ctx, key1).Result()
		assert.Equal(t, redis.Nil, err)

		key2 := sessionManager.GetSessionKey(userID, sessionID2)
		storedToken, err := client.Get(ctx, key2).Result()
		require.NoError(t, err)
		assert.Equal(t, token2, storedToken)
	})
}

func TestSessionManager_GetSessionKey(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	logger := zap.NewNop()
	sessionManager := NewSessionManager(client, logger)

	t.Run("should generate correct session key", func(t *testing.T) {
		userID := uint(123)
		sessionID := "session-abc"

		key := sessionManager.GetSessionKey(userID, sessionID)

		assert.Equal(t, "session:123:session-abc", key)
	})

	t.Run("should generate unique keys for different users", func(t *testing.T) {
		sessionID := "session-abc"

		key1 := sessionManager.GetSessionKey(uint(123), sessionID)
		key2 := sessionManager.GetSessionKey(uint(456), sessionID)

		assert.NotEqual(t, key1, key2)
		assert.Equal(t, "session:123:session-abc", key1)
		assert.Equal(t, "session:456:session-abc", key2)
	})

	t.Run("should generate unique keys for different sessions", func(t *testing.T) {
		userID := uint(123)

		key1 := sessionManager.GetSessionKey(userID, "session-1")
		key2 := sessionManager.GetSessionKey(userID, "session-2")

		assert.NotEqual(t, key1, key2)
		assert.Equal(t, "session:123:session-1", key1)
		assert.Equal(t, "session:123:session-2", key2)
	})
}
