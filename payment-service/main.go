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
	"github.com/yourusername/ticket-system/payment-service/config"
	"github.com/yourusername/ticket-system/payment-service/handler"
	"github.com/yourusername/ticket-system/payment-service/middleware"
	"github.com/yourusername/ticket-system/payment-service/provider"
	"github.com/yourusername/ticket-system/payment-service/repository"
	"github.com/yourusername/ticket-system/payment-service/service"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Warn("Error loading .env file, using environment variables")
	}

	// Initialize logger
	logLevel, err := logrus.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	logrus.SetLevel(logLevel)
	logrus.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	logrus.Info("Starting Payment Service")

	// Connect to PostgreSQL
	db, err := config.NewPostgresDB()
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Connect to RabbitMQ
	ramqConn, err := config.NewRabbitMQConnection()
	if err != nil {
		logrus.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer ramqConn.Close()

	// Declare exchanges
	exchanges := []string{"payment_events", "ticket_events", "notification_events"}
	for _, exchange := range exchanges {
		if err := config.DeclareExchange(ramqConn, exchange); err != nil {
			logrus.Fatalf("Failed to declare exchange %s: %v", exchange, err)
		}
	}

	// Initialize repositories
	paymentRepo := repository.NewPaymentRepository(db)

	// Initialize payment provider
	paymentProvider := provider.NewPaymentProvider()

	// Initialize services
	paymentService := service.NewPaymentService(paymentRepo, paymentProvider, ramqConn)

	// Initialize handlers
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// Initialize Gin router
	router := gin.New()
	router.Use(middleware.Logger())
	router.Use(middleware.Metrics())
	router.Use(gin.Recovery())

	// Set up routes
	paymentHandler.SetupRoutes(router, middleware.JWTAuth())

	// Set up Prometheus metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Set up health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})

	// Set up consumers
	go setupConsumers(ramqConn, paymentService)

	// Start server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8082"
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logrus.Infof("Server listening on port %s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logrus.Info("Shutting down server...")

	// Create a deadline for server shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Shutdown server
	if err := srv.Shutdown(ctx); err != nil {
		logrus.Fatalf("Server forced to shutdown: %v", err)
	}

	logrus.Info("Server exiting")
}

// setupConsumers sets up RabbitMQ consumers
func setupConsumers(conn *config.RabbitMQConnection, paymentService service.PaymentService) {
	// Consume booking events
	go func() {
		queueName := "payment_service_booking_events"
		exchangeName := "ticket_events"
		routingKey := "booking.#"

		ch, err := conn.Connection.Channel()
		if err != nil {
			logrus.Fatalf("Failed to open a channel: %v", err)
		}
		defer ch.Close()

		// Declare a queue
		q, err := ch.QueueDeclare(
			queueName, // name
			true,     // durable
			false,    // delete when unused
			false,    // exclusive
			false,    // no-wait
			nil,      // arguments
		)
		if err != nil {
			logrus.Fatalf("Failed to declare a queue: %v", err)
		}

		// Bind the queue to the exchange
		err = ch.QueueBind(
			q.Name,       // queue name
			routingKey,   // routing key
			exchangeName, // exchange
			false,        // no-wait
			nil,          // arguments
		)
		if err != nil {
			logrus.Fatalf("Failed to bind a queue: %v", err)
		}

		// Consume messages
		msgs, err := ch.Consume(
			q.Name, // queue
			"",     // consumer
			false,  // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		if err != nil {
			logrus.Fatalf("Failed to register a consumer: %v", err)
		}

		logrus.Infof("Started consuming messages from queue: %s", queueName)

		// Process messages
		for msg := range msgs {
			logrus.Infof("Received a message: %s", msg.RoutingKey)

			// Process the message based on the routing key
			switch msg.RoutingKey {
			case "booking.created":
				// Handle booking created event
				// This could trigger payment creation
				logrus.Info("Processing booking.created event")
				// paymentService.HandleBookingCreatedEvent(msg.Body)

			case "booking.cancelled":
				// Handle booking cancelled event
				// This could trigger payment refund
				logrus.Info("Processing booking.cancelled event")
				// paymentService.HandleBookingCancelledEvent(msg.Body)

			default:
				logrus.Warnf("Unknown routing key: %s", msg.RoutingKey)
			}

			// Acknowledge the message
			msg.Ack(false)
		}
	}()

	// Consume user events
	go func() {
		queueName := "payment_service_user_events"
		exchangeName := "user_events"
		routingKey := "user.#"

		ch, err := conn.Connection.Channel()
		if err != nil {
			logrus.Fatalf("Failed to open a channel: %v", err)
		}
		defer ch.Close()

		// Declare a queue
		q, err := ch.QueueDeclare(
			queueName, // name
			true,     // durable
			false,    // delete when unused
			false,    // exclusive
			false,    // no-wait
			nil,      // arguments
		)
		if err != nil {
			logrus.Fatalf("Failed to declare a queue: %v", err)
		}

		// Bind the queue to the exchange
		err = ch.QueueBind(
			q.Name,       // queue name
			routingKey,   // routing key
			exchangeName, // exchange
			false,        // no-wait
			nil,          // arguments
		)
		if err != nil {
			logrus.Fatalf("Failed to bind a queue: %v", err)
		}

		// Consume messages
		msgs, err := ch.Consume(
			q.Name, // queue
			"",     // consumer
			false,  // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		if err != nil {
			logrus.Fatalf("Failed to register a consumer: %v", err)
		}

		logrus.Infof("Started consuming messages from queue: %s", queueName)

		// Process messages
		for msg := range msgs {
			logrus.Infof("Received a message: %s", msg.RoutingKey)

			// Process the message based on the routing key
			switch msg.RoutingKey {
			case "user.deleted":
				// Handle user deleted event
				// This could trigger payment data anonymization
				logrus.Info("Processing user.deleted event")
				// paymentService.HandleUserDeletedEvent(msg.Body)

			default:
				logrus.Warnf("Unknown routing key: %s", msg.RoutingKey)
			}

			// Acknowledge the message
			msg.Ack(false)
		}
	}()
}