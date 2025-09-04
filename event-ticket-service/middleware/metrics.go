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
		[]string{"method", "path", "status"},
	)

	// httpRequestsTotal tracks the total number of HTTP requests
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// httpRequestSize tracks the size of HTTP requests
	httpRequestSize = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_size_bytes",
			Help: "Size of HTTP requests in bytes",
		},
		[]string{"method", "path"},
	)

	// httpResponseSize tracks the size of HTTP responses
	httpResponseSize = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_response_size_bytes",
			Help: "Size of HTTP responses in bytes",
		},
		[]string{"method", "path"},
	)
)

// Metrics is a middleware that collects HTTP metrics
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Get request size
		requestSize := 0
		if c.Request.ContentLength > 0 {
			requestSize = int(c.Request.ContentLength)
		}

		// Create response writer wrapper to capture response size
		writer := &responseWriterWrapper{ResponseWriter: c.Writer}
		c.Writer = writer

		// Process request
		c.Next()

		// Get path and method
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}
		method := c.Request.Method

		// Get status code
		status := strconv.Itoa(c.Writer.Status())

		// Record metrics
		elapsed := time.Since(start).Seconds()
		httpRequestDuration.WithLabelValues(method, path, status).Observe(elapsed)
		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestSize.WithLabelValues(method, path).Observe(float64(requestSize))
		httpResponseSize.WithLabelValues(method, path).Observe(float64(writer.size))
	}
}

// responseWriterWrapper wraps a gin.ResponseWriter to capture the response size
type responseWriterWrapper struct {
	gin.ResponseWriter
	size int
}

// Write captures the response size
func (w *responseWriterWrapper) Write(data []byte) (int, error) {
	size, err := w.ResponseWriter.Write(data)
	w.size += size
	return size, err
}

// WriteString captures the response size
func (w *responseWriterWrapper) WriteString(s string) (int, error) {
	size, err := w.ResponseWriter.WriteString(s)
	w.size += size
	return size, err
}