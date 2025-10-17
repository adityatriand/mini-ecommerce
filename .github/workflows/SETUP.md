# GitHub Actions Workflows for Mini E-Commerce Microservices

This directory contains GitHub Actions workflows for CI/CD of the Mini E-Commerce microservices project.

## Workflows

### 1. CI/CD Pipeline (`ci.yml`)

**Triggers:**
- Push to `main` or `develop` branches
- Pull requests to `main` or `develop` branches

**Jobs:**

#### Test Microservices
- **Services**: PostgreSQL 16, Redis 7
- **Steps**:
  - Checkout code
  - Setup Go 1.24
  - Cache Go modules
  - Install dependencies
  - Run tests for shared libraries
  - Run tests for each service (user, product, order, gateway)
  - Run integration tests
  - Generate test coverage
  - Upload coverage to Codecov

#### Lint Code
- **Tools**: golangci-lint
- **Purpose**: Code quality and style checking

#### Build Microservices
- **Matrix Strategy**: Builds each service (gateway, user, product, order)
- **Steps**:
  - Build Go binaries
  - Test Docker builds

#### Security Scan
- **Tools**: Gosec Security Scanner
- **Output**: SARIF format for GitHub Security tab

#### Deploy to Staging
- **Trigger**: Push to `develop` branch
- **Purpose**: Deploy to staging environment

#### Deploy to Production
- **Trigger**: Push to `main` branch
- **Purpose**: Deploy to production environment

## Environment Variables

The following environment variables are used in the workflows:

```yaml
env:
  GO_VERSION: '1.24'
  DOCKER_BUILDKIT: 1
```

## Service Dependencies

The workflows use the following services:

- **PostgreSQL 16**: Database for testing
- **Redis 7**: Cache layer for testing

## Coverage Reporting

Test coverage is generated for all services and uploaded to Codecov for tracking.

## Security

Security scanning is performed using Gosec, which checks for common security issues in Go code.

## Deployment

Deployment workflows are configured for both staging and production environments. You'll need to customize the deployment steps based on your infrastructure:

- **Kubernetes**: Use `kubectl apply` commands
- **Docker Swarm**: Use `docker stack deploy` commands
- **Cloud Platforms**: Use platform-specific deployment tools

## Customization

### Adding New Services

To add a new service to the CI/CD pipeline:

1. Add the service to the matrix strategy in the build job
2. Add test steps for the new service
3. Update the coverage generation script

### Environment-Specific Configuration

Create environment-specific configuration files:

- `.github/workflows/staging.yml` - Staging-specific workflows
- `.github/workflows/production.yml` - Production-specific workflows

### Secrets

Add the following secrets to your GitHub repository:

- `DATABASE_URL` - Database connection string
- `REDIS_ADDR` - Redis connection string
- `JWT_SECRET` - JWT signing secret
- `DOCKER_REGISTRY_TOKEN` - Docker registry authentication
- `KUBECONFIG` - Kubernetes configuration (if using K8s)

## Monitoring

The workflows include monitoring and alerting:

- **Test Results**: Automatic notifications on test failures
- **Security Issues**: SARIF uploads to GitHub Security tab
- **Coverage Reports**: Codecov integration for coverage tracking

## Troubleshooting

### Common Issues

1. **Test Failures**: Check service dependencies and environment variables
2. **Build Failures**: Verify Dockerfile and build context
3. **Security Scan Failures**: Review and fix security issues
4. **Deployment Failures**: Check deployment configuration and secrets

### Debug Mode

To enable debug mode, add `ACTIONS_STEP_DEBUG: true` to the environment variables.

## Best Practices

1. **Keep workflows fast**: Use caching and parallel jobs
2. **Fail fast**: Run quick checks first (lint, format)
3. **Security first**: Always run security scans
4. **Test thoroughly**: Include unit, integration, and e2e tests
5. **Monitor deployments**: Use health checks and rollback strategies