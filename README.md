# Mini E-Commerce

A mini e-commerce backend application built with Go, Gin, PostgreSQL, and Redis with integrated monitoring using Prometheus and Grafana.

## Features

- 🔐 User authentication with JWT and session management
- 📦 Product management with caching
- 🛒 Order management
- 🚀 RESTful API
- 📚 Swagger documentation
- 🗄️ Database migrations
- ⚡ Redis caching for heavy endpoints
- 📊 Prometheus metrics collection
- 📈 Grafana dashboards for monitoring
- 🐳 Docker & Docker Compose support
- ✅ Health check endpoints
- 🧪 Comprehensive unit tests (187 test cases)

## Getting Started

### Quick Start with Docker (Recommended)

The fastest way to get started is using Docker Compose:

```bash
# Clone the repository
git clone <repository-url>
cd mini-e-commerce

# Copy environment file
cp .env.docker .env

# Start all services (app, postgres, redis, prometheus, grafana)
make docker-up
```

That's it! All services will be running with monitoring enabled.

**Access points:**
- API: http://localhost:8080
- Swagger: http://localhost:8080/swagger/index.html
- Prometheus: http://localhost:9090
- Grafana: http://localhost:3000 (admin/admin)

For detailed Docker setup instructions, see [DOCKER_SETUP.md](DOCKER_SETUP.md).

### Local Development Setup

#### Prerequisites

- Go 1.22 or higher
- PostgreSQL 14+
- Redis 7+
- Make
- Docker & Docker Compose (for containerized deployment)

#### Setup

Clone the repository and run setup:

```bash
git clone <repository-url>
cd mini-e-commerce
make setup
```

This will:
- Install git hooks for automated testing
- Download dependencies
- Install development tools

#### Configuration

**For local development** (without Docker), copy the example file:

```bash
cp .env.example .env
```

Then update `.env` with your local database and Redis credentials:
- `DATABASE_URL`: Point to your local PostgreSQL (e.g., `localhost:5432`)
- `REDIS_ADDR`: Point to your local Redis (e.g., `localhost:6379`)
- `JWT_SECRET`: Generate with `openssl rand -base64 32`

**Note:** Use `.env.docker` template if running with Docker Compose (see Quick Start section).

#### Database Migrations

Run migrations:

```bash
make migrate-up
```

Create a new migration:

```bash
make migrate-create name=your_migration_name
```

#### Running the Application Locally

```bash
go run cmd/main.go
```

The server will start on the port specified in your configuration (default: 8080).

## API Documentation

After starting the server, visit:

```
http://localhost:8080/swagger/index.html
```

## Health Checks

The application provides health check endpoints for monitoring and orchestration:

- `GET /health` - Simple health check
- `GET /health/ready` - Readiness probe (checks database & Redis connectivity)
- `GET /health/live` - Liveness probe

Example:
```bash
curl http://localhost:8080/health/ready
```

## Monitoring & Metrics

### Prometheus Metrics

Metrics are available at `/metrics` endpoint:
```bash
curl http://localhost:8080/metrics
```

**Available metrics:**
- `http_requests_total` - Total HTTP requests by method, path, and status
- `http_request_duration_seconds` - Request duration histogram
- `cache_hits_total` / `cache_misses_total` - Cache performance
- `db_queries_total` - Database query counts
- `orders_created_total` - Business metrics
- `auth_attempts_total` - Authentication metrics
- Plus 50+ Go runtime metrics

### Grafana Dashboards

When running with Docker, Grafana is pre-configured with dashboards:

1. Open http://localhost:3000
2. Login: `admin` / `admin`
3. View "Mini E-Commerce Dashboard"

The dashboard shows:
- Request rate and latency (p95, p99)
- HTTP status code distribution
- Cache hit/miss rates
- Memory and goroutine metrics
- Database performance

## Development

### Running Tests

Run all tests:
```bash
make test
```

Run specific package tests:
```bash
make test-auth      # Authentication tests
make test-product   # Product tests
make test-order     # Order tests
```

**Test Coverage:** 187 test cases across 6 packages:
- `internal/auth` - 25 tests
- `internal/product` - 44 tests
- `internal/order` - 58 tests
- `internal/utils` - 21 tests
- `internal/middleware` - 21 tests
- `internal/cache` - 18 tests

### Git Hooks

Pre-commit hooks are automatically installed during setup. They will:
- Run tests for packages you modified
- Prevent commits if tests fail
- Keep the codebase stable
- Work dynamically with all packages

### Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed contribution guidelines.

## Project Structure

```
.
├── cmd/                    # Application entrypoints
├── internal/
│   ├── auth/              # Authentication & user management
│   ├── product/           # Product management
│   ├── order/             # Order management
│   ├── cache/             # Redis caching layer
│   ├── config/            # Configuration management
│   ├── database/          # Database connection & migrations
│   ├── health/            # Health check handlers
│   ├── logger/            # Structured logging
│   ├── metrics/           # Prometheus metrics
│   ├── middleware/        # HTTP middleware (auth, logging, metrics)
│   ├── response/          # Response helpers
│   ├── swagger/           # Swagger setup
│   └── utils/             # Utility functions
├── routes/                # Route registration
├── migrations/            # Database migrations (SQL)
├── monitoring/            # Monitoring configurations
│   ├── prometheus/        # Prometheus config
│   └── grafana/           # Grafana dashboards & provisioning
├── scripts/               # Utility scripts and hooks
├── docs/                  # Swagger documentation
├── Dockerfile             # Application container
├── docker-compose.yml     # Multi-container orchestration
└── .dockerignore          # Docker build exclusions
```

## Available Commands

### Development Commands
```bash
make setup              # Setup development environment
make install-hooks      # Install git hooks
make test               # Run all tests
make test-auth          # Run auth package tests
make test-product       # Run product package tests
make test-order         # Run order package tests
make pre-commit         # Run pre-commit checks manually
```

### Database Commands
```bash
make migrate-up         # Run database migrations
make migrate-down       # Rollback last migration
make migrate-create     # Create new migration (requires name=xxx)
make migrate-force      # Force migration version (requires version=N)
make migrate-version    # Show current migration version
```

### Docker Commands
```bash
make docker-build       # Build Docker images
make docker-up          # Start all services (app, db, redis, monitoring)
make docker-down        # Stop all services
make docker-logs        # View logs from all services
make docker-ps          # Show running containers
make docker-clean       # Remove all containers and volumes (with confirmation)
```

**Docker services include:**
- Application (Go API)
- PostgreSQL database
- Redis cache
- Prometheus (metrics)
- Grafana (dashboards)

## Technology Stack

**Backend:**
- Go 1.22
- Gin Web Framework
- GORM (ORM)
- JWT & Session-based Authentication

**Database & Cache:**
- PostgreSQL 16
- Redis 7
- golang-migrate (migrations)

**Monitoring & Observability:**
- Prometheus (metrics collection)
- Grafana (visualization)
- Custom business metrics
- Health check endpoints

**Development & Testing:**
- testify (testing framework)
- go-sqlmock (database mocking)
- miniredis (Redis mocking)
- Air (live reload)
- Pre-commit hooks

**DevOps:**
- Docker & Docker Compose
- Multi-stage builds
- Health checks
- Graceful shutdown

**Documentation:**
- Swagger/OpenAPI
- Comprehensive README
- Docker setup guide

## Environment Configuration

The project uses environment variables for configuration. Two template files are provided:

| File | Purpose | Use When |
|------|---------|----------|
| `.env.docker` | Docker deployment template | Running with `docker-compose` |
| `.env.example` | Local development template | Running locally without Docker |

### Key Differences:

**`.env.docker`** uses Docker service names:
- `DATABASE_URL`: `postgres:5432` (container name)
- `REDIS_ADDR`: `redis:6379` (container name)
- `TRUSTED_PROXIES`: `0.0.0.0/0` (Docker network)

**`.env.example`** uses localhost:
- `DATABASE_URL`: `localhost:5432` (local machine)
- `REDIS_ADDR`: `localhost:6379` (local machine)
- `TRUSTED_PROXIES`: `127.0.0.1,::1` (localhost only)

Both files contain placeholders - generate secure secrets for production:
```bash
openssl rand -base64 32  # Generate JWT_SECRET
```

## Documentation

- [DOCKER_SETUP.md](DOCKER_SETUP.md) - Complete Docker & monitoring setup guide
- [CONTRIBUTING.md](CONTRIBUTING.md) - Contribution guidelines
- [Swagger UI](http://localhost:8080/swagger/index.html) - API documentation (when running)

## License

MIT
