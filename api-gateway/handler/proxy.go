package handler

import (
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ServiceConfig holds configuration for each microservice
type ServiceConfig struct {
	Name string
	Port string
	Host string
}

// ProxyHandler handles proxy requests to microservices
type ProxyHandler struct {
	services map[string]ServiceConfig
	client   *http.Client
}

// NewProxyHandler creates a new proxy handler
func NewProxyHandler() *ProxyHandler {
	return &ProxyHandler{
		services: map[string]ServiceConfig{
			"user": {
				Name: "user-service",
				Port: "8081",
				Host: getEnvOrDefault("USER_SERVICE_HOST", "localhost"),
			},
			"event": {
				Name: "event-ticket-service",
				Port: "8082",
				Host: getEnvOrDefault("EVENT_TICKET_SERVICE_HOST", "localhost"),
			},
			"payment": {
				Name: "payment-service",
				Port: "8083",
				Host: getEnvOrDefault("PAYMENT_SERVICE_HOST", "localhost"),
			},
			"notification": {
				Name: "notification-service",
				Port: "8084",
				Host: getEnvOrDefault("NOTIFICATION_SERVICE_HOST", "localhost"),
			},
		},
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ProxyToService proxies request to specified service
func (p *ProxyHandler) ProxyToService(serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		service, exists := p.services[serviceName]
		if !exists {
			logrus.Errorf("Service %s not found", serviceName)
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}

		p.proxyRequest(c, service)
	}
}

// proxyRequest handles the actual proxy logic
func (p *ProxyHandler) proxyRequest(c *gin.Context, service ServiceConfig) {
	// Build target URL
	targetURL := "http://" + service.Host + ":" + service.Port + c.Request.URL.Path
	if c.Request.URL.RawQuery != "" {
		targetURL += "?" + c.Request.URL.RawQuery
	}

	logrus.Infof("Proxying %s %s to %s", c.Request.Method, c.Request.URL.Path, targetURL)

	// Create new request
	req, err := http.NewRequest(c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		logrus.Errorf("Failed to create request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}

	// Copy headers
	for key, values := range c.Request.Header {
		// Skip hop-by-hop headers
		if isHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Add X-Forwarded headers
	req.Header.Set("X-Forwarded-For", c.ClientIP())
	req.Header.Set("X-Forwarded-Proto", "http")
	req.Header.Set("X-Forwarded-Host", c.Request.Host)

	// Make request
	resp, err := p.client.Do(req)
	if err != nil {
		logrus.Errorf("Failed to proxy request to %s: %v", service.Name, err)
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error": "Service unavailable",
			"service": service.Name,
		})
		return
	}
	defer resp.Body.Close()

	// Copy response headers
	for key, values := range resp.Header {
		// Skip hop-by-hop headers
		if isHopByHopHeader(key) {
			continue
		}
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Set status code and copy body
	c.Status(resp.StatusCode)
	c.Stream(func(w io.Writer) bool {
		_, err := io.Copy(w, resp.Body)
		return err == nil
	})
}

// getEnvOrDefault returns environment variable value or default
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// isHopByHopHeader checks if header should not be forwarded
func isHopByHopHeader(header string) bool {
	hopByHopHeaders := []string{
		"Connection",
		"Keep-Alive",
		"Proxy-Authenticate",
		"Proxy-Authorization",
		"Te",
		"Trailers",
		"Transfer-Encoding",
		"Upgrade",
	}

	for _, h := range hopByHopHeaders {
		if header == h {
			return true
		}
	}
	return false
}