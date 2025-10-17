package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// Interface defines the cache interface
type Interface interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeletePattern(ctx context.Context, pattern string) error
	Ping(ctx context.Context) error
}

// RedisCache implements the cache interface using Redis
type RedisCache struct {
	client *redis.Client
	logger *zap.Logger
}

// Config represents Redis configuration
type Config struct {
	Addr     string
	Password string
	DB       int
}

func NewRedisCache(config Config, logger *zap.Logger) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Error("Failed to connect to Redis", zap.Error(err))
		return nil, err
	}

	logger.Info("Redis connection established successfully",
		zap.String("addr", config.Addr),
		zap.Int("db", config.DB),
	)

	return &RedisCache{
		client: client,
		logger: logger,
	}, nil
}

// Get retrieves a value from cache
func (r *RedisCache) Get(ctx context.Context, key string, dest any) error {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			r.logger.Debug("Cache miss", zap.String("key", key))
		} else {
			r.logger.Error("Cache get error", zap.String("key", key), zap.Error(err))
		}
		return err
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		r.logger.Error("Cache unmarshal error", zap.String("key", key), zap.Error(err))
		return err
	}

	r.logger.Debug("Cache hit", zap.String("key", key))
	return nil
}

// Set stores a value in cache
func (r *RedisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		r.logger.Error("Cache marshal error", zap.String("key", key), zap.Error(err))
		return err
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		r.logger.Error("Cache set error", zap.String("key", key), zap.Error(err))
		return err
	}

	r.logger.Debug("Cache set", zap.String("key", key), zap.Duration("ttl", ttl))
	return nil
}

// Delete removes keys from cache
func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		r.logger.Error("Cache delete error", zap.Strings("keys", keys), zap.Error(err))
		return err
	}

	r.logger.Debug("Cache deleted", zap.Strings("keys", keys))
	return nil
}

// DeletePattern removes keys matching a pattern
func (r *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		r.logger.Error("Cache scan error", zap.String("pattern", pattern), zap.Error(err))
		return err
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			r.logger.Error("Cache delete pattern error", 
				zap.String("pattern", pattern), 
				zap.Int("count", len(keys)), 
				zap.Error(err))
			return err
		}
		r.logger.Debug("Cache pattern deleted", 
			zap.String("pattern", pattern), 
			zap.Int("count", len(keys)))
	} else {
		r.logger.Debug("Cache pattern not found", zap.String("pattern", pattern))
	}

	return nil
}

// Ping checks Redis connection
func (r *RedisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}
