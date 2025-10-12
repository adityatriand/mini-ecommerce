# Docker Setup Guide

This guide explains how to run the Mini E-Commerce application using Docker and Docker Compose with integrated monitoring (Prometheus & Grafana).

## Prerequisites

- Docker (v20.10 or higher)
- Docker Compose (v2.0 or higher)

## Architecture

The Docker setup includes the following services:

1. **PostgreSQL** - Main database (port 5432)
2. **Redis** - Cache layer (port 6379)
3. **Go Application** - E-commerce API (port 8080)
4. **Prometheus** - Metrics collection (port 9090)
5. **Grafana** - Metrics visualization (port 3000)

## Quick Start

### 1. Configure Environment Variables

Copy the Docker environment template:
```bash
cp .env.docker .env
```

Edit `.env` and update:
- `JWT_SECRET` - Your JWT secret key (important for production!)
- `GRAFANA_PASSWORD` - Grafana admin password
- `APP_ENV` - Set to `production` for production deployment

### 2. Build and Start All Services

```bash
docker-compose up --build
```

Or run in detached mode:
```bash
docker-compose up -d --build
```

### 3. Verify Services

Check all services are running:
```bash
docker-compose ps
```

You should see all services with status "Up (healthy)".

### 4. Access the Services

- **API**: http://localhost:8080
- **API Documentation (Swagger)**: http://localhost:8080/swagger/index.html
- **Prometheus**: http://localhost:9090
- **Grafana**: http://localhost:3000 (admin/admin or admin/your-password)

## Health Checks

The application provides three health check endpoints:

- `GET /health` - Simple health check
- `GET /health/ready` - Readiness probe (checks database & Redis)
- `GET /health/live` - Liveness probe

Example:
```bash
curl http://localhost:8080/health/ready
```

## Monitoring with Prometheus & Grafana

### Prometheus

1. Access Prometheus at http://localhost:9090
2. Query metrics using PromQL, for example:
   - `http_requests_total` - Total HTTP requests
   - `http_request_duration_seconds` - Request latency
   - `cache_hits_total` - Cache hits
   - `cache_misses_total` - Cache misses

### Grafana

1. Access Grafana at http://localhost:3000
2. Login with `admin` / `admin` (or your configured password)
3. The "Mini E-Commerce Dashboard" is automatically provisioned
4. View metrics including:
   - Request rate and duration
   - HTTP status codes distribution
   - Cache performance
   - Memory and goroutines
   - Database query performance

### Available Metrics

**HTTP Metrics:**
- `http_requests_total{method, path, status_code}` - Total HTTP requests
- `http_request_duration_seconds{method, path}` - Request duration histogram

**Cache Metrics:**
- `cache_hits_total` - Total cache hits
- `cache_misses_total` - Total cache misses

**Database Metrics:**
- `db_queries_total{operation, table}` - Total database queries
- `db_query_duration_seconds{operation, table}` - Query duration histogram

**Business Metrics:**
- `orders_created_total` - Total orders created
- `products_viewed_total` - Total product views
- `users_registered_total` - Total users registered
- `auth_attempts_total{status}` - Authentication attempts (success/failure)

**Go Runtime Metrics:**
- `go_memstats_alloc_bytes` - Memory allocation
- `go_goroutines` - Number of goroutines
- And many more Go runtime metrics...

## Docker Commands

### Start services
```bash
docker-compose up -d
```

### Stop services
```bash
docker-compose down
```

### Stop services and remove volumes (WARNING: deletes all data)
```bash
docker-compose down -v
```

### View logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f app
docker-compose logs -f postgres
docker-compose logs -f prometheus
docker-compose logs -f grafana
```

### Rebuild a specific service
```bash
docker-compose up -d --build app
```

### Execute commands in containers
```bash
# Access application container
docker-compose exec app sh

# Access PostgreSQL
docker-compose exec postgres psql -U ecommerce -d mini_ecommerce

