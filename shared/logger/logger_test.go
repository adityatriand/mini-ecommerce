package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestNewConfig(t *testing.T) {
	config := NewConfig()
	
	assert.NotNil(t, config)
	assert.Equal(t, "info", config.Level)
	assert.Equal(t, "json", config.Format)
	assert.Equal(t, "stdout", config.OutputPath)
}

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		serviceName string
		expectError bool
	}{
		{
			name: "valid config",
			config: Config{
				Level:      "info",
				Format:     "json",
				OutputPath: "stdout",
			},
			serviceName: "test-service",
			expectError: false,
		},
		{
			name: "invalid level",
			config: Config{
				Level:      "invalid",
				Format:     "json",
				OutputPath: "stdout",
			},
			serviceName: "test-service",
			expectError: false, // Invalid level defaults to info
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config, tt.serviceName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, logger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, logger)
			}
		})
	}
}

func TestLogger_Methods(t *testing.T) {
	config := NewConfig()
	logger, err := NewLogger(config, "test-service")
	require.NoError(t, err)
	require.NotNil(t, logger)

	// Test Info
	logger.Info("test info message", zap.String("key", "value"))

	// Test Error
	logger.Error("test error message", zap.String("error", "test error"))

	// Test GetZapLogger
	zapLogger := logger.GetZapLogger()
	assert.NotNil(t, zapLogger)
	assert.IsType(t, &zap.Logger{}, zapLogger)

	// Test Sync
	err = logger.Sync()
	assert.NoError(t, err)
}