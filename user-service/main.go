package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/ticket-system/user-service/config"
	"github.com/yourusername/ticket-system/user-service/handler"
	"github.com/yourusername/ticket-system/user-service/middleware"
	"github.com/yourusername/ticket-system/user-service/repository"
	"github.com/yourusername/ticket-system/user-service/service"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Info("No .env file found, using environment variables")
	}

	// Initialize logger
	logrus.SetFormatter(&logrus.JSONFormatter{})
	logrus.SetOutput(os.Stdout)

	// Set log level based on environment
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to parse log level")
	}
	logrus.SetLevel(level)

	// Initialize database connection
	db, err := config.NewPostgresDB()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to database")
	}

	// Initialize RabbitMQ connection
	rmq, err := config.NewRabbitMQ()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to connect to RabbitMQ")
	}
	defer rmq.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(db)

	// Initialize services
	userService := service.NewUserService(userRepo, rmq)

	// Initialize handlers
	userHandler := handler.NewUserHandler(userService)

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.Metrics())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Setup API routes
	userHandler.SetupRoutes(router, middleware.JWTAuth())

	// Get server port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	// Create HTTP server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logrus.Infof("Starting user service on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := server.Shutdown(ctx); err != nil {
		logrus.WithError(err).Fatal("Server forced to shutdown")
	}

	logrus.Info("Server exiting")
}