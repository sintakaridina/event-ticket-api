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
	"github.com/yourusername/ticket-system/event-ticket-service/config"
	"github.com/yourusername/ticket-system/event-ticket-service/handler"
	"github.com/yourusername/ticket-system/event-ticket-service/middleware"
	"github.com/yourusername/ticket-system/event-ticket-service/repository"
	"github.com/yourusername/ticket-system/event-ticket-service/service"
)

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		logrus.Warn("Error loading .env file, using environment variables")
	}

	// Initialize logging
	logLevel, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logrus.SetLevel(logLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	logrus.Info("Starting Event & Ticket Service")

	// Connect to PostgreSQL
	db, err := config.NewPostgresDB()
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}
	logrus.Info("Connected to PostgreSQL database")

	// Connect to RabbitMQ
	rmq, err := config.NewRabbitMQ()
	if err != nil {
		logrus.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	logrus.Info("Connected to RabbitMQ")

	// Declare exchanges
	exchanges := []string{"user_events", "ticket_events", "payment_events", "notification_events"}
	for _, exchange := range exchanges {
		err = rmq.DeclareExchange(exchange)
		if err != nil {
			logrus.Fatalf("Failed to declare exchange %s: %v", exchange, err)
		}
	}

	// Initialize repositories
	eventRepo := repository.NewEventRepositoryImpl(db)
	ticketRepo := repository.NewTicketRepositoryImpl(db)
	bookingRepo := repository.NewBookingRepositoryImpl(db)

	// Initialize services
	eventService := service.NewEventService(eventRepo, ticketRepo, rmq)
	bookingService := service.NewBookingService(bookingRepo, eventRepo, ticketRepo, db, rmq)

	// Initialize handlers
	eventHandler := handler.NewEventHandler(eventService)
	bookingHandler := handler.NewBookingHandler(bookingService)

	// Initialize Gin router
	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(middleware.Metrics())

	// Set up metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Set up health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Set up routes
	eventHandler.SetupRoutes(router, middleware.JWTAuth())
	bookingHandler.SetupRoutes(router, middleware.JWTAuth())

	// Set up consumer for payment events
	go func() {
		queueName := "event_ticket_payment_events"
		routingKey := "payment.#"

		err := rmq.ConsumeMessages("payment_events", queueName, routingKey, func(message []byte) error {
			logrus.Infof("Received payment event: %s", string(message))
			// Process payment event
			// TODO: Implement payment event processing
			return nil
		})

		if err != nil {
			logrus.Errorf("Failed to consume payment events: %v", err)
		}
	}()

	// Set up consumer for user events
	go func() {
		queueName := "event_ticket_user_events"
		routingKey := "user.#"

		err := rmq.ConsumeMessages("user_events", queueName, routingKey, func(message []byte) error {
			logrus.Infof("Received user event: %s", string(message))
			// Process user event
			// TODO: Implement user event processing
			return nil
		})

		if err != nil {
			logrus.Errorf("Failed to consume user events: %v", err)
		}
	}()

	// Start HTTP server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8081"
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logrus.Infof("Server listening on port %s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Shutdown server
	if err := server.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server forced to shutdown: %v", err)
	}

	// Close RabbitMQ connection
	rmq.Close()

	logrus.Info("Server exited properly")
}