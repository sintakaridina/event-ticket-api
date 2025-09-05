package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"

	"./handler"
	"./middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Warn("No .env file found")
	}

	// Setup logging
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetLevel(logrus.InfoLevel)

	// Initialize Gin router
	router := gin.New()

	// Initialize proxy handler
	proxyHandler := handler.NewProxyHandler()

	// Initialize rate limiter (100 requests per minute)
	rateLimiter := middleware.NewRateLimiter(100)

	// Middleware
	router.Use(middleware.LoggingMiddleware())
	router.Use(gin.Recovery())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.MetricsMiddleware())
	router.Use(rateLimiter.RateLimitMiddleware())

	// CORS configuration
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"}
	router.Use(cors.New(config))

	// JWT secret key
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-secret-key" // Default for development
		logrus.Warn("JWT_SECRET not set, using default key")
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"timestamp": time.Now().Unix(),
			"service": "api-gateway",
		})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// API routes
	api := router.Group("/api/v1")
	{
		// User service routes (some require auth)
		userGroup := api.Group("/users")
		{
			// Public routes
			userGroup.POST("/register", proxyHandler.ProxyToService("user"))
			userGroup.POST("/login", proxyHandler.ProxyToService("user"))
			
			// Protected routes
			protectedUser := userGroup.Group("")
			protectedUser.Use(middleware.AuthMiddleware(jwtSecret))
			{
				protectedUser.GET("/profile", proxyHandler.ProxyToService("user"))
				protectedUser.PUT("/profile", proxyHandler.ProxyToService("user"))
			}
		}

		// Event service routes
		eventGroup := api.Group("/events")
		{
			// Public routes
			eventGroup.GET("", proxyHandler.ProxyToService("event"))
			eventGroup.GET("/:id", proxyHandler.ProxyToService("event"))
			
			// Protected routes (admin only)
			protectedEvent := eventGroup.Group("")
			protectedEvent.Use(middleware.AuthMiddleware(jwtSecret))
			protectedEvent.Use(middleware.RoleMiddleware("admin", "organizer"))
			{
				protectedEvent.POST("", proxyHandler.ProxyToService("event"))
				protectedEvent.PUT("/:id", proxyHandler.ProxyToService("event"))
				protectedEvent.DELETE("/:id", proxyHandler.ProxyToService("event"))
			}
		}

		// Ticket booking routes (require auth)
		bookingGroup := api.Group("/bookings")
		bookingGroup.Use(middleware.AuthMiddleware(jwtSecret))
		{
			bookingGroup.POST("", proxyHandler.ProxyToService("event"))
			bookingGroup.GET("/:id", proxyHandler.ProxyToService("event"))
			bookingGroup.GET("/user/:userId", proxyHandler.ProxyToService("event"))
			bookingGroup.PUT("/:id/cancel", proxyHandler.ProxyToService("event"))
		}

		// Payment service routes
		paymentGroup := api.Group("/payments")
		{
			// Public webhook
			paymentGroup.POST("/webhook", proxyHandler.ProxyToService("payment"))
			
			// Protected routes
			protectedPayment := paymentGroup.Group("")
			protectedPayment.Use(middleware.AuthMiddleware(jwtSecret))
			{
				protectedPayment.POST("", proxyHandler.ProxyToService("payment"))
				protectedPayment.GET("/:id", proxyHandler.ProxyToService("payment"))
				protectedPayment.POST("/:id/confirm", proxyHandler.ProxyToService("payment"))
			}
		}

		// Notification service routes (admin only)
		notificationGroup := api.Group("/notifications")
		notificationGroup.Use(middleware.AuthMiddleware(jwtSecret))
		notificationGroup.Use(middleware.RoleMiddleware("admin"))
		{
			notificationGroup.POST("/send", proxyHandler.ProxyToService("notification"))
			notificationGroup.GET("/templates", proxyHandler.ProxyToService("notification"))
			notificationGroup.POST("/templates", proxyHandler.ProxyToService("notification"))
		}
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logrus.Infof("API Gateway starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}