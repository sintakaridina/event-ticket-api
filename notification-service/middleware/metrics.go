package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics variables
var (
	// RequestDuration tracks the duration of HTTP requests
	RequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "notification_http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// TotalRequests tracks the total number of HTTP requests
	TotalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "notification_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// RequestSize tracks the size of HTTP requests
	RequestSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "notification_http_request_size_bytes",
			Help: "Size of HTTP requests in bytes",
		},
		[]string{"method", "path"},
	)

	// ResponseSize tracks the size of HTTP responses
	ResponseSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "notification_http_response_size_bytes",
			Help: "Size of HTTP responses in bytes",
		},
		[]string{"method", "path"},
	)
)

// InitMetrics initializes the Prometheus metrics
func InitMetrics() {
	// Register metrics with Prometheus
	prometheus.MustRegister(RequestDuration)
	prometheus.MustRegister(TotalRequests)
	prometheus.MustRegister(RequestSize)
	prometheus.MustRegister(ResponseSize)
}

// Metrics is a middleware for collecting HTTP metrics
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Get request details
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		// Track request size
		RequestSize.WithLabelValues(method, path).Observe(float64(c.Request.ContentLength))

		// Process request
		c.Next()

		// Calculate duration
		duration := time.Since(startTime).Seconds()

		// Get status code
		statusCode := strconv.Itoa(c.Writer.Status())

		// Track metrics
		RequestDuration.WithLabelValues(method, path, statusCode).Observe(duration)
		TotalRequests.WithLabelValues(method, path, statusCode).Inc()
		ResponseSize.WithLabelValues(method, path).Observe(float64(c.Writer.Size()))
	}
}