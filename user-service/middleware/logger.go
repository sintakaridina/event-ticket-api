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

		// Stop timer
		timestamp := time.Now()
		latency := timestamp.Sub(start)

		// Get client IP
		clientIP := c.ClientIP()

		// Get method and status code
		method := c.Request.Method
		statusCode := c.Writer.Status()

		// Get error if any
		error := c.Errors.ByType(gin.ErrorTypePrivate).String()

		// Append query string if any
		if raw != "" {
			path = path + "?" + raw
		}

		// Log request details
		logEntry := logrus.WithFields(logrus.Fields{
			"status_code": statusCode,
			"latency":     latency,
			"client_ip":   clientIP,
			"method":      method,
			"path":        path,
			"error":       error,
		})

		// Log based on status code
		if statusCode >= 500 {
			logEntry.Error("Server error")
		} else if statusCode >= 400 {
			logEntry.Warn("Client error")
		} else {
			logEntry.Info("Request processed")
		}
	}
}