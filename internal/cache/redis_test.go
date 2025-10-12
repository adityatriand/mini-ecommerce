package cache

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type TestData struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func setupTestRedis(t *testing.T) (*RedisCache, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	logger := zap.NewNop()
	cache := NewRedisCache(client, logger)
	return cache, mr
}

func TestNewRedisCache(t *testing.T) {
	t.Run("should create redis cache successfully", func(t *testing.T) {
		mr := miniredis.RunT(t)
		defer mr.Close()

		client := redis.NewClient(&redis.Options{
			Addr: mr.Addr(),
		})
		logger := zap.NewNop()

		cache := NewRedisCache(client, logger)

		assert.NotNil(t, cache)
		assert.NotNil(t, cache.client)
		assert.NotNil(t, cache.logger)
	})
}

func TestRedisCache_Get(t *testing.T) {
	ctx := context.Background()

	t.Run("should get value successfully", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		testData := TestData{ID: 1, Name: "Test"}
		data, _ := json.Marshal(testData)
		mr.Set("test-key", string(data))

		var result TestData
		err := cache.Get(ctx, "test-key", &result)

		assert.NoError(t, err)
		assert.Equal(t, testData.ID, result.ID)
		assert.Equal(t, testData.Name, result.Name)
	})

	t.Run("should return error when key does not exist", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		var result TestData
		err := cache.Get(ctx, "non-existent-key", &result)

		assert.Error(t, err)
		assert.Equal(t, redis.Nil, err)
	})

	t.Run("should return error when JSON unmarshal fails", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		mr.Set("invalid-json", "not a valid json")

		var result TestData
		err := cache.Get(ctx, "invalid-json", &result)

		assert.Error(t, err)
	})

	t.Run("should get complex nested data", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		type ComplexData struct {
			ID    int                    `json:"id"`
			Items []string               `json:"items"`
			Meta  map[string]interface{} `json:"meta"`
		}

		testData := ComplexData{
			ID:    42,
			Items: []string{"item1", "item2", "item3"},
			Meta: map[string]interface{}{
				"key1": "value1",
				"key2": 123,
			},
		}
		data, _ := json.Marshal(testData)
		mr.Set("complex-key", string(data))

		var result ComplexData
		err := cache.Get(ctx, "complex-key", &result)

		assert.NoError(t, err)
		assert.Equal(t, testData.ID, result.ID)
		assert.Equal(t, testData.Items, result.Items)
		assert.Equal(t, "value1", result.Meta["key1"])
	})
}

func TestRedisCache_Set(t *testing.T) {
	ctx := context.Background()

	t.Run("should set value successfully", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		testData := TestData{ID: 1, Name: "Test"}
		err := cache.Set(ctx, "test-key", testData, time.Minute)

		assert.NoError(t, err)
		assert.True(t, mr.Exists("test-key"))

		val, _ := mr.Get("test-key")
		var result TestData
		json.Unmarshal([]byte(val), &result)
		assert.Equal(t, testData.ID, result.ID)
		assert.Equal(t, testData.Name, result.Name)
	})

	t.Run("should set value with TTL", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		testData := TestData{ID: 2, Name: "TTL Test"}
		ttl := 5 * time.Second
		err := cache.Set(ctx, "ttl-key", testData, ttl)

		assert.NoError(t, err)
		assert.True(t, mr.Exists("ttl-key"))

		actualTTL := mr.TTL("ttl-key")
		assert.Greater(t, actualTTL, time.Duration(0))
		assert.LessOrEqual(t, actualTTL, ttl)
	})

	t.Run("should set with zero TTL (no expiration)", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		testData := TestData{ID: 3, Name: "No TTL"}
		err := cache.Set(ctx, "no-ttl-key", testData, 0)

		assert.NoError(t, err)
		assert.True(t, mr.Exists("no-ttl-key"))

		actualTTL := mr.TTL("no-ttl-key")
		// TTL 0 means no expiration in miniredis
		assert.Equal(t, time.Duration(0), actualTTL)
	})

	t.Run("should overwrite existing key", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		oldData := TestData{ID: 1, Name: "Old"}
		cache.Set(ctx, "overwrite-key", oldData, time.Minute)

		newData := TestData{ID: 2, Name: "New"}
		err := cache.Set(ctx, "overwrite-key", newData, time.Minute)

		assert.NoError(t, err)

		val, _ := mr.Get("overwrite-key")
		var result TestData
		json.Unmarshal([]byte(val), &result)
		assert.Equal(t, newData.ID, result.ID)
		assert.Equal(t, newData.Name, result.Name)
	})

	t.Run("should handle marshal error for unmarshalable data", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		// Channels cannot be marshaled to JSON
		unmarshalableData := make(chan int)
		err := cache.Set(ctx, "unmarshalable", unmarshalableData, time.Minute)

		assert.Error(t, err)
		assert.False(t, mr.Exists("unmarshalable"))
	})
}

