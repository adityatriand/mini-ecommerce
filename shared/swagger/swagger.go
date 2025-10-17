package swagger

import "fmt"

// SwaggerInfo represents the swagger information for a service
type SwaggerInfo struct {
	Title       string
	Description string
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
}

// Validate validates the swagger info
func (s SwaggerInfo) Validate() error {
	if s.Title == "" {
		return fmt.Errorf("title is required")
	}
	if s.Version == "" {
		return fmt.Errorf("version is required")
	}
	if len(s.Schemes) == 0 {
		return fmt.Errorf("at least one scheme is required")
	}
	return nil
}

// GetGatewaySwaggerInfo returns swagger info for the gateway service
func GetGatewaySwaggerInfo() SwaggerInfo {
	return SwaggerInfo{
		Title:       "Mini E-Commerce API Gateway",
		Description: "API Gateway for Mini E-Commerce Microservices",
		Version:     "1.0",
		Host:        "localhost:8000",
		BasePath:    "/api/v1",
		Schemes:     []string{"http", "https"},
	}
}

// GetUserSwaggerInfo returns swagger info for the user service
func GetUserSwaggerInfo() SwaggerInfo {
	return SwaggerInfo{
		Title:       "User Service API",
		Description: "User management and authentication service",
		Version:     "1.0",
		Host:        "localhost:8001",
		BasePath:    "/api/v1",
		Schemes:     []string{"http", "https"},
	}
}

// GetProductSwaggerInfo returns swagger info for the product service
func GetProductSwaggerInfo() SwaggerInfo {
	return SwaggerInfo{
		Title:       "Product Service API",
		Description: "Product catalog and inventory management service",
		Version:     "1.0",
		Host:        "localhost:8002",
		BasePath:    "/api/v1",
		Schemes:     []string{"http", "https"},
	}
}

// GetOrderSwaggerInfo returns swagger info for the order service
func GetOrderSwaggerInfo() SwaggerInfo {
	return SwaggerInfo{
		Title:       "Order Service API",
		Description: "Order processing and management service",
		Version:     "1.0",
		Host:        "localhost:8003",
		BasePath:    "/api/v1",
		Schemes:     []string{"http", "https"},
	}
}
