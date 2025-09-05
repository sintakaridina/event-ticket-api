package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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
			
			// Parse payment event
			var paymentEvent map[string]interface{}
			if err := json.Unmarshal(message, &paymentEvent); err != nil {
				logrus.WithError(err).Error("Failed to parse payment event")
				return err
			}
			
			// Process based on event type
			eventType, ok := paymentEvent["event_type"].(string)
			if !ok {
				logrus.Error("Invalid event type in payment event")
				return fmt.Errorf("invalid event type")
			}
			
			switch eventType {
			case "payment.completed":
				return handlePaymentCompleted(paymentEvent, bookingService)
			case "payment.failed":
				return handlePaymentFailed(paymentEvent, bookingService)
			case "payment.refunded":
				return handlePaymentRefunded(paymentEvent, bookingService)
			default:
				logrus.Warnf("Unknown payment event type: %s", eventType)
				return nil
			}
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
			
			// Parse user event
			var userEvent map[string]interface{}
			if err := json.Unmarshal(message, &userEvent); err != nil {
				logrus.WithError(err).Error("Failed to parse user event")
				return err
			}
			
			// Process based on event type
			eventType, ok := userEvent["event_type"].(string)
			if !ok {
				logrus.Error("Invalid event type in user event")
				return fmt.Errorf("invalid event type")
			}
			
			switch eventType {
			case "user.deleted":
				return handleUserDeleted(userEvent, bookingService)
			case "user.suspended":
				return handleUserSuspended(userEvent, bookingService)
			default:
				logrus.Warnf("Unknown user event type: %s", eventType)
				return nil
			}
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

// handlePaymentCompleted handles payment completed events
func handlePaymentCompleted(paymentEvent map[string]interface{}, bookingService service.BookingService) error {
	// Extract booking ID from payment event
	bookingIDStr, ok := paymentEvent["booking_id"].(string)
	if !ok {
		logrus.Error("Missing booking_id in payment completed event")
		return fmt.Errorf("missing booking_id")
	}

	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		logrus.WithError(err).Error("Invalid booking_id in payment completed event")
		return err
	}

	// Update booking status to confirmed
	err = bookingService.UpdateBookingStatus(bookingID, "confirmed")
	if err != nil {
		logrus.WithError(err).Errorf("Failed to confirm booking %s", bookingID)
		return err
	}

	logrus.Infof("Booking %s confirmed after payment completion", bookingID)
	return nil
}

// handlePaymentFailed handles payment failed events
func handlePaymentFailed(paymentEvent map[string]interface{}, bookingService service.BookingService) error {
	// Extract booking ID from payment event
	bookingIDStr, ok := paymentEvent["booking_id"].(string)
	if !ok {
		logrus.Error("Missing booking_id in payment failed event")
		return fmt.Errorf("missing booking_id")
	}

	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		logrus.WithError(err).Error("Invalid booking_id in payment failed event")
		return err
	}

	// Cancel booking due to payment failure
	err = bookingService.CancelBooking(bookingID)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to cancel booking %s after payment failure", bookingID)
		return err
	}

	logrus.Infof("Booking %s cancelled after payment failure", bookingID)
	return nil
}

// handlePaymentRefunded handles payment refunded events
func handlePaymentRefunded(paymentEvent map[string]interface{}, bookingService service.BookingService) error {
	// Extract booking ID from payment event
	bookingIDStr, ok := paymentEvent["booking_id"].(string)
	if !ok {
		logrus.Error("Missing booking_id in payment refunded event")
		return fmt.Errorf("missing booking_id")
	}

	bookingID, err := uuid.Parse(bookingIDStr)
	if err != nil {
		logrus.WithError(err).Error("Invalid booking_id in payment refunded event")
		return err
	}

	// Update booking status to refunded
	err = bookingService.UpdateBookingStatus(bookingID, "refunded")
	if err != nil {
		logrus.WithError(err).Errorf("Failed to update booking %s to refunded", bookingID)
		return err
	}

	logrus.Infof("Booking %s marked as refunded", bookingID)
	return nil
}

// handleUserDeleted handles user deleted events
func handleUserDeleted(userEvent map[string]interface{}, bookingService service.BookingService) error {
	// Extract user ID from user event
	userIDStr, ok := userEvent["user_id"].(string)
	if !ok {
		logrus.Error("Missing user_id in user deleted event")
		return fmt.Errorf("missing user_id")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.WithError(err).Error("Invalid user_id in user deleted event")
		return err
	}

	// Get all pending bookings for the user
	bookings, err := bookingService.GetBookingsByUserID(userID)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to get bookings for deleted user %s", userID)
		return err
	}

	// Cancel all pending bookings
	for _, booking := range bookings {
		if booking.Status == "pending" || booking.Status == "confirmed" {
			err = bookingService.CancelBooking(booking.ID)
			if err != nil {
				logrus.WithError(err).Errorf("Failed to cancel booking %s for deleted user %s", booking.ID, userID)
				continue
			}
			logrus.Infof("Cancelled booking %s for deleted user %s", booking.ID, userID)
		}
	}

	return nil
}

// handleUserSuspended handles user suspended events
func handleUserSuspended(userEvent map[string]interface{}, bookingService service.BookingService) error {
	// Extract user ID from user event
	userIDStr, ok := userEvent["user_id"].(string)
	if !ok {
		logrus.Error("Missing user_id in user suspended event")
		return fmt.Errorf("missing user_id")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		logrus.WithError(err).Error("Invalid user_id in user suspended event")
		return err
	}

	// Get all pending bookings for the user
	bookings, err := bookingService.GetBookingsByUserID(userID)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to get bookings for suspended user %s", userID)
		return err
	}

	// Cancel all pending bookings
	for _, booking := range bookings {
		if booking.Status == "pending" {
			err = bookingService.CancelBooking(booking.ID)
			if err != nil {
				logrus.WithError(err).Errorf("Failed to cancel booking %s for suspended user %s", booking.ID, userID)
				continue
			}
			logrus.Infof("Cancelled pending booking %s for suspended user %s", booking.ID, userID)
		}
	}

	return nil
}