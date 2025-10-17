.PHONY: help dev-up dev-down dev-logs dev-ps test-all test-user test-product test-order test-gateway build-all build-user build-product build-order build-gateway clean migrate-user migrate-product migrate-order migrate-all config-generate config-local config-dev config-staging config-prod

# Default target
help:
	@echo "Available commands:"
	@echo "  dev-up          - Start all microservices in development mode"
	@echo "  dev-down        - Stop all microservices"
	@echo "  dev-logs        - View logs from all services"
	@echo "  dev-ps          - Show running containers"
	@echo "  test-all        - Run tests for all services"
	@echo "  test-user       - Run tests for user service"
	@echo "  test-product    - Run tests for product service"
	@echo "  test-order      - Run tests for order service"
	@echo "  test-gateway    - Run tests for gateway service"
	@echo "  build-all       - Build all services"
	@echo "  build-user      - Build user service"
	@echo "  build-product   - Build product service"
	@echo "  build-order     - Build order service"
	@echo "  build-gateway   - Build gateway service"
	@echo "  migrate-all     - Run migrations for all services"
	@echo "  migrate-user    - Run migrations for user service"
	@echo "  migrate-product - Run migrations for product service"
	@echo "  migrate-order   - Run migrations for order service"
	@echo "  config-generate - Generate configuration files for current environment"
	@echo "  config-local    - Generate configuration files for local environment"
	@echo "  config-dev      - Generate configuration files for development environment"
	@echo "  config-staging  - Generate configuration files for staging environment"
	@echo "  config-prod     - Generate configuration files for production environment"
	@echo "  clean           - Clean up containers and volumes"

# Development commands
dev-up:
	@echo "Starting microservices in development mode..."
	@docker-compose -f deployments/docker/docker-compose.microservices.yml up -d
	@echo "Services started! Access points:"
	@echo "  - API Gateway: http://localhost:8000"
	@echo "  - User Service: http://localhost:8001"
	@echo "  - Product Service: http://localhost:8002"
	@echo "  - Order Service: http://localhost:8003"
	@echo "  - Prometheus: http://localhost:9090"
	@echo "  - Grafana: http://localhost:3000"
	@echo "  - Loki: http://localhost:3100"

dev-down:
	@echo "Stopping all microservices..."
	@docker-compose -f deployments/docker/docker-compose.microservices.yml down

dev-logs:
	@docker-compose -f deployments/docker/docker-compose.microservices.yml logs -f

dev-ps:
	@docker-compose -f deployments/docker/docker-compose.microservices.yml ps

# Testing commands
test-all:
	@echo "Running tests for all services..."
	@go test ./services/user/... -v
	@go test ./services/product/... -v
	@go test ./services/order/... -v
	@go test ./services/gateway/... -v
	@go test ./shared/... -v

test-user:
	@echo "Running tests for user service..."
	@go test ./services/user/... -v

test-product:
	@echo "Running tests for product service..."
	@go test ./services/product/... -v

test-order:
	@echo "Running tests for order service..."
	@go test ./services/order/... -v

test-gateway:
	@echo "Running tests for gateway service..."
	@go test ./services/gateway/... -v

# Build commands
build-all: build-user build-product build-order build-gateway

build-user:
	@echo "Building user service..."
	@docker build -f services/user/Dockerfile -t mini-ecommerce-user:latest .

build-product:
	@echo "Building product service..."
	@docker build -f services/product/Dockerfile -t mini-ecommerce-product:latest .

build-order:
	@echo "Building order service..."
	@docker build -f services/order/Dockerfile -t mini-ecommerce-order:latest .

build-gateway:
	@echo "Building gateway service..."
	@docker build -f services/gateway/Dockerfile -t mini-ecommerce-gateway:latest .

# Cleanup commands
clean:
	@echo "Cleaning up containers and volumes..."
	@docker-compose -f deployments/docker/docker-compose.microservices.yml down -v
	@docker system prune -f

