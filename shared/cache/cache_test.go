package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

func TestNewRedisCache(t *testing.T) {
	// Start a test Redis server
	s := miniredis.RunT(t)
	defer s.Close()

	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "valid config",
			config: Config{
				Addr:     s.Addr(),
				Password: "",
				DB:       0,
			},
			expectError: false,
		},
		{
			name: "invalid address",
			config: Config{
				Addr:     "invalid:6379",
				Password: "",
				DB:       0,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			cache, err := NewRedisCache(tt.config, logger)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, cache)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, cache)
			}
		})
	}
}

func TestRedisCache_Operations(t *testing.T) {
	// Start a test Redis server
	s := miniredis.RunT(t)
	defer s.Close()

	config := Config{
		Addr:     s.Addr(),
		Password: "",
		DB:       0,
	}

	logger := zaptest.NewLogger(t)
	cache, err := NewRedisCache(config, logger)
	require.NoError(t, err)
	require.NotNil(t, cache)

	ctx := context.Background()

	t.Run("Set and Get", func(t *testing.T) {
		key := "test-key"
		value := "test-value"
		ttl := time.Hour

		// Set value
		err := cache.Set(ctx, key, value, ttl)
		assert.NoError(t, err)

		// Get value
		var result string
		err = cache.Get(ctx, key, &result)
		assert.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("Delete", func(t *testing.T) {
		key := "delete-key"
		value := "delete-value"
		ttl := time.Hour

		// Set value
		err := cache.Set(ctx, key, value, ttl)
		assert.NoError(t, err)

		// Delete value
		err = cache.Delete(ctx, key)
		assert.NoError(t, err)

		// Verify it's gone
		var result string
		err = cache.Get(ctx, key, &result)
		assert.Error(t, err)
	})
}