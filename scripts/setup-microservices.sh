#!/bin/bash

# Setup script for Mini E-Commerce Microservices
# This script sets up the development environment for microservices

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Function to print colored output
print_header() {
    echo -e "${PURPLE}================================${NC}"
    echo -e "${PURPLE}$1${NC}"
    echo -e "${PURPLE}================================${NC}"
}

print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_step() {
    echo -e "${CYAN}[STEP]${NC} $1"
}

# Check if we're in the right directory
if [ ! -f "go.mod" ]; then
    print_error "go.mod not found. Please run this script from the project root"
    exit 1
fi

print_header "Mini E-Commerce Microservices Setup"
echo "This script will set up your development environment for microservices"
echo ""

# Check prerequisites
print_step "Checking prerequisites..."

# Check Go version
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go 1.24 or later"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
print_success "Go version: $GO_VERSION"

# Check Docker
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker"
    exit 1
fi

DOCKER_VERSION=$(docker --version | awk '{print $3}' | sed 's/,//')
print_success "Docker version: $DOCKER_VERSION"

# Check Docker Compose
if ! command -v docker-compose &> /dev/null; then
    print_error "Docker Compose is not installed. Please install Docker Compose"
    exit 1
fi

COMPOSE_VERSION=$(docker-compose --version | awk '{print $3}' | sed 's/,//')
print_success "Docker Compose version: $COMPOSE_VERSION"

# Check Make
if ! command -v make &> /dev/null; then
    print_error "Make is not installed. Please install Make"
    exit 1
fi

print_success "Make is available"

echo ""

# Setup Go modules
print_step "Setting up Go modules..."

print_status "Initializing Go modules..."
go mod tidy

print_status "Downloading dependencies..."
go mod download

print_success "Go modules setup complete"

# Setup git hooks
print_step "Setting up Git hooks..."

if [ -f ".git/hooks/pre-commit" ]; then
    print_warning "Pre-commit hook already exists. Backing up..."
    mv .git/hooks/pre-commit .git/hooks/pre-commit.backup
fi

print_status "Installing pre-commit hook..."
cp scripts/hooks/pre-commit.sh .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

print_success "Git hooks setup complete"

# Create necessary directories
print_step "Creating necessary directories..."

directories=(
    "deployments/configs"
    "deployments/monitoring/grafana/provisioning/datasources"
    "deployments/monitoring/grafana/provisioning/dashboards"
    "logs"
    "tmp"
)

for dir in "${directories[@]}"; do
    if [ ! -d "$dir" ]; then
        print_status "Creating directory: $dir"
        mkdir -p "$dir"
    fi
done

print_success "Directories created"

# Setup environment files
print_step "Setting up environment files..."

# Create .env.microservices if it doesn't exist
if [ ! -f ".env.microservices" ]; then
    print_status "Creating .env.microservices..."
    if [ -f ".env" ]; then
        cp .env .env.microservices
        print_success ".env.microservices created from existing .env"
    else
        cat > .env.microservices << EOF
# Mini E-Commerce Microservices Environment Variables

# Database Configuration
POSTGRES_USER=ecommerce
POSTGRES_PASSWORD=ecommerce_password
POSTGRES_DB=mini_ecommerce
POSTGRES_HOST=postgres
POSTGRES_PORT=5432

# Redis Configuration
REDIS_ADDR=redis:6379
REDIS_PASSWORD=

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production
JWT_EXP_MINUTES=15
JWT_REFRESH_EXP_HOURS=168

# Service Configuration
USER_SERVICE_URL=http://user-service:8001
PRODUCT_SERVICE_URL=http://product-service:8002
ORDER_SERVICE_URL=http://order-service:8003
GATEWAY_SERVICE_URL=http://api-gateway:8000

# Monitoring Configuration
PROMETHEUS_ENABLED=true
GRAFANA_ENABLED=true
LOKI_ENABLED=true

# Environment
ENVIRONMENT=development
LOG_LEVEL=info
EOF
        print_success ".env.microservices created"
    fi
else
    print_warning ".env.microservices already exists"
fi

# Create service-specific configs if they don't exist
services=("gateway" "user" "product" "order")
for service in "${services[@]}"; do
    config_file="services/$service/configs/config.yaml"
    if [ ! -f "$config_file" ]; then
        print_status "Creating config for $service service..."
        mkdir -p "services/$service/configs"
        
        # Set default port based on service
        case $service in
            "gateway") port="8000" ;;
            "user") port="8001" ;;
            "product") port="8002" ;;
            "order") port="8003" ;;
        esac
        
        cat > "$config_file" << EOF
service_name: $service-service
port: "$port"
environment: development
log_level: info

database:
  url: postgresql://ecommerce:ecommerce_password@postgres:5432/mini_ecommerce?sslmode=disable
  max_open_conns: 25
  max_idle_conns: 5
  conn_max_lifetime: "5m"

