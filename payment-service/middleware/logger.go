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
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		size := c.Writer.Size()
		if raw != "" {
			path = path + "?" + raw
		}

		// Get user ID from context if available
		userID, exists := c.Get("userID")
		userIDStr := "anonymous"
		if exists {
			userIDStr = userID.(string)
		}

		// Create log entry
		entry := logrus.WithFields(logrus.Fields{
			"status_code": statusCode,
			"latency":     latency,
			"client_ip":   clientIP,
			"method":      method,
			"path":        path,
			"user_agent":  c.Request.UserAgent(),
			"user_id":     userIDStr,
			"size":        size,
		})

		// Log based on status code
		if statusCode >= 500 {
			entry.Error("Server error")
		} else if statusCode >= 400 {
			entry.Warn("Client error")
		} else {
			entry.Info("Request processed")
		}
	}
}