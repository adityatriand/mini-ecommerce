# 🛒 Mini E-Commerce Microservices

A modern, scalable e-commerce platform built with Go microservices architecture, featuring user management, product catalog, order processing, and comprehensive monitoring.

## 🏗️ Architecture

### Microservices Overview
```
┌─────────────────────────────────────────────────────────────┐
│                    API Gateway (8000)                      │
│  ┌─────────────────────────────────────────────────────────┐│
│  │  Routing │ Auth │ Rate Limiting │ Circuit Breaker      ││
│  └─────────────────────────────────────────────────────────┘│
└─────────────────────────────────────────────────────────────┘
         │              │              │              │
    ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐
    │  User   │   │Product  │   │ Order   │   │Monitoring│
    │ Service │   │Service  │   │Service  │   │ Stack   │
    │  (8001) │   │ (8002)  │   │ (8003)  │   │         │
    └─────────┘   └─────────┘   └─────────┘   └─────────┘
         │              │              │              │
    ┌─────────┐   ┌─────────┘   ┌─────────┘   ┌─────────┐
    │   DB    │   │    DB      │    DB      │Prometheus│
    └─────────┘   └─────────┘   └─────────┘   └─────────┘
```

### Services

| Service | Port | Description | Database |
|---------|------|-------------|----------|
| **API Gateway** | 8000 | Request routing, authentication, load balancing | - |
| **User Service** | 8001 | User management, authentication, JWT/session handling | `users` |
| **Product Service** | 8002 | Product catalog, inventory management, caching | `products` |
| **Order Service** | 8003 | Order processing, status management, inter-service communication | `orders`, `order_items` |

### Technology Stack

- **Language**: Go 1.24
- **Framework**: Gin Web Framework
- **Database**: PostgreSQL 16
- **Cache**: Redis 7
- **Authentication**: JWT + Session-based
- **Monitoring**: Prometheus, Grafana, Loki
- **Containerization**: Docker & Docker Compose
- **Configuration**: Viper
- **Logging**: Zap (Structured Logging)

## 🚀 Quick Start

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Make

### 1. Clone and Setup

```bash
git clone <repository-url>
cd mini-ecommerce
chmod +x scripts/setup-microservices.sh
./scripts/setup-microservices.sh
```

### 2. Start Microservices

```bash
# Start all services
make dev-up

# Check health
make health-check
```

### 3. Test the API

```bash
# Register a user
curl -X POST http://localhost:8000/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","first_name":"Test","last_name":"User"}'

# Login
curl -X POST http://localhost:8000/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Create a product (requires auth)
curl -X POST http://localhost:8000/api/v1/products \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Product","price":1000,"stock":10,"category":"Electronics"}'

# Create an order (requires auth)
curl -X POST http://localhost:8000/api/v1/orders \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"items":[{"product_id":1,"quantity":2}]}'
```

## 🛠️ Available Commands

### Development
```bash
make dev-up          # Start all microservices in development mode
make dev-down        # Stop all microservices
make dev-logs        # View logs from all services
make dev-ps          # Show running containers
```

### Testing
```bash
make test-all        # Run tests for all services
make test-user       # Run tests for user service
make test-product    # Run tests for product service
make test-order      # Run tests for order service
make test-gateway    # Run tests for gateway service
```

### Building
```bash
make build-all       # Build all services
make build-user      # Build user service
make build-product   # Build product service
make build-order     # Build order service
make build-gateway   # Build gateway service
```

### Database Migrations
```bash
make migrate-all     # Run migrations for all services
make migrate-user    # Run migrations for user service
make migrate-product # Run migrations for product service
make migrate-order   # Run migrations for order service
```

### Configuration
```bash
make config-generate # Generate configuration files for current environment
make config-local    # Generate configuration files for local environment
make config-dev      # Generate configuration files for development environment
make config-staging  # Generate configuration files for staging environment
make config-prod     # Generate configuration files for production environment
```

### Utilities
```bash
make clean           # Clean up containers and volumes
make help            # Show all available commands
```

## 📊 Monitoring & Observability

### Access Points

- **API Gateway**: http://localhost:8000
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin)
- **Loki**: http://localhost:3100

### Health Checks

- **Gateway**: http://localhost:8000/health
- **User Service**: http://localhost:8001/health
- **Product Service**: http://localhost:8002/health
- **Order Service**: http://localhost:8003/health

## 🛠️ Development

### Project Structure

```
mini-ecommerce/
├── services/                    # Microservices
│   ├── gateway/                # API Gateway
│   ├── user/                   # User Service
│   ├── product/                # Product Service
│   └── order/                  # Order Service
├── shared/                     # Shared Libraries
│   ├── config/                 # Configuration management
│   ├── logger/                 # Logging utilities
│   ├── database/               # Database utilities
│   ├── cache/                  # Cache utilities
│   ├── response/               # Response helpers
│   └── middleware/             # Common middleware
├── deployments/                # Deployment configs
│   ├── docker/
│   └── monitoring/
├── scripts/                    # Setup and utility scripts
└── Makefile      # Microservices commands
```

### Development Commands

```bash
# Service Management
make dev-up          # Start all services
make dev-down        # Stop all services
make dev-logs        # View logs
make health-check    # Check health

# Individual Services
make dev-up-user     # Start user service
make dev-up-product  # Start product service
make dev-up-order    # Start order service
make dev-up-gateway  # Start gateway

# Building
make build-all       # Build all services
make build-user      # Build user service

# Testing
make test-all        # Test all services
make test-user       # Test user service

# Logs
make logs-user       # View user service logs
make logs-product    # View product service logs
```

