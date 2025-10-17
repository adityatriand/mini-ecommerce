# Microservices Architecture

This directory contains the microservices implementation of the mini-ecommerce application.

## Architecture Overview

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   API Gateway   │    │  User Service   │    │ Product Service │
│                 │    │                 │    │                 │
│ - Routing       │    │ - Authentication│    │ - Product CRUD  │
│ - Auth Proxy    │    │ - User Mgmt     │    │ - Inventory     │
│ - Rate Limiting │    │ - JWT/Sessions  │    │ - Caching       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌─────────────────┐
                    │  Order Service  │
                    │                 │
                    │ - Order CRUD    │
                    │ - Payment       │
                    │ - Fulfillment   │
                    └─────────────────┘
```

## Services

### 1. API Gateway (`services/gateway/`)
- **Purpose**: Single entry point for all client requests
- **Responsibilities**:
  - Request routing to appropriate services
  - Authentication/authorization proxy
  - Rate limiting and throttling
  - Request/response transformation
  - Circuit breaker pattern
  - Load balancing

### 2. User Service (`services/user/`)
- **Purpose**: User management and authentication
- **Responsibilities**:
  - User registration/login
  - JWT token management
  - Session management
  - User profile management
  - Password management
- **Database**: `users` table
- **Port**: 8001

### 3. Product Service (`services/product/`)
- **Purpose**: Product catalog and inventory management
- **Responsibilities**:
  - Product CRUD operations
  - Inventory management
  - Product search and filtering
  - Caching (Redis)
- **Database**: `products` table
- **Port**: 8002

### 4. Order Service (`services/order/`)
- **Purpose**: Order processing and management
- **Responsibilities**:
  - Order creation and management
  - Order status tracking
  - Integration with Product Service for inventory
  - Order history
- **Database**: `orders`, `order_items` tables
- **Port**: 8003

## Inter-Service Communication

### Synchronous Communication
- **HTTP REST APIs** for most operations
- **gRPC** for high-performance internal communication (optional)

### Asynchronous Communication
- **Redis Pub/Sub** for event-driven communication
- **Message Queues** (future: RabbitMQ/Kafka)

## Shared Components

### Common Libraries (`pkg/`)
- **Logger**: Structured logging with Zap
- **Database**: Database connection and migration utilities
- **Cache**: Redis client wrapper
- **Metrics**: Prometheus metrics collection
- **Config**: Configuration management with Viper
- **Auth**: JWT and session management utilities
- **Response**: Standardized API response formats
- **Middleware**: Common HTTP middleware

## Data Management

### Database per Service
- Each service owns its data
- No direct database access between services
- Data consistency through events and transactions

### Shared Databases
- **Redis**: Shared cache and session storage
- **PostgreSQL**: Separate schemas per service

## Deployment

### Docker Compose
- Each service runs in its own container
- Shared infrastructure (Redis, PostgreSQL, monitoring)
- Service discovery through Docker networking

### Kubernetes (Future)
- Horizontal scaling per service
- Service mesh (Istio) for advanced networking
- ConfigMaps and Secrets for configuration

## Monitoring & Observability

### Distributed Tracing
- **Jaeger** for request tracing across services
- **OpenTelemetry** for instrumentation

### Metrics
- **Prometheus** for metrics collection
- **Grafana** for visualization
- Service-specific dashboards

### Logging
- **Loki** for centralized logging
- **Fluentd** for log aggregation
- Structured logging with correlation IDs

## Development Workflow

### Local Development
```bash
# Start all services
make dev-up

# Start specific service
make dev-up-gateway
make dev-up-user
make dev-up-product
make dev-up-order

# Run tests for all services
make test-all

# Run tests for specific service
make test-user
make test-product
make test-order
```

### Service Communication Testing
```bash
# Test service health
curl http://localhost:8000/health
curl http://localhost:8001/health
curl http://localhost:8002/health
curl http://localhost:8003/health

# Test through API Gateway
curl http://localhost:8000/api/v1/users/register
curl http://localhost:8000/api/v1/products
curl http://localhost:8000/api/v1/orders
```

## Migration Strategy

### Phase 1: Extract Services
1. Create service directories
2. Move domain logic to respective services
3. Implement shared libraries
4. Setup basic inter-service communication

### Phase 2: API Gateway
1. Implement API Gateway
2. Route requests to appropriate services
3. Implement authentication proxy

### Phase 3: Database Separation
1. Separate databases per service
2. Implement data synchronization
3. Handle distributed transactions

### Phase 4: Advanced Features
1. Implement event-driven architecture
2. Add circuit breakers and retries
3. Implement distributed tracing
4. Add advanced monitoring

## Benefits

### Scalability
- Scale services independently based on demand
- Optimize resource usage per service

### Technology Diversity
- Use different technologies per service if needed
- Independent deployment cycles

### Team Autonomy
- Teams can work independently on services
- Reduced coordination overhead

### Fault Isolation
- Service failures don't affect entire system
- Better error handling and recovery

## Challenges & Solutions

### Data Consistency
- **Challenge**: Maintaining data consistency across services
- **Solution**: Event-driven architecture with eventual consistency

### Network Latency
- **Challenge**: Increased network calls between services
- **Solution**: Caching, connection pooling, and async communication

### Complexity
- **Challenge**: Increased operational complexity
- **Solution**: Comprehensive monitoring, logging, and automation

### Testing
- **Challenge**: Testing distributed systems
- **Solution**: Contract testing, integration tests, and test containers

