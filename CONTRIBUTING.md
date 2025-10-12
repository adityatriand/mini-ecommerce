# Contributing Guide

Thank you for contributing to this project!

## Getting Started

### First Time Setup

#### Option 1: Docker (Recommended)

The fastest way to get started:

```bash
# Clone and setup
git clone <repository-url>
cd mini-e-commerce

# Copy environment file
cp .env.docker .env

# Start all services
make docker-up
```

This starts the app, database, Redis, Prometheus, and Grafana.

#### Option 2: Local Development

For local development without Docker:

```bash
# Setup development environment
make setup

# Configure environment
cp .env.example .env
# Edit .env with your local database/Redis credentials

# Run migrations
make migrate-up

# Start the application
go run cmd/main.go
```

### What Gets Configured

The setup configures:
- Git hooks path to use `.githooks` directory (version controlled)
- Downloads Go dependencies
- Installs development tools

The pre-commit hook:
- Runs tests for packages you modified
- Prevents commits if tests fail
- Works dynamically with all test packages

## Development Workflow

### Making Changes

1. Create a new branch:
```bash
git checkout -b feat/your-feature
```

2. Make your changes

3. Write tests for your changes

4. Commit your changes:
```bash
git add .
git commit -m "Add your feature"
```

The pre-commit hook will automatically run tests for affected packages.

### Running Tests

Run all tests (187 test cases):
```bash
make test
```

Run specific package tests:
```bash
make test-auth      # Authentication tests
make test-product   # Product tests
make test-order     # Order tests
make test-utils     # Utility tests
make test-cache     # Cache tests
make test-middleware # Middleware tests
```

Test the pre-commit hook manually:
```bash
make pre-commit
```

### Working with Docker

Start services:
```bash
make docker-up
```

View logs:
```bash
make docker-logs
```

Stop services:
```bash
make docker-down
```

Check service status:
```bash
make docker-ps
```

### Commit Guidelines

The pre-commit hook ensures:
- All tests pass before commit
- Only affected packages are tested
- Fast feedback loop

If tests fail, fix them before committing.

### Bypass Hook (Emergency Only)

In rare cases, you can skip the hook:
```bash
git commit --no-verify -m "Emergency fix"
```

Use this sparingly.

## Project Structure

```
.
├── cmd/                    # Application entrypoints
├── internal/
│   ├── auth/              # Authentication & user management
│   ├── product/           # Product management
│   ├── order/             # Order management
│   ├── cache/             # Redis caching layer
│   ├── config/            # Configuration with Viper
│   ├── database/          # Database connection
│   ├── health/            # Health check handlers
│   ├── logger/            # Structured logging
│   ├── metrics/           # Prometheus metrics
│   ├── middleware/        # HTTP middleware
│   ├── response/          # Response helpers
│   ├── swagger/           # Swagger setup
│   └── utils/             # Utility functions
├── routes/                # Route registration
├── migrations/            # Database migrations
├── monitoring/            # Prometheus & Grafana configs
│   ├── prometheus/        # Prometheus config
│   └── grafana/           # Grafana dashboards
├── scripts/
│   ├── hooks/             # Git hooks
│   └── setup.sh           # Setup script
├── Dockerfile             # Docker image definition
├── docker-compose.yml     # Multi-service orchestration
└── Makefile               # Build & deployment commands
```

## Writing Tests

### Test File Naming

Tests should be in the same package as the code being tested:
```
internal/auth/service.go
internal/auth/service_test.go
```

### Test Coverage

When adding new features:
1. Write unit tests for all public functions
2. Test both success and error cases
3. Use mocks for external dependencies

### Example Test

```go
func TestService_RegisterUser(t *testing.T) {
    t.Run("should register user successfully", func(t *testing.T) {
        mockRepo := new(MockRepository)
        service := NewService(mockRepo, ...)

        input := RegisterRequest{
            Email:    "test@example.com",
            Password: "password123",
        }

        mockRepo.On("Create", mock.Anything, mock.Anything).Return(nil)

        user, err := service.RegisterUser(ctx, input)

        require.NoError(t, err)
        assert.NotNil(t, user)
        mockRepo.AssertExpectations(t)
    })
}
```

## Questions?

If you have questions or run into issues, please open an issue on GitHub.