# Access Redis CLI
docker-compose exec redis redis-cli
```

### Check service health
```bash
docker-compose ps
```

## Database Management

### Run Migrations Manually

Migrations run automatically on startup, but you can trigger them manually:

```bash
docker-compose exec app ./main migrate
```

### Access Database

```bash
docker-compose exec postgres psql -U ecommerce -d mini_ecommerce
```

Common PostgreSQL commands:
```sql
\dt                -- List tables
\d+ table_name     -- Describe table
SELECT * FROM users LIMIT 10;
```

## Troubleshooting

### Service won't start

Check logs:
```bash
docker-compose logs app
```

### Database connection issues

1. Ensure PostgreSQL is healthy:
```bash
docker-compose ps postgres
```

2. Check database logs:
```bash
docker-compose logs postgres
```

3. Verify connectivity:
```bash
docker-compose exec app wget -O- http://postgres:5432 || echo "Cannot connect"
```

### Redis connection issues

1. Test Redis:
```bash
docker-compose exec redis redis-cli ping
```

### Application won't build

1. Clean Docker cache:
```bash
docker-compose down
docker system prune -a
docker-compose up --build
```

### Metrics not appearing in Prometheus

1. Check Prometheus targets: http://localhost:9090/targets
2. Verify app metrics endpoint: http://localhost:8080/metrics
3. Check Prometheus configuration: `monitoring/prometheus/prometheus.yml`

### Grafana dashboard is empty

1. Verify Prometheus datasource is connected in Grafana
2. Check data retention in Prometheus (default: 15 days)
3. Generate some traffic to the application to create metrics

## Production Deployment

For production deployment, make these changes:

1. **Update Environment Variables:**
   ```bash
   APP_ENV=production
   JWT_SECRET=<strong-random-secret>
   GRAFANA_PASSWORD=<strong-password>
   ```

2. **Use External Databases (Recommended):**
   - Use managed PostgreSQL (RDS, Cloud SQL, etc.)
   - Use managed Redis (ElastiCache, MemoryStore, etc.)
   - Update `docker-compose.yml` to remove database services
   - Update connection strings in environment variables

3. **Enable TLS/HTTPS:**
   - Add reverse proxy (Nginx, Traefik) with SSL certificates
   - Configure HTTPS in application

4. **Resource Limits:**
   Add resource limits to `docker-compose.yml`:
   ```yaml
   services:
     app:
       deploy:
         resources:
           limits:
             cpus: '1'
             memory: 512M
   ```

5. **Persistent Volumes:**
   Ensure volumes are properly backed up:
   ```bash
   docker-compose exec postgres pg_dump -U ecommerce mini_ecommerce > backup.sql
   ```

6. **Security:**
   - Change all default passwords
   - Use secrets management (Docker secrets, HashiCorp Vault)
   - Implement rate limiting
   - Enable authentication for Prometheus and Grafana

## Development Tips

### Live Reload with Air

For development with live reload, you can use Air (already configured):

```bash
# Run locally with Air (outside Docker)
air
```

### Running Tests in Docker

```bash
docker-compose exec app go test ./... -v
```

### Accessing Container Shell

```bash
docker-compose exec app sh
```

## Network Architecture

All services run on a custom bridge network (`ecommerce-network`) which allows:
- Service discovery by service name (e.g., `postgres:5432`)
- Isolated network from host
- Easy inter-service communication

## Volumes

The setup uses named volumes for data persistence:

- `postgres_data` - PostgreSQL data
- `redis_data` - Redis data
- `prometheus_data` - Prometheus metrics data
- `grafana_data` - Grafana dashboards and settings

To backup volumes:
```bash
docker run --rm -v postgres_data:/data -v $(pwd):/backup alpine tar czf /backup/postgres_backup.tar.gz -C /data .
```

## Support

For issues or questions:
1. Check logs: `docker-compose logs -f`
2. Verify health checks: `curl http://localhost:8080/health/ready`
3. Review configuration files in `monitoring/` directory
4. Check GitHub issues or create a new one
