package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger is a middleware that logs request details
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Get status code
		statusCode := c.Writer.Status()

		// Get client IP
		clientIP := c.ClientIP()

		// Get method
		method := c.Request.Method

		// Get path with query string
		if raw != "" {
			path = path + "?" + raw
		}

		// Get error if any
		err, _ := c.Get("error")

		// Create log entry
		entry := logrus.WithFields(logrus.Fields{
			"status_code": statusCode,
			"latency":     latency,
			"client_ip":   clientIP,
			"method":      method,
			"path":        path,
			"user_agent":  c.Request.UserAgent(),
		})

		// Log based on status code
		switch {
		case statusCode >= 500:
			entry.Error(err)
		case statusCode >= 400:
			entry.Warn(err)
		case statusCode >= 300:
			entry.Info("Redirected")
		default:
			entry.Info("Request processed")
		}
	}
}