### Configuration

**🔒 Secure Template-Based Configuration**

The project uses a secure template-based configuration system to protect sensitive data:

#### Configuration Strategy
- **Templates only**: `config.yaml.template` files contain environment variable placeholders
- **No hardcoded configs**: No `config.yaml` files with sensitive data are committed to git
- **Environment-specific**: Generated configs are created from templates with actual values
- **Version control**: Only templates are committed, never actual configs with secrets

#### Generate Configuration Files
```bash
# Generate for specific environment
make config-local     # Local development
make config-dev       # Development environment
make config-staging   # Staging environment
make config-prod      # Production environment

# Or use environment variable
ENVIRONMENT=production ./scripts/generate-configs.sh
```

#### Environment Variables
Create `.env.microservices` file with your actual secrets:
```bash
# Database
DATABASE_URL=postgresql://user:secure_password@localhost:5432/db?sslmode=disable

# JWT
JWT_SECRET=your-super-secure-jwt-secret-key-here

# Redis
REDIS_PASSWORD=your-redis-password

# Environment
ENVIRONMENT=development
LOG_LEVEL=info
```

#### File Structure
```
services/
├── user/
│   └── configs/
│       ├── config.yaml.template          # Template (committed)
│       ├── config.local.yaml            # Generated (ignored)
│       ├── config.development.yaml      # Generated (ignored)
│       ├── config.staging.yaml          # Generated (ignored)
│       └── config.production.yaml       # Generated (ignored)
```

#### Security Features
- ✅ No sensitive data in version control
- ✅ Environment-specific configurations
- ✅ Automatic secret generation
- ✅ Environment variable overrides
- ✅ Clear error messages when configs are missing

## 🔐 Authentication

### Authentication Flow

1. **User Registration/Login** → User Service
2. **JWT Token Generation** → User Service
3. **Token Validation** → API Gateway
4. **Request Routing** → Backend Services

### Supported Authentication Methods

- **JWT Tokens** (Primary)
- **Session-based** (Fallback)
- **Refresh Tokens** (Automatic renewal)

## 📈 API Endpoints

### Public Endpoints
- `POST /api/v1/users/register` - User registration
- `POST /api/v1/users/login` - User login

### Protected Endpoints (Require Authentication)

#### User Management
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user
- `GET /api/v1/users` - Get all users

#### Product Management
- `GET /api/v1/products` - Get products
- `POST /api/v1/products` - Create product
- `PUT /api/v1/products/:id` - Update product
- `DELETE /api/v1/products/:id` - Delete product
- `PATCH /api/v1/products/:id/stock` - Update stock

#### Order Management
- `GET /api/v1/orders` - Get orders
- `POST /api/v1/orders` - Create order
- `PUT /api/v1/orders/:id` - Update order
- `DELETE /api/v1/orders/:id` - Delete order

## 🔄 Inter-Service Communication

### Communication Patterns

- **Synchronous HTTP** - REST APIs for client requests
- **Inter-service HTTP** - Service-to-service communication
- **Service Discovery** - Docker networking for service location

### Example: Order Creation Flow

```
Client → Gateway → Order Service → Product Service (stock validation)
                ↓
            Order Service → Product Service (stock update)
                ↓
            Order Service → Database (order creation)
```

## 🧪 Testing

### Running Tests

```bash
# All services
make test-all

# Individual services
make test-user
make test-product
make test-order
make test-gateway
```

### Test Coverage

- **Unit Tests** - Individual service components
- **Integration Tests** - Service-to-service communication
- **End-to-End Tests** - Complete user workflows

## 🚀 Deployment

### Docker Deployment

```bash
# Build and start all services
make build-all
make dev-up
```

### Production Considerations

- **Environment Variables** - Set production values
- **Database Security** - Use strong passwords and SSL
- **JWT Secrets** - Use cryptographically secure secrets
- **Monitoring** - Enable all monitoring components
- **Logging** - Configure appropriate log levels

## 🔧 Troubleshooting

### Common Issues

1. **Service Not Starting**
   ```bash
   make logs-user
   curl http://localhost:8001/health
   ```

2. **Database Connection Issues**
   ```bash
   docker-compose -f deployments/docker/docker-compose.microservices.yml ps postgres
   ```

3. **Authentication Issues**
   ```bash
   curl -X POST http://localhost:8000/api/v1/users/login \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com","password":"password123"}'
   ```

### Debug Commands

```bash
# View all container logs
make dev-logs

# Check service health
make health-check

# View container status
make dev-ps
```

## 🔮 Future Enhancements

### Planned Features

1. **Event-Driven Architecture**
   - Message queues (RabbitMQ/Kafka)
   - Event sourcing
   - CQRS pattern

2. **Advanced Monitoring**
   - Distributed tracing (Jaeger)
   - APM integration
   - Custom business metrics

3. **Security Enhancements**
   - OAuth2/OIDC integration
   - Rate limiting per user
   - API versioning

4. **Scalability**
   - Kubernetes deployment
   - Auto-scaling
   - Service mesh (Istio)

## 📚 Documentation

- **`MICROSERVICES_COMPLETE.md`** - Complete implementation guide
- **`MICROSERVICES_MIGRATION.md`** - Migration from monolithic
- **`services/README.md`** - Service architecture details
- **`Makefile`** - All available commands

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🆘 Support

For questions or issues, please open a GitHub issue or contact the development team.

---

**Built with ❤️ using Go microservices architecture**