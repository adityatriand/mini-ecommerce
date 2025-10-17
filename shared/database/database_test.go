package database

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
)

// TestModel represents a simple model for testing
type TestModel struct {
	ID        uint      `gorm:"primaryKey"`
	Name      string    `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func TestConnect(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "valid config",
			config: Config{
				URL:             "sqlite://:memory:",
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: time.Hour,
			},
			expectError: false,
		},
		{
			name: "invalid URL",
			config: Config{
				URL: "invalid://url",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			db, err := Connect(tt.config, logger)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, db)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, db)
			}
		})
	}
}

func TestMigrate(t *testing.T) {
	config := Config{
		URL:             "sqlite://:memory:",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}

	logger := zaptest.NewLogger(t)
	db, err := Connect(config, logger)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Test migration
	models := []interface{}{&TestModel{}}
	err = Migrate(db, models, logger)
	assert.NoError(t, err)

	// Verify table was created
	var count int64
	err = db.Model(&TestModel{}).Count(&count).Error
	assert.NoError(t, err)
	assert.Equal(t, int64(0), count)
}

func TestHealthCheck(t *testing.T) {
	config := Config{
		URL:             "sqlite://:memory:",
		MaxOpenConns:    10,
		MaxIdleConns:    5,
		ConnMaxLifetime: time.Hour,
	}

	logger := zaptest.NewLogger(t)
	db, err := Connect(config, logger)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Test health check
	err = HealthCheck(db)
	assert.NoError(t, err)
}