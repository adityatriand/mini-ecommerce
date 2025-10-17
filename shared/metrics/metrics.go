package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request metrics
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status", "service"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status", "service"},
	)

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path", "service"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "path", "status", "service"},
	)

	// Active connections
	httpRequestsInFlight = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Current number of HTTP requests being processed",
		},
		[]string{"service"},
	)
)

// PrometheusMiddleware creates a Gin middleware for Prometheus metrics
func PrometheusMiddleware(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Increment in-flight requests
		httpRequestsInFlight.WithLabelValues(serviceName).Inc()
		defer httpRequestsInFlight.WithLabelValues(serviceName).Dec()

		// Record request size
		requestSize := computeApproximateRequestSize(c)
		httpRequestSize.WithLabelValues(c.Request.Method, path, serviceName).Observe(float64(requestSize))

		// Process request
		c.Next()

		// Record metrics
		status := strconv.Itoa(c.Writer.Status())
		duration := time.Since(start).Seconds()
		responseSize := float64(c.Writer.Size())

		httpRequestsTotal.WithLabelValues(c.Request.Method, path, status, serviceName).Inc()
		httpRequestDuration.WithLabelValues(c.Request.Method, path, status, serviceName).Observe(duration)
		httpResponseSize.WithLabelValues(c.Request.Method, path, status, serviceName).Observe(responseSize)
	}
}

// computeApproximateRequestSize computes the approximate size of the request
func computeApproximateRequestSize(c *gin.Context) int {
	s := 0
	if c.Request.URL != nil {
		s = len(c.Request.URL.Path)
	}

	s += len(c.Request.Method)
	s += len(c.Request.Proto)
	for name, values := range c.Request.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(c.Request.Host)

	// N.B. c.Request.Form and c.Request.MultipartForm are assumed to be included in c.Request.Body.

	if c.Request.ContentLength != -1 {
		s += int(c.Request.ContentLength)
	}
	return s
}