redis:
  addr: redis:6379
  password: ""
  db: 0

jwt:
  secret: your-super-secret-jwt-key-change-this-in-production
  expiration: "15m"
  refresh_expiration: "168h"

services:
  user_service:
    host: user-service
    port: "8001"
    url: "http://user-service:8001"
  product_service:
    host: product-service
    port: "8002"
    url: "http://product-service:8002"
  order_service:
    host: order-service
    port: "8003"
    url: "http://order-service:8003"
  gateway_service:
    host: api-gateway
    port: "8000"
    url: "http://api-gateway:8000"

monitoring:
  prometheus:
    enabled: true
    port: "9090"
    path: "/metrics"
  grafana:
    enabled: true
    url: "http://grafana:3000"
    user: "admin"
    pass: "admin"
  loki:
    enabled: true
    url: "http://loki:3100"
EOF
        print_success "Config created for $service service"
    else
        print_warning "Config already exists for $service service"
    fi

    # Create service-specific .env file if it doesn't exist
    env_file="services/$service/.env"
    env_example="services/$service/env.example"
    if [ ! -f "$env_file" ] && [ -f "$env_example" ]; then
        print_status "Creating .env for $service service..."
        cp "$env_example" "$env_file"
        print_success ".env created for $service service"
    elif [ -f "$env_file" ]; then
        print_warning ".env already exists for $service service"
    fi
done

# Setup monitoring configuration
print_step "Setting up monitoring configuration..."

# Create Grafana datasource configuration
if [ ! -f "deployments/monitoring/grafana/provisioning/datasources/datasource.yml" ]; then
    print_status "Creating Grafana datasource configuration..."
    cat > deployments/monitoring/grafana/provisioning/datasources/datasource.yml << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
    editable: true

  - name: Loki
    type: loki
    access: proxy
    url: http://loki:3100
    editable: true
EOF
    print_success "Grafana datasource configuration created"
fi

# Create Grafana dashboard provisioning
if [ ! -f "deployments/monitoring/grafana/provisioning/dashboards/dashboard.yml" ]; then
    print_status "Creating Grafana dashboard provisioning..."
    cat > deployments/monitoring/grafana/provisioning/dashboards/dashboard.yml << EOF
apiVersion: 1

providers:
  - name: 'default'
    orgId: 1
    folder: ''
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    allowUiUpdates: true
    options:
      path: /var/lib/grafana/dashboards
EOF
    print_success "Grafana dashboard provisioning created"
fi

# Build services
print_step "Building microservices..."

print_status "Building all services..."
if make build-all; then
    print_success "All services built successfully"
else
    print_error "Failed to build services"
    exit 1
fi

# Run tests
print_step "Running tests..."

print_status "Running tests for all services..."
if make test-all; then
    print_success "All tests passed"
else
    print_warning "Some tests failed, but continuing with setup"
fi

# Setup Docker network
print_step "Setting up Docker network..."

print_status "Creating Docker network for microservices..."
if docker network ls | grep -q "ecommerce-network"; then
    print_warning "Docker network already exists"
else
    docker network create ecommerce-network
    print_success "Docker network created"
fi

# Generate secure configuration files
print_step "Generating secure configuration files..."

if [ -f "scripts/generate-configs.sh" ]; then
    print_status "Running configuration generation script..."
    ./scripts/generate-configs.sh
    print_success "Generated secure configuration files"
else
    print_warning "Configuration generation script not found, skipping"
fi

# Final instructions
print_header "Setup Complete! 🎉"
echo ""
print_success "Your microservices development environment is ready!"
echo ""
echo "Next steps:"
echo "1. Start the microservices stack:"
echo "   ${CYAN}make dev-up${NC}"
echo ""
echo "2. Check service health:"
echo "   ${CYAN}make health-check${NC}"
echo ""
echo "3. View logs:"
echo "   ${CYAN}make dev-logs${NC}"
echo ""
echo "4. Access the services:"
echo "   ${CYAN}API Gateway: http://localhost:8000${NC}"
echo "   ${CYAN}Prometheus: http://localhost:9090${NC}"
echo "   ${CYAN}Grafana: http://localhost:3000 (admin/admin)${NC}"
echo "   ${CYAN}Loki: http://localhost:3100${NC}"
echo ""
echo "5. Test the API:"
echo "   ${CYAN}curl -X POST http://localhost:8000/api/v1/users/register \\${NC}"
echo "   ${CYAN}  -H \"Content-Type: application/json\" \\${NC}"
echo "   ${CYAN}  -d '{\"email\":\"test@example.com\",\"password\":\"password123\",\"first_name\":\"Test\",\"last_name\":\"User\"}'${NC}"
echo ""
print_success "Happy coding! 🚀"