# Service-specific development commands
dev-up-user:
	@echo "Starting user service only..."
	@docker-compose -f deployments/docker/docker-compose.microservices.yml up -d postgres redis user

dev-up-product:
	@echo "Starting product service only..."
	@docker-compose -f deployments/docker/docker-compose.microservices.yml up -d postgres redis product

dev-up-order:
	@echo "Starting order service only..."
	@docker-compose -f deployments/docker/docker-compose.microservices.yml up -d postgres redis product order

dev-up-gateway:
	@echo "Starting gateway service only..."
	@docker-compose -f deployments/docker/docker-compose.microservices.yml up -d postgres redis user product order gateway

# Health check commands
health-check:
	@echo "Checking service health..."
	@curl -s http://localhost:8000/health | jq .
	@curl -s http://localhost:8001/health | jq .
	@curl -s http://localhost:8002/health | jq .
	@curl -s http://localhost:8003/health | jq .

# Database migration commands
migrate-all: migrate-user migrate-product migrate-order

migrate-user:
	@echo "Running migrations for user service..."
	@migrate -path services/user/migrations -database "${USER_DATABASE_URL}" up

migrate-product:
	@echo "Running migrations for product service..."
	@migrate -path services/product/migrations -database "${PRODUCT_DATABASE_URL}" up

migrate-order:
	@echo "Running migrations for order service..."
	@migrate -path services/order/migrations -database "${ORDER_DATABASE_URL}" up

migrate-down-user:
	@echo "Rolling back last migration for user service..."
	@migrate -path services/user/migrations -database "${USER_DATABASE_URL}" down 1

migrate-down-product:
	@echo "Rolling back last migration for product service..."
	@migrate -path services/product/migrations -database "${PRODUCT_DATABASE_URL}" down 1

migrate-down-order:
	@echo "Rolling back last migration for order service..."
	@migrate -path services/order/migrations -database "${ORDER_DATABASE_URL}" down 1

# Monitoring commands
monitor-up:
	@echo "Starting monitoring stack..."
	@docker-compose -f deployments/docker/docker-compose.microservices.yml up -d prometheus grafana loki

monitor-down:
	@echo "Stopping monitoring stack..."
	@docker-compose -f deployments/docker/docker-compose.microservices.yml stop prometheus grafana loki

# Logs commands
logs-user:
	@docker-compose -f deployments/docker/docker-compose.microservices.yml logs -f user

logs-product:
	@docker-compose -f deployments/docker/docker-compose.microservices.yml logs -f product

logs-order:
	@docker-compose -f deployments/docker/docker-compose.microservices.yml logs -f order

logs-gateway:
	@docker-compose -f deployments/docker/docker-compose.microservices.yml logs -f gateway

# Development setup
setup-dev:
	@echo "Setting up development environment..."
	@cp .env.example .env
	@echo "Please update .env with your configuration"
	@echo "Installing dependencies..."
	@go mod download
	@echo "Development environment ready!"

# Production commands
prod-up:
	@echo "Starting microservices in production mode..."
	@docker-compose -f deployments/docker/docker-compose.microservices.yml -f deployments/docker/docker-compose.prod.yml up -d

prod-down:
	@echo "Stopping production services..."
	@docker-compose -f deployments/docker/docker-compose.microservices.yml -f deployments/docker/docker-compose.prod.yml down

# Configuration generation commands
config-generate:
	@echo "Generating configuration files for current environment..."
	@./scripts/generate-configs.sh

config-local:
	@echo "Generating configuration files for local environment..."
	@ENVIRONMENT=local ./scripts/generate-configs.sh

config-dev:
	@echo "Generating configuration files for development environment..."
	@ENVIRONMENT=development ./scripts/generate-configs.sh

config-staging:
	@echo "Generating configuration files for staging environment..."
	@ENVIRONMENT=staging ./scripts/generate-configs.sh

config-prod:
	@echo "Generating configuration files for production environment..."
	@ENVIRONMENT=production ./scripts/generate-configs.sh
