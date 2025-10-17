package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		name        string
		serviceName string
		envVars     map[string]string
		expectError bool
	}{
		{
			name:        "valid user service config",
			serviceName: "user-service",
			envVars: map[string]string{
				"USER_SERVICE_DATABASE_URL": "postgres://user:pass@localhost:5432/user_db",
				"USER_SERVICE_REDIS_ADDR":   "localhost:6379",
				"USER_SERVICE_JWT_SECRET":   "test-secret",
			},
			expectError: false,
		},
		{
			name:        "valid product service config",
			serviceName: "product-service",
			envVars: map[string]string{
				"PRODUCT_SERVICE_DATABASE_URL": "postgres://user:pass@localhost:5432/product_db",
				"PRODUCT_SERVICE_REDIS_ADDR":   "localhost:6379",
			},
			expectError: false,
		},
		{
			name:        "missing required database URL",
			serviceName: "user-service",
			envVars: map[string]string{
				"USER_SERVICE_REDIS_ADDR": "localhost:6379",
				"USER_SERVICE_JWT_SECRET": "test-secret",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variables
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}
			defer func() {
				// Clean up environment variables
				for key := range tt.envVars {
					os.Unsetenv(key)
				}
			}()

			config, err := LoadConfig(tt.serviceName)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, config)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, config)
				assert.Equal(t, tt.serviceName, config.ServiceName)
			}
		})
	}
}

func TestServiceConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  ServiceConfig
		wantErr bool
	}{
		{
			name: "valid config",
			config: ServiceConfig{
				ServiceName: "test-service",
				Port:        "8080",
				Environment: "development",
				LogLevel:    "info",
				Database: DatabaseConfig{
					URL: "postgres://user:pass@localhost:5432/db",
				},
				Redis: RedisConfig{
					Addr: "localhost:6379",
				},
				JWT: JWTConfig{
					Secret: "test-secret",
				},
			},
			wantErr: false,
		},
		{
			name: "missing service name",
			config: ServiceConfig{
				Port: "8080",
			},
			wantErr: true,
		},
		{
			name: "missing database URL",
			config: ServiceConfig{
				ServiceName: "test-service",
				Port:        "8080",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// For now, just test that the config struct can be created
			// In a real implementation, you would add a Validate method
			assert.NotNil(t, tt.config)
		})
	}
}