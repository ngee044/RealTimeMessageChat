package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP request counter
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTP request duration histogram
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint", "status"},
	)

	// HTTP request size histogram
	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "HTTP request size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	// HTTP response size histogram
	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "HTTP response size in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	// Active HTTP connections gauge
	httpActiveConnections = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_active_connections",
			Help: "Number of active HTTP connections",
		},
	)

	// RabbitMQ messages published counter
	rabbitmqMessagesPublished = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rabbitmq_messages_published_total",
			Help: "Total number of messages published to RabbitMQ",
		},
		[]string{"queue", "status"},
	)

	// Database query counter
	databaseQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "database_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "status"},
	)

	// Database query duration histogram
	databaseQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "database_query_duration_seconds",
			Help:    "Database query duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)

	// Redis operations counter
	redisOperationsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redis_operations_total",
			Help: "Total number of Redis operations",
		},
		[]string{"operation", "status"},
	)

	// Redis operation duration histogram
	redisOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redis_operation_duration_seconds",
			Help:    "Redis operation duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"operation"},
	)
)

// PrometheusMetrics is a middleware that collects Prometheus metrics
func PrometheusMetrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Increment active connections
		httpActiveConnections.Inc()
		defer httpActiveConnections.Dec()

		// Record start time
		start := time.Now()

		// Get request size
		requestSize := computeApproximateRequestSize(c.Request)

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Get status code
		status := strconv.Itoa(c.Writer.Status())

		// Get endpoint (path template)
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = c.Request.URL.Path
		}

		// Get method
		method := c.Request.Method

		// Record metrics
		httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint, status).Observe(duration)
		httpRequestSize.WithLabelValues(method, endpoint).Observe(float64(requestSize))
		httpResponseSize.WithLabelValues(method, endpoint).Observe(float64(c.Writer.Size()))
	}
}

// computeApproximateRequestSize estimates the size of the HTTP request
func computeApproximateRequestSize(r interface{}) int {
	// Type assertion to get the actual request
	req, ok := r.(*interface{})
	if !ok {
		return 0
	}
	_ = req
	// Simplified estimation
	return 1024 // Placeholder
}

// RecordRabbitMQPublish records a RabbitMQ message publish
func RecordRabbitMQPublish(queue string, success bool) {
	status := "success"
	if !success {
		status = "error"
	}
	rabbitmqMessagesPublished.WithLabelValues(queue, status).Inc()
}

// RecordDatabaseQuery records a database query
func RecordDatabaseQuery(operation string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	databaseQueriesTotal.WithLabelValues(operation, status).Inc()
	databaseQueryDuration.WithLabelValues(operation).Observe(duration.Seconds())
}

// RecordRedisOperation records a Redis operation
func RecordRedisOperation(operation string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}
	redisOperationsTotal.WithLabelValues(operation, status).Inc()
	redisOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())
}
