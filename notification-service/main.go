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
	"github.com/streadway/amqp"

	"notification-service/config"
	"notification-service/handler"
	"notification-service/middleware"
	"notification-service/model"
	"notification-service/provider"
	"notification-service/repository"
	"notification-service/service"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		logrus.Warn("Error loading .env file, using environment variables")
	}

	// Initialize logging
	initLogging()

	// Connect to PostgreSQL
	db, err := config.NewPostgresDB()
	if err != nil {
		logrus.Fatalf("Failed to connect to database: %v", err)
	}
	logrus.Info("Connected to PostgreSQL database")

	// Connect to RabbitMQ
	rabbitMQ, err := config.NewRabbitMQ()
	if err != nil {
		logrus.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	logrus.Info("Connected to RabbitMQ")

	// Declare exchanges
	exchanges := []string{"notification_events", "payment_events", "ticket_events", "user_events"}
	for _, exchange := range exchanges {
		if err := rabbitMQ.DeclareExchange(exchange); err != nil {
			logrus.Fatalf("Failed to declare exchange %s: %v", exchange, err)
		}
	}

	// Initialize repositories
	repo := repository.NewNotificationRepositoryImpl(db)

	// Run database migrations
	if err := repo.AutoMigrate(); err != nil {
		logrus.Fatalf("Failed to run database migrations: %v", err)
	}
	logrus.Info("Database migrations completed successfully")

	// Initialize providers
	emailProvider := provider.NewEmailProvider()
	smsProvider := provider.NewSMSProvider()
	pushProvider := provider.NewPushProvider()

	// Initialize services
	notificationService := service.NewNotificationService(repo, rabbitMQ, emailProvider, smsProvider, pushProvider)

	// Initialize handlers
	notificationHandler := handler.NewNotificationHandler(notificationService)

	// Initialize Gin router
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.Metrics())

	// Initialize Prometheus metrics
	middleware.InitMetrics()

	// Set up routes
	notificationHandler.SetupRoutes(router)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Set up RabbitMQ consumers
	setupConsumers(rabbitMQ, notificationService)

	// Start HTTP server
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8083"
	}

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		logrus.Infof("Starting server on port %s", port)
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
	if err := rabbitMQ.Close(); err != nil {
		logrus.Errorf("Failed to close RabbitMQ connection: %v", err)
	}

	// Close database connection
	if err := db.Close(); err != nil {
		logrus.Errorf("Failed to close database connection: %v", err)
	}

	logrus.Info("Server exited properly")
}

// initLogging initializes the logging configuration
func initLogging() {
	// Set log level
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		logrus.Warnf("Invalid log level %s, defaulting to info", logLevel)
		level = logrus.InfoLevel
	}

	logrus.SetLevel(level)

	// Set log format
	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp: true,
		})
	}

	logrus.Info("Logging initialized")
}

// setupConsumers sets up RabbitMQ consumers
func setupConsumers(rabbitMQ *config.RabbitMQ, notificationService service.NotificationService) {
	// Payment events consumer
	go func() {
		queueName := "notification_payment_events"
		if err := rabbitMQ.DeclareQueue(queueName); err != nil {
			logrus.Fatalf("Failed to declare queue %s: %v", queueName, err)
		}

		if err := rabbitMQ.BindQueue(queueName, "payment_events", ""); err != nil {
			logrus.Fatalf("Failed to bind queue %s to exchange: %v", queueName, err)
		}

		messages, err := rabbitMQ.ConsumeMessages(queueName)
		if err != nil {
			logrus.Fatalf("Failed to consume messages from queue %s: %v", queueName, err)
		}

		logrus.Infof("Started consuming payment events from queue %s", queueName)

		for msg := range messages {
			logrus.Debugf("Received payment event: %s", string(msg.Body))
			if err := notificationService.HandlePaymentEvent(msg.Body); err != nil {
				logrus.Errorf("Failed to handle payment event: %v", err)
				// Nack the message to requeue it
				if err := msg.Nack(false, true); err != nil {
					logrus.Errorf("Failed to nack message: %v", err)
				}
			} else {
				// Ack the message
				if err := msg.Ack(false); err != nil {
					logrus.Errorf("Failed to ack message: %v", err)
				}
			}
		}
	}()

	// Ticket events consumer
	go func() {
		queueName := "notification_ticket_events"
		if err := rabbitMQ.DeclareQueue(queueName); err != nil {
			logrus.Fatalf("Failed to declare queue %s: %v", queueName, err)
		}

		if err := rabbitMQ.BindQueue(queueName, "ticket_events", ""); err != nil {
			logrus.Fatalf("Failed to bind queue %s to exchange: %v", queueName, err)
		}

		messages, err := rabbitMQ.ConsumeMessages(queueName)
		if err != nil {
			logrus.Fatalf("Failed to consume messages from queue %s: %v", queueName, err)
		}

		logrus.Infof("Started consuming ticket events from queue %s", queueName)

		for msg := range messages {
			logrus.Debugf("Received ticket event: %s", string(msg.Body))
			if err := notificationService.HandleTicketEvent(msg.Body); err != nil {
				logrus.Errorf("Failed to handle ticket event: %v", err)
				// Nack the message to requeue it
				if err := msg.Nack(false, true); err != nil {
					logrus.Errorf("Failed to nack message: %v", err)
				}
			} else {
				// Ack the message
				if err := msg.Ack(false); err != nil {
					logrus.Errorf("Failed to ack message: %v", err)
				}
			}
		}
	}()

	// User events consumer
	go func() {
		queueName := "notification_user_events"
		if err := rabbitMQ.DeclareQueue(queueName); err != nil {
			logrus.Fatalf("Failed to declare queue %s: %v", queueName, err)
		}

		if err := rabbitMQ.BindQueue(queueName, "user_events", ""); err != nil {
			logrus.Fatalf("Failed to bind queue %s to exchange: %v", queueName, err)
		}

		messages, err := rabbitMQ.ConsumeMessages(queueName)
		if err != nil {
			logrus.Fatalf("Failed to consume messages from queue %s: %v", queueName, err)
		}

		logrus.Infof("Started consuming user events from queue %s", queueName)

		for msg := range messages {
			logrus.Debugf("Received user event: %s", string(msg.Body))
			if err := notificationService.HandleUserEvent(msg.Body); err != nil {
				logrus.Errorf("Failed to handle user event: %v", err)
				// Nack the message to requeue it
				if err := msg.Nack(false, true); err != nil {
					logrus.Errorf("Failed to nack message: %v", err)
				}
			} else {
				// Ack the message
				if err := msg.Ack(false); err != nil {
					logrus.Errorf("Failed to ack message: %v", err)
				}
			}
		}
	}()

	// Create default templates
	createDefaultTemplates(notificationService)
}

