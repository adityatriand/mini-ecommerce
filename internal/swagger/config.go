package swagger

import (
	"mini-e-commerce/docs"
)

func SetupSwaggerInfo() {
	docs.SwaggerInfo.Title = "Mini E-Commerce API"
	docs.SwaggerInfo.Description = "This is a simple e-commerce API with product, order, and auth."
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.Host = "localhost:8080"
	docs.SwaggerInfo.BasePath = "/api"
	docs.SwaggerInfo.Schemes = []string{"http", "https"}
}
