package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Logger is a middleware for logging HTTP requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		startTime := time.Now()

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)

		// Get request details
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		path := c.Request.URL.Path
		userAgent := c.Request.UserAgent()

		// Get user ID if available
		userID, exists := c.Get("userID")
		userIDStr := "anonymous"
		if exists {
			userIDStr = userID.(string)
		}

		// Create log entry
		logEntry := logrus.WithFields(logrus.Fields{
			"status_code": statusCode,
			"latency":     latency,
			"client_ip":   clientIP,
			"method":      method,
			"path":        path,
			"user_agent":  userAgent,
			"user_id":     userIDStr,
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