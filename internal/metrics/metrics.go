package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	HttpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status_code"},
	)

	HttpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	CacheHitsTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
	)

	CacheMissesTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
	)

	DbQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table"},
	)

	DbQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation", "table"},
	)

	OrdersCreatedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "orders_created_total",
			Help: "Total number of orders created",
		},
	)

	ProductsViewedTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "products_viewed_total",
			Help: "Total number of product views",
		},
	)

	UsersRegisteredTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "users_registered_total",
			Help: "Total number of users registered",
		},
	)

	AuthAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_attempts_total",
			Help: "Total number of authentication attempts",
		},
		[]string{"status"},
	)
)

// PrometheusMiddleware creates a Gin middleware that records metrics for each HTTP request
func PrometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		c.Next()

		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		HttpRequestsTotal.WithLabelValues(method, path, statusCode).Inc()
		HttpRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}

// RecordCacheHit records a cache hit
func RecordCacheHit() {
	CacheHitsTotal.Inc()
}

// RecordCacheMiss records a cache miss
func RecordCacheMiss() {
	CacheMissesTotal.Inc()
}

// RecordDbQuery records a database query
func RecordDbQuery(operation, table string, duration float64) {
	DbQueriesTotal.WithLabelValues(operation, table).Inc()
	DbQueryDuration.WithLabelValues(operation, table).Observe(duration)
}

// RecordOrderCreated records an order creation
func RecordOrderCreated() {
	OrdersCreatedTotal.Inc()
}

// RecordProductViewed records a product view
func RecordProductViewed() {
	ProductsViewedTotal.Inc()
}

// RecordUserRegistered records a user registration
func RecordUserRegistered() {
	UsersRegisteredTotal.Inc()
}

// RecordAuthAttempt records an authentication attempt
func RecordAuthAttempt(success bool) {
	status := "failure"
	if success {
		status = "success"
	}
	AuthAttemptsTotal.WithLabelValues(status).Inc()
}
