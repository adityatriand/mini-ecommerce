# Loki Centralized Logging Setup

This guide explains how to set up and use Loki for centralized logging in the mini-ecommerce project.

## Architecture

We use the **Docker Loki Logging Driver** approach for log collection:

```
Application (Zap JSON logs) → stdout/stderr → Docker Loki Driver → Loki → Grafana
```

### Benefits

- **Zero Application Changes** - Your existing Zap logger already outputs JSON to stdout
- **No Additional Containers** - No need for Promtail sidecar
- **Automatic Collection** - Docker handles log shipping automatically
- **Container Metadata** - Automatically includes container name, ID, labels
- **Simple Configuration** - Just Docker Compose configuration

## Prerequisites

Before starting, you need to install the Docker Loki logging plugin:

```bash
docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions
```

Verify the plugin is installed:

```bash
docker plugin ls
```

You should see output like:
```
ID             NAME          DESCRIPTION           ENABLED
ac720b8fcfdb   loki:latest   Loki Logging Driver   true
```

## Configuration

### 1. Loki Service

The `docker-compose.yml` includes a Loki service that:
- Listens on port 3100
- Stores logs in a persistent volume (`loki_data`)
- Uses configuration from `monitoring/loki/loki-config.yml`

### 2. Docker Logging Driver

Each service in `docker-compose.yml` is configured with the Loki logging driver:

```yaml
logging:
  driver: loki
  options:
    loki-url: "http://localhost:3100/loki/api/v1/push"
    loki-retries: "5"
    loki-batch-size: "400"
    labels: "service=app,environment=docker"
```

This configuration:
- Sends all container stdout/stderr to Loki
- Includes service labels for filtering
- Retries failed log pushes up to 5 times
- Batches logs for efficiency

### 3. Grafana Datasource

Loki is automatically provisioned as a datasource in Grafana via:
`monitoring/grafana/provisioning/datasources/datasource.yml`

## Usage

### Start the Stack

```bash
# Install Loki Docker plugin first (one-time setup)
docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions

# Start all services
docker-compose up -d
```

### Access Logs in Grafana

1. Open Grafana: http://localhost:3000 (admin/admin)
2. Go to "Explore" (compass icon in left sidebar)
3. Select "Loki" datasource from the dropdown

### Query Examples

#### View All Application Logs
```logql
{service="app"}
```

#### View Errors Only
```logql
{service="app"} |= "error"
```

#### View Logs from Specific Endpoint
```logql
{service="app"} |= "/api/products"
```

#### View PostgreSQL Logs
```logql
{service="postgres"}
```

#### View Redis Logs
```logql
{service="redis"}
```

#### Count Error Rate (Last 5 Minutes)
```logql
sum(rate({service="app"} |= "error" [5m]))
```

#### Extract and Filter JSON Fields
Since Zap outputs structured JSON logs:

```logql
{service="app"} | json | level="error"
```

```logql
{service="app"} | json | method="POST" | path="/api/orders"
```

#### Time Range Queries
```logql
{service="app"} | json | level="error" | __timestamp__ > now() - 1h
```

### Log Retention

By default, logs are retained for **7 days** (168 hours). You can change this in `monitoring/loki/loki-config.yml`:

```yaml
limits_config:
  retention_period: 168h  # Change to desired retention
```

## Troubleshooting

### Logs Not Appearing in Loki

1. **Check Loki plugin is installed:**
   ```bash
   docker plugin ls
   ```

2. **Check container logs:**
   ```bash
   docker logs mini-ecommerce-app
   ```

3. **Check Loki service:**
   ```bash
   docker logs mini-ecommerce-loki
   ```

4. **Verify Loki is receiving logs:**
   ```bash
   curl -s "http://localhost:3100/loki/api/v1/labels"
   ```

### Loki Docker Driver Not Found

If you see an error like "docker: logging driver does not exist: loki", install the plugin:

```bash
docker plugin install grafana/loki-docker-driver:latest --alias loki --grant-all-permissions
```

### Logs Delayed or Missing

Increase batch size and check retry settings in docker-compose.yml:

```yaml
logging:
  driver: loki
  options:
    loki-batch-size: "1000"  # Increase from 400
    loki-retries: "10"        # Increase from 5
```

## Integration with Existing Monitoring

### Correlating Logs with Metrics

In Grafana, you can:
1. Split view to show metrics (Prometheus) and logs (Loki) side-by-side
2. Click on a metric spike to view corresponding logs
3. Use derived fields to link trace IDs between datasources

### Adding Logs to Dashboard

You can add log panels to your existing `mini-ecommerce.json` dashboard:

1. Edit dashboard in Grafana
2. Add new panel
3. Select Loki datasource
4. Choose visualization: "Logs" or "Time series" for log counts

Example panel query:
```logql
sum by (level) (count_over_time({service="app"} | json [5m]))
```

## Zap Logger Integration

Your existing Zap logger (internal/logger/logger.go) works perfectly with Loki because:

1. **JSON Output** - Zap outputs structured JSON logs in production
2. **Structured Fields** - All log fields are queryable in Loki
3. **Log Levels** - Filter by level (debug, info, warn, error)
4. **No Code Changes** - Everything works via stdout

Example Zap log output:
```json
{
  "level": "info",
  "ts": "2024-01-15T10:30:45.123Z",
  "caller": "handlers/product.go:45",
  "msg": "Product retrieved successfully",
  "product_id": "123",
  "user_id": "456"
}
```

Query in Loki:
```logql
{service="app"} | json | product_id="123"
```

## Best Practices

1. **Use Structured Logging** - Continue using Zap's structured fields for rich queryability
2. **Add Context** - Include user_id, request_id, trace_id in log fields
3. **Log Levels** - Use appropriate levels (debug in dev, info/warn/error in prod)
4. **Label Cardinality** - Keep Docker labels low (service, environment) - don't add high-cardinality labels
5. **Query Optimization** - Filter by labels first, then parse JSON

## Monitoring Loki

Check Loki's internal metrics:
```bash
curl -s http://localhost:3100/metrics | grep loki_
```

Key metrics:
- `loki_ingester_streams_created_total` - Streams created
- `loki_distributor_bytes_received_total` - Bytes ingested
- `loki_ingester_chunks_flushed_total` - Chunks persisted

## References

- [Loki Documentation](https://grafana.com/docs/loki/latest/)
- [Docker Loki Driver](https://grafana.com/docs/loki/latest/send-data/docker-driver/)
- [LogQL Query Language](https://grafana.com/docs/loki/latest/query/)
- [Zap Logger](https://github.com/uber-go/zap)
