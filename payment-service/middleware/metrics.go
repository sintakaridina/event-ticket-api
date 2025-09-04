package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics variables
var (
	// httpRequestDuration tracks the duration of HTTP requests
	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"path", "method", "status"},
	)

	// httpRequestsTotal tracks the total number of HTTP requests
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"path", "method", "status"},
	)

	// httpRequestSize tracks the size of HTTP requests
	httpRequestSize = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_size_bytes",
			Help: "Size of HTTP requests in bytes",
		},
		[]string{"path", "method", "status"},
	)

	// httpResponseSize tracks the size of HTTP responses
	httpResponseSize = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_response_size_bytes",
			Help: "Size of HTTP responses in bytes",
		},
		[]string{"path", "method", "status"},
	)
)

// Metrics is a middleware that collects HTTP metrics
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		// Process request
		c.Next()

		// Stop timer
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method

		// Record metrics
		httpRequestDuration.WithLabelValues(path, method, status).Observe(duration)
		httpRequestsTotal.WithLabelValues(path, method, status).Inc()

		// Record request and response size if available
		if c.Request.ContentLength > 0 {
			httpRequestSize.WithLabelValues(path, method, status).Observe(float64(c.Request.ContentLength))
		}

		responseSize := float64(c.Writer.Size())
		if responseSize > 0 {
			httpResponseSize.WithLabelValues(path, method, status).Observe(responseSize)
		}
	}
}