func TestRedisCache_Delete(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete single key successfully", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		mr.Set("delete-key", "value")
		require.True(t, mr.Exists("delete-key"))

		err := cache.Delete(ctx, "delete-key")

		assert.NoError(t, err)
		assert.False(t, mr.Exists("delete-key"))
	})

	t.Run("should delete multiple keys successfully", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		mr.Set("key1", "value1")
		mr.Set("key2", "value2")
		mr.Set("key3", "value3")
		require.True(t, mr.Exists("key1"))
		require.True(t, mr.Exists("key2"))
		require.True(t, mr.Exists("key3"))

		err := cache.Delete(ctx, "key1", "key2", "key3")

		assert.NoError(t, err)
		assert.False(t, mr.Exists("key1"))
		assert.False(t, mr.Exists("key2"))
		assert.False(t, mr.Exists("key3"))
	})

	t.Run("should not error when deleting non-existent key", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		err := cache.Delete(ctx, "non-existent-key")

		assert.NoError(t, err)
	})

	t.Run("should delete some keys when only some exist", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		mr.Set("exists", "value")
		require.True(t, mr.Exists("exists"))

		err := cache.Delete(ctx, "exists", "does-not-exist")

		assert.NoError(t, err)
		assert.False(t, mr.Exists("exists"))
		assert.False(t, mr.Exists("does-not-exist"))
	})
}

func TestRedisCache_DeletePattern(t *testing.T) {
	ctx := context.Background()

	t.Run("should delete keys matching pattern", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		mr.Set("user:1:profile", "data1")
		mr.Set("user:2:profile", "data2")
		mr.Set("user:3:profile", "data3")
		mr.Set("product:1", "product1")

		err := cache.DeletePattern(ctx, "user:*:profile")

		assert.NoError(t, err)
		assert.False(t, mr.Exists("user:1:profile"))
		assert.False(t, mr.Exists("user:2:profile"))
		assert.False(t, mr.Exists("user:3:profile"))
		assert.True(t, mr.Exists("product:1")) // Should not be deleted
	})

	t.Run("should handle pattern with no matches", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		mr.Set("key1", "value1")
		mr.Set("key2", "value2")

		err := cache.DeletePattern(ctx, "non-matching-pattern:*")

		assert.NoError(t, err)
		assert.True(t, mr.Exists("key1"))
		assert.True(t, mr.Exists("key2"))
	})

	t.Run("should delete all keys matching wildcard", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		mr.Set("cache:user:1", "data1")
		mr.Set("cache:user:2", "data2")
		mr.Set("cache:product:1", "data3")
		mr.Set("other:key", "data4")

		err := cache.DeletePattern(ctx, "cache:*")

		assert.NoError(t, err)
		assert.False(t, mr.Exists("cache:user:1"))
		assert.False(t, mr.Exists("cache:user:2"))
		assert.False(t, mr.Exists("cache:product:1"))
		assert.True(t, mr.Exists("other:key"))
	})

	t.Run("should handle empty pattern gracefully", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		mr.Set("key1", "value1")

		err := cache.DeletePattern(ctx, "")

		assert.NoError(t, err)
		// Empty pattern in Redis SCAN matches all keys, so key1 will be deleted
		assert.False(t, mr.Exists("key1"))
	})
}

func TestRedisCache_IntegrationScenario(t *testing.T) {
	ctx := context.Background()

	t.Run("should handle complete cache lifecycle", func(t *testing.T) {
		cache, mr := setupTestRedis(t)
		defer mr.Close()

		// Set initial data
		user1 := TestData{ID: 1, Name: "Alice"}
		user2 := TestData{ID: 2, Name: "Bob"}
		cache.Set(ctx, "user:1", user1, time.Minute)
		cache.Set(ctx, "user:2", user2, time.Minute)

		// Get data
		var retrievedUser1 TestData
		err := cache.Get(ctx, "user:1", &retrievedUser1)
		assert.NoError(t, err)
		assert.Equal(t, user1.Name, retrievedUser1.Name)

		// Update data
		user1.Name = "Alice Updated"
		cache.Set(ctx, "user:1", user1, time.Minute)

		var updatedUser1 TestData
		cache.Get(ctx, "user:1", &updatedUser1)
		assert.Equal(t, "Alice Updated", updatedUser1.Name)

		// Delete one key
		cache.Delete(ctx, "user:2")
		var deletedUser TestData
		err = cache.Get(ctx, "user:2", &deletedUser)
		assert.Error(t, err)

		// Delete by pattern
		cache.Set(ctx, "user:3", TestData{ID: 3, Name: "Charlie"}, time.Minute)
		cache.DeletePattern(ctx, "user:*")
		assert.False(t, mr.Exists("user:1"))
		assert.False(t, mr.Exists("user:3"))
	})
}
