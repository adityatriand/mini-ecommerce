# Contributing Guide for Mini E-Commerce Microservices

Thank you for contributing to the Mini E-Commerce microservices project! This guide will help you get started with development and understand our contribution process.

## 🏗️ Architecture Overview

This project uses a microservices architecture with the following services:

- **API Gateway** (Port 8000) - Request routing and authentication
- **User Service** (Port 8001) - User management and authentication
- **Product Service** (Port 8002) - Product catalog and inventory
- **Order Service** (Port 8003) - Order processing and management

## 🚀 Getting Started

### Prerequisites

- Go 1.24+
- Docker & Docker Compose
- Make
- Git

### First Time Setup

#### Option 1: Automated Setup (Recommended)

The fastest way to get started:

```bash
# Clone the repository
git clone <repository-url>
cd mini-ecommerce

# Run the setup script
chmod +x scripts/setup-microservices.sh
./scripts/setup-microservices.sh
```

This script will:
- Check prerequisites
- Setup Go modules
- Install Git hooks
- Create necessary directories
- Setup environment files
- Build all services
- Run tests

#### Option 2: Manual Setup

If you prefer manual setup:

```bash
# Clone the repository
git clone <repository-url>
cd mini-ecommerce

# Setup Go modules
go mod tidy
go mod download

# Install Git hooks
cp .githooks/pre-commit .git/hooks/pre-commit
chmod +x .git/hooks/pre-commit

# Create environment file
cp .env.microservices.example .env.microservices

# Build services
make build-all

# Start services
make dev-up
```

## 🛠️ Development Workflow

### Starting Development

1. **Start all services**:
   ```bash
   make dev-up
   ```

2. **Check service health**:
   ```bash
   make health-check
   ```

3. **View logs**:
   ```bash
   make dev-logs
   ```

### Working on a Specific Service

1. **Start individual service**:
   ```bash
   make dev-up-user
   make dev-up-product
   make dev-up-order
   make dev-up-gateway
   ```

2. **Run tests for specific service**:
   ```bash
   make test-user
   make test-product
   make test-order
   make test-gateway
   ```

3. **View logs for specific service**:
   ```bash
   make logs-user
   make logs-product
   make logs-order
   make logs-gateway
   ```

### Live Reloading

For development with live reloading, you can use Air:

```bash
# Install Air (if not already installed)
go install github.com/cosmtrek/air@latest

# Run Air for a specific service
cd services/user
air

# Or run Air from project root
air -c .air.toml -s user
```

## 📁 Project Structure

```
mini-ecommerce/
├── services/                    # Microservices
│   ├── gateway/                # API Gateway
│   │   ├── cmd/main.go
│   │   ├── internal/
│   │   ├── configs/
│   │   └── Dockerfile
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

# Integration tests
make test-integration
```

### Test Coverage

```bash
# Generate coverage report
make test-coverage

# View coverage in browser
go tool cover -html=coverage.out
```

### Writing Tests

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test service-to-service communication
- **End-to-End Tests**: Test complete user workflows

Example test structure:
```go
func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    // Act
    // Assert
}
```

## 🔧 Code Quality

### Pre-commit Hooks

The project includes pre-commit hooks that automatically run:

- `go mod tidy` - Clean dependencies
- `go vet` - Static analysis
- `go fmt` - Code formatting
- Unit tests
- Test coverage

### Code Style

- Follow Go conventions and best practices
- Use meaningful variable and function names
- Add comments for public functions
- Keep functions small and focused
- Use proper error handling

### Linting

```bash
# Run golangci-lint
golangci-lint run

# Run specific linters
golangci-lint run --enable=gofmt,govet,errcheck
```

## 📝 Making Changes

### Branch Strategy

- **`main`** - Production-ready code
- **`develop`** - Integration branch for features
- **`feature/*`** - Feature branches
- **`bugfix/*`** - Bug fix branches
- **`hotfix/*`** - Critical production fixes

### Commit Messages

Use conventional commit format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes
- `refactor`: Code refactoring
- `test`: Test changes
- `chore`: Build process or auxiliary tool changes

Examples:
```
feat(user): add user registration endpoint
fix(product): resolve stock update race condition
docs(api): update API documentation
```

### Pull Request Process

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes** and commit:
   ```bash
   git add .
   git commit -m "feat(service): add new feature"
   ```

3. **Push your branch**:
   ```bash
   git push origin feature/your-feature-name
   ```

4. **Create a Pull Request**:
   - Use the PR template
   - Add a clear description
   - Link related issues
   - Request reviews from team members

5. **Address feedback** and make necessary changes

6. **Merge** after approval and CI passes

## 🐛 Bug Reports

When reporting bugs, please include:

1. **Environment**: OS, Go version, Docker version
2. **Steps to reproduce**: Clear, numbered steps
3. **Expected behavior**: What should happen
4. **Actual behavior**: What actually happens
5. **Logs**: Relevant error messages or logs
6. **Screenshots**: If applicable

## 💡 Feature Requests

When requesting features, please include:

1. **Use case**: Why is this feature needed?
2. **Proposed solution**: How should it work?
3. **Alternatives**: Other solutions considered
4. **Impact**: Who will benefit from this feature?

## 🔍 Code Review Guidelines

### For Reviewers

- **Be constructive**: Provide helpful feedback
- **Be specific**: Point out exact issues
- **Be respectful**: Maintain a positive tone
- **Focus on code**: Avoid personal comments
- **Suggest improvements**: Don't just point out problems

### For Authors

- **Be responsive**: Address feedback promptly
- **Be open**: Accept constructive criticism
- **Be thorough**: Test your changes
- **Be clear**: Explain complex logic
- **Be patient**: Reviews take time

## 🚀 Deployment

### Local Development

```bash
# Start all services
make dev-up

# Check health
make health-check
```

### Staging Environment

```bash
# Deploy to staging
make deploy-staging
```

### Production Environment

```bash
# Deploy to production
make deploy-production
```

## 📊 Monitoring

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

## 🆘 Getting Help

- **Documentation**: Check the README and service-specific docs
- **Issues**: Search existing issues or create a new one
- **Discussions**: Use GitHub Discussions for questions
- **Team**: Contact the development team

## 📄 License

By contributing to this project, you agree that your contributions will be licensed under the MIT License.

## 🙏 Thank You

Thank you for contributing to the Mini E-Commerce microservices project! Your contributions help make this project better for everyone.

---

**Happy coding! 🚀**