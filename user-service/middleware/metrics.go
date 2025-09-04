package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics variables
var (
	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	requestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	requestSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_request_size_bytes",
			Help: "Size of HTTP requests in bytes",
		},
		[]string{"method", "path", "status"},
	)

	responseSize = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Name: "http_response_size_bytes",
			Help: "Size of HTTP responses in bytes",
		},
		[]string{"method", "path", "status"},
	)
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(requestDuration, requestTotal, requestSize, responseSize)
}

// Metrics is a middleware that collects HTTP metrics
func Metrics() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Collect metrics after request is processed
		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		method := c.Request.Method
		path := c.FullPath()
		if path == "" {
			path = "unknown"
		}

		// Record metrics
		requestDuration.WithLabelValues(method, path, status).Observe(elapsed)
		requestTotal.WithLabelValues(method, path, status).Inc()

		// Record request and response sizes if available
		if c.Request.ContentLength > 0 {
			requestSize.WithLabelValues(method, path, status).Observe(float64(c.Request.ContentLength))
		}
		responseSize.WithLabelValues(method, path, status).Observe(float64(c.Writer.Size()))
	}
}