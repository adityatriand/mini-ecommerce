package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type RedisCache struct {
	client *redis.Client
	logger *zap.Logger
}

func NewRedisCache(client *redis.Client, logger *zap.Logger) *RedisCache {
	return &RedisCache{
		client: client,
		logger: logger,
	}
}

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

func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		r.logger.Error("Cache delete error", zap.Strings("keys", keys), zap.Error(err))
		return err
	}

	r.logger.Debug("Cache deleted", zap.Strings("keys", keys))
	return nil
}

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
			r.logger.Error("Cache delete pattern error", zap.String("pattern", pattern), zap.Int("count", len(keys)), zap.Error(err))
			return err
		}
		r.logger.Debug("Cache pattern deleted", zap.String("pattern", pattern), zap.Int("count", len(keys)))
	} else {
		r.logger.Debug("Cache pattern not found", zap.String("pattern", pattern))
	}

	return nil
}

