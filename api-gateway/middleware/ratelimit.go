package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter represents a simple rate limiter
type RateLimiter struct {
	clients map[string]*ClientLimiter
	mutex   sync.RWMutex
	rate    int           // requests per minute
	window  time.Duration // time window
}

// ClientLimiter tracks requests for a specific client
type ClientLimiter struct {
	requests  []time.Time
	mutex     sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(requestsPerMinute int) *RateLimiter {
	rl := &RateLimiter{
		clients: make(map[string]*ClientLimiter),
		rate:    requestsPerMinute,
		window:  time.Minute,
	}

	// Clean up old clients every 5 minutes
	go rl.cleanup()

	return rl
}

// RateLimitMiddleware applies rate limiting
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		if !rl.allowRequest(clientIP) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": "60 seconds",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// allowRequest checks if request is allowed for the client
func (rl *RateLimiter) allowRequest(clientIP string) bool {
	rl.mutex.Lock()
	client, exists := rl.clients[clientIP]
	if !exists {
		client = &ClientLimiter{
			requests: make([]time.Time, 0),
		}
		rl.clients[clientIP] = client
	}
	rl.mutex.Unlock()

	client.mutex.Lock()
	defer client.mutex.Unlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Remove old requests outside the time window
	validRequests := make([]time.Time, 0)
	for _, reqTime := range client.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}
	client.requests = validRequests

	// Check if rate limit is exceeded
	if len(client.requests) >= rl.rate {
		return false
	}

	// Add current request
	client.requests = append(client.requests, now)
	return true
}

// cleanup removes old clients that haven't made requests recently
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()
		cutoff := time.Now().Add(-10 * time.Minute)

		for clientIP, client := range rl.clients {
			client.mutex.Lock()
			if len(client.requests) == 0 {
				delete(rl.clients, clientIP)
			} else {
				// Check if last request is too old
				lastRequest := client.requests[len(client.requests)-1]
				if lastRequest.Before(cutoff) {
					delete(rl.clients, clientIP)
				}
			}
			client.mutex.Unlock()
		}
		rl.mutex.Unlock()
	}
}