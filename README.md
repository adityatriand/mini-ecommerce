# Mini E-Commerce

A mini e-commerce backend application built with Go, Gin, PostgreSQL, and Redis.

## Features

- User authentication with JWT and session management
- Product management with caching
- Order management
- RESTful API
- Swagger documentation
- Database migrations
- Redis caching for heavy endpoints

## Getting Started

### Prerequisites

- Go 1.24.5 or higher
- PostgreSQL
- Redis
- Make

### Setup

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

### Configuration

Copy the example environment file and configure:

```bash
cp .env.example .env
```

Update the values in `.env` according to your environment.

### Database Migrations

Run migrations:

```bash
make migrate-up
```

Create a new migration:

```bash
make migrate-create name=your_migration_name
```

### Running the Application

```bash
go run cmd/main.go
```

The server will start on the port specified in your configuration (default: 8080).

## API Documentation

After starting the server, visit:

```
http://localhost:8080/swagger/index.html
```

## Development

### Running Tests

Run all tests:
```bash
make test
```

Run specific package tests:
```bash
make test-auth
make test-product
make test-order
```

### Git Hooks

Pre-commit hooks are automatically installed during setup. They will:
- Run tests for packages you modified
- Prevent commits if tests fail
- Keep the codebase stable

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
│   ├── cache/             # Caching layer
│   ├── config/            # Configuration
│   ├── database/          # Database connection
│   ├── logger/            # Logging
│   ├── middleware/        # HTTP middleware
│   └── response/          # Response helpers
├── routes/                # Route registration
├── migrations/            # Database migrations
├── scripts/               # Utility scripts and hooks
└── docs/                  # Swagger documentation
```

## Available Commands

```bash
make setup              # Setup development environment
make install-hooks      # Install git hooks
make test               # Run all tests
make test-auth          # Run auth tests
make test-product       # Run product tests
make test-order         # Run order tests
make migrate-up         # Run database migrations
make migrate-down       # Rollback last migration
make migrate-create     # Create new migration
```

## License

MIT
