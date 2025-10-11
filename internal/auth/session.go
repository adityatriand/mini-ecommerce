package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var (
	ErrSessionNotFound     = errors.New("session not found")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrSessionStoreFailed  = errors.New("failed to store session")
	ErrSessionDeleteFailed = errors.New("failed to delete session")
)

type SessionManagerInterface interface {
	StoreRefreshToken(ctx context.Context, userID uint, sessionID, token string, ttl time.Duration) error
	ValidateRefreshToken(ctx context.Context, userID uint, sessionID, token string) error
	DeleteRefreshToken(ctx context.Context, userID uint, sessionID string) error
	GetSessionKey(userID uint, sessionID string) string
}

type SessionManager struct {
	client *redis.Client
	logger *zap.Logger
}

func NewSessionManager(client *redis.Client, logger *zap.Logger) SessionManagerInterface {
	return &SessionManager{
		client: client,
		logger: logger,
	}
}

func (s *SessionManager) StoreRefreshToken(ctx context.Context, userID uint, sessionID, token string, ttl time.Duration) error {
	key := fmt.Sprintf("session:%d:%s", userID, sessionID)
	if err := s.client.Set(ctx, key, token, ttl).Err(); err != nil {
		s.logger.Error("Failed to store refresh token",
			zap.Error(err),
			zap.Uint("user_id", userID),
			zap.String("session_id", sessionID),
		)
		return ErrSessionStoreFailed
	}

	s.logger.Debug("Refresh token stored successfully",
		zap.Uint("user_id", userID),
		zap.String("session_id", sessionID),
		zap.Duration("ttl", ttl),
	)
	return nil
}

func (s *SessionManager) ValidateRefreshToken(ctx context.Context, userID uint, sessionID, token string) error {
	key := fmt.Sprintf("session:%d:%s", userID, sessionID)
	val, err := s.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			s.logger.Warn("Session not found",
				zap.Uint("user_id", userID),
				zap.String("session_id", sessionID),
			)
			return ErrSessionNotFound
		}
		s.logger.Error("Failed to get session",
			zap.Error(err),
			zap.Uint("user_id", userID),
			zap.String("session_id", sessionID),
		)
		return err
	}

	if val != token {
		s.logger.Warn("Invalid refresh token provided",
			zap.Uint("user_id", userID),
			zap.String("session_id", sessionID),
		)
		return ErrInvalidRefreshToken
	}

	s.logger.Debug("Refresh token validated successfully",
		zap.Uint("user_id", userID),
		zap.String("session_id", sessionID),
	)
	return nil
}

func (s *SessionManager) DeleteRefreshToken(ctx context.Context, userID uint, sessionID string) error {
	key := fmt.Sprintf("session:%d:%s", userID, sessionID)
	if err := s.client.Del(ctx, key).Err(); err != nil {
		s.logger.Error("Failed to delete session",
			zap.Error(err),
			zap.Uint("user_id", userID),
			zap.String("session_id", sessionID),
		)
		return ErrSessionDeleteFailed
	}

	s.logger.Debug("Refresh token deleted successfully",
		zap.Uint("user_id", userID),
		zap.String("session_id", sessionID),
	)
	return nil
}

func (s *SessionManager) GetSessionKey(userID uint, sessionID string) string {
	return fmt.Sprintf("session:%d:%s", userID, sessionID)
}