// createDefaultTemplates creates default notification templates
func createDefaultTemplates(notificationService service.NotificationService) {
	templates := []model.CreateTemplateRequest{
		{
			Code:        "welcome_email",
			Title:       "Welcome to Event Ticket Platform",
			Content:     "<h1>Welcome, {{username}}!</h1><p>Thank you for joining our event ticket platform. We're excited to have you on board!</p>",
			Description: "Welcome email for new users",
		},
		{
			Code:        "password_reset",
			Title:       "Password Reset Request",
			Content:     "<h1>Password Reset</h1><p>You requested a password reset. Click the link below to reset your password:</p><p><a href='{{reset_url}}'>Reset Password</a></p>",
			Description: "Password reset email",
		},
		{
			Code:        "payment_success",
			Title:       "Payment Successful",
			Content:     "<h1>Payment Successful</h1><p>Your payment of {{amount}} for booking {{booking_id}} has been successfully processed.</p>",
			Description: "Payment success notification",
		},
		{
			Code:        "payment_failed",
			Title:       "Payment Failed",
			Content:     "<h1>Payment Failed</h1><p>Your payment for booking {{booking_id}} has failed. Reason: {{reason}}</p>",
			Description: "Payment failure notification",
		},
		{
			Code:        "payment_refunded",
			Title:       "Payment Refunded",
			Content:     "<h1>Payment Refunded</h1><p>Your payment of {{amount}} for booking {{booking_id}} has been refunded.</p>",
			Description: "Payment refund notification",
		},
		{
			Code:        "ticket_booked",
			Title:       "Ticket Booking Confirmed",
			Content:     "<h1>Booking Confirmed</h1><p>Your booking for {{event_name}} on {{event_date}} at {{event_location}} has been confirmed. You have booked {{ticket_count}} ticket(s).</p>",
			Description: "Ticket booking confirmation",
		},
		{
			Code:        "ticket_cancelled",
			Title:       "Ticket Booking Cancelled",
			Content:     "<h1>Booking Cancelled</h1><p>Your booking {{booking_id}} for {{event_name}} has been cancelled. Reason: {{reason}}</p>",
			Description: "Ticket cancellation notification",
		},
		{
			Code:        "event_reminder",
			Title:       "Event Reminder",
			Content:     "<h1>Event Reminder</h1><p>This is a reminder that {{event_name}} is scheduled for {{event_date}} at {{event_location}}. We look forward to seeing you there!</p>",
			Description: "Event reminder notification",
		},
		{
			Code:        "event_reminder_sms",
			Title:       "Event Reminder",
			Content:     "Reminder: {{event_name}} is on {{event_date}} at {{event_location}}. See you there!",
			Description: "Event reminder SMS",
		},
	}

	for _, template := range templates {
		// Check if template already exists
		existing, err := notificationService.GetTemplateByCode(template.Code)
		if err == nil && existing != nil {
			logrus.Infof("Template %s already exists, skipping", template.Code)
			continue
		}

		// Create template
		_, err := notificationService.CreateTemplate(template)
		if err != nil {
			logrus.Errorf("Failed to create template %s: %v", template.Code, err)
		} else {
			logrus.Infof("Created default template: %s", template.Code)
		}
	}
}