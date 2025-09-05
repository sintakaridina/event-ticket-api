package middleware

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
			Name: "api_gateway_http_requests_total",
			Help: "Total number of HTTP requests processed by the API Gateway",
		},
		[]string{"method", "path", "status_code", "service"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_gateway_http_request_duration_seconds",
			Help:    "Duration of HTTP requests processed by the API Gateway",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "service"},
	)

	httpRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_gateway_http_request_size_bytes",
			Help:    "Size of HTTP requests processed by the API Gateway",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "path", "service"},
	)

	httpResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "api_gateway_http_response_size_bytes",
			Help:    "Size of HTTP responses from the API Gateway",
			Buckets: []float64{100, 1000, 10000, 100000, 1000000},
		},
		[]string{"method", "path", "service"},
	)

	// Service health metrics
	serviceHealthStatus = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "api_gateway_service_health_status",
			Help: "Health status of backend services (1 = healthy, 0 = unhealthy)",
		},
		[]string{"service"},
	)

	// Rate limiting metrics
	rateLimitExceeded = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_gateway_rate_limit_exceeded_total",
			Help: "Total number of requests that exceeded rate limits",
		},
		[]string{"client_ip"},
	)
)

// MetricsMiddleware collects HTTP metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Determine target service from path
		service := getServiceFromPath(path)

		// Record request size
		if c.Request.ContentLength > 0 {
			httpRequestSize.WithLabelValues(method, path, service).Observe(float64(c.Request.ContentLength))
		}

		c.Next()

		// Record metrics after request processing
		duration := time.Since(start).Seconds()
		statusCode := strconv.Itoa(c.Writer.Status())

		httpRequestsTotal.WithLabelValues(method, path, statusCode, service).Inc()
		httpRequestDuration.WithLabelValues(method, path, service).Observe(duration)

		// Record response size
		responseSize := c.Writer.Size()
		if responseSize > 0 {
			httpResponseSize.WithLabelValues(method, path, service).Observe(float64(responseSize))
		}
	}
}

// getServiceFromPath determines which service the request is targeting
func getServiceFromPath(path string) string {
	if len(path) < 8 { // "/api/v1/" is 8 characters
		return "unknown"
	}

	// Remove "/api/v1/" prefix
	if path[:8] == "/api/v1/" {
		path = path[8:]
	}

	// Determine service based on path prefix
	switch {
	case len(path) >= 5 && path[:5] == "users":
		return "user-service"
	case len(path) >= 6 && path[:6] == "events":
		return "event-ticket-service"
	case len(path) >= 8 && path[:8] == "bookings":
		return "event-ticket-service"
	case len(path) >= 8 && path[:8] == "payments":
		return "payment-service"
	case len(path) >= 13 && path[:13] == "notifications":
		return "notification-service"
	default:
		return "unknown"
	}
}

// UpdateServiceHealth updates the health status of a service
func UpdateServiceHealth(serviceName string, healthy bool) {
	status := 0.0
	if healthy {
		status = 1.0
	}
	serviceHealthStatus.WithLabelValues(serviceName).Set(status)
}

// RecordRateLimitExceeded records when rate limit is exceeded
func RecordRateLimitExceeded(clientIP string) {
	rateLimitExceeded.WithLabelValues(clientIP).Inc()
}