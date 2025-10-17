package swagger

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetGatewaySwaggerInfo(t *testing.T) {
	info := GetGatewaySwaggerInfo()
	
	assert.Equal(t, "Mini E-Commerce API Gateway", info.Title)
	assert.Equal(t, "API Gateway for Mini E-Commerce Microservices", info.Description)
	assert.Equal(t, "1.0", info.Version)
	assert.Equal(t, "localhost:8000", info.Host)
	assert.Equal(t, "/api/v1", info.BasePath)
	assert.Equal(t, []string{"http", "https"}, info.Schemes)
}

func TestGetUserSwaggerInfo(t *testing.T) {
	info := GetUserSwaggerInfo()
	
	assert.Equal(t, "User Service API", info.Title)
	assert.Equal(t, "User management and authentication service", info.Description)
	assert.Equal(t, "1.0", info.Version)
	assert.Equal(t, "localhost:8001", info.Host)
	assert.Equal(t, "/api/v1", info.BasePath)
	assert.Equal(t, []string{"http", "https"}, info.Schemes)
}

func TestGetProductSwaggerInfo(t *testing.T) {
	info := GetProductSwaggerInfo()
	
	assert.Equal(t, "Product Service API", info.Title)
	assert.Equal(t, "Product catalog and inventory management service", info.Description)
	assert.Equal(t, "1.0", info.Version)
	assert.Equal(t, "localhost:8002", info.Host)
	assert.Equal(t, "/api/v1", info.BasePath)
	assert.Equal(t, []string{"http", "https"}, info.Schemes)
}

func TestGetOrderSwaggerInfo(t *testing.T) {
	info := GetOrderSwaggerInfo()
	
	assert.Equal(t, "Order Service API", info.Title)
	assert.Equal(t, "Order processing and management service", info.Description)
	assert.Equal(t, "1.0", info.Version)
	assert.Equal(t, "localhost:8003", info.Host)
	assert.Equal(t, "/api/v1", info.BasePath)
	assert.Equal(t, []string{"http", "https"}, info.Schemes)
}

func TestSwaggerInfo_Validate(t *testing.T) {
	tests := []struct {
		name    string
		info    SwaggerInfo
		wantErr bool
	}{
		{
			name: "valid swagger info",
			info: SwaggerInfo{
				Title:       "Test API",
				Description: "Test Description",
				Version:     "1.0",
				Host:        "localhost:8080",
				BasePath:    "/api/v1",
				Schemes:     []string{"http"},
			},
			wantErr: false,
		},
		{
			name: "missing title",
			info: SwaggerInfo{
				Description: "Test Description",
				Version:     "1.0",
				Host:        "localhost:8080",
				BasePath:    "/api/v1",
				Schemes:     []string{"http"},
			},
			wantErr: true,
		},
		{
			name: "missing version",
			info: SwaggerInfo{
				Title:       "Test API",
				Description: "Test Description",
				Host:        "localhost:8080",
				BasePath:    "/api/v1",
				Schemes:     []string{"http"},
			},
			wantErr: true,
		},
		{
			name: "empty schemes",
			info: SwaggerInfo{
				Title:       "Test API",
				Description: "Test Description",
				Version:     "1.0",
				Host:        "localhost:8080",
				BasePath:    "/api/v1",
				Schemes:     []string{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.info.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

