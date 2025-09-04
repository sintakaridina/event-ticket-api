package service

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"github.com/yourusername/ticket-system/payment-service/config"
	"github.com/yourusername/ticket-system/payment-service/model"
	"github.com/yourusername/ticket-system/payment-service/repository"
	"github.com/yourusername/ticket-system/payment-service/provider"
)

// PaymentService defines the interface for payment service operations
type PaymentService interface {
	CreatePayment(userID uuid.UUID, req model.CreatePaymentRequest) (*model.PaymentResponse, error)
	ProcessPayment(userID uuid.UUID, req model.ProcessPaymentRequest) (*model.PaymentResponse, error)
	GetPaymentByID(id uuid.UUID) (*model.PaymentResponse, error)
	GetPaymentByBookingID(bookingID uuid.UUID) (*model.PaymentResponse, error)
	GetUserPayments(userID uuid.UUID, page, pageSize int) ([]model.PaymentResponse, int64, error)
	UpdatePaymentStatus(id uuid.UUID, req model.UpdatePaymentStatusRequest) (*model.PaymentResponse, error)
	RefundPayment(id uuid.UUID, req model.RefundRequest) (*model.PaymentResponse, error)
}

// paymentService implements PaymentService interface
type paymentService struct {
	paymentRepo    repository.PaymentRepository
	paymentProvider provider.PaymentProvider
	rmq            *config.RabbitMQ
}

// NewPaymentService creates a new payment service
func NewPaymentService(
	paymentRepo repository.PaymentRepository,
	paymentProvider provider.PaymentProvider,
	rmq *config.RabbitMQ,
) PaymentService {
	return &paymentService{
		paymentRepo:    paymentRepo,
		paymentProvider: paymentProvider,
		rmq:            rmq,
	}
}

// CreatePayment creates a new payment
func (s *paymentService) CreatePayment(userID uuid.UUID, req model.CreatePaymentRequest) (*model.PaymentResponse, error) {
	// Check if payment already exists for booking
	existingPayment, err := s.paymentRepo.FindByBookingID(req.BookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing payment: %w", err)
	}

	if existingPayment != nil {
		return nil, fmt.Errorf("payment already exists for this booking")
	}

	// Create payment
	payment := &model.Payment{
		UserID:        userID,
		BookingID:     req.BookingID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		Status:        "pending",
		PaymentMethod: req.PaymentMethod,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Save payment to database
	if err := s.paymentRepo.Create(payment); err != nil {
		return nil, fmt.Errorf("failed to create payment: %w", err)
	}

	// Publish payment created event
	s.publishPaymentEvent("payment.created", payment)

	// Return payment response
	paymentResponse := payment.ToResponse()
	return &paymentResponse, nil
}

// ProcessPayment processes a payment
func (s *paymentService) ProcessPayment(userID uuid.UUID, req model.ProcessPaymentRequest) (*model.PaymentResponse, error) {
	// Check if payment already exists for booking
	existingPayment, err := s.paymentRepo.FindByBookingID(req.BookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing payment: %w", err)
	}

	var payment *model.Payment

	if existingPayment != nil {
		// Use existing payment if it's in pending status
		if existingPayment.Status != "pending" {
			return nil, fmt.Errorf("payment already processed for this booking")
		}
		payment = existingPayment
	} else {
		// Create new payment
		payment = &model.Payment{
			UserID:        userID,
			BookingID:     req.BookingID,
			Amount:        req.Amount,
			Currency:      req.Currency,
			Status:        "pending",
			PaymentMethod: req.PaymentMethod,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		// Save payment to database
		if err := s.paymentRepo.Create(payment); err != nil {
			return nil, fmt.Errorf("failed to create payment: %w", err)
		}
	}

	// Process payment with payment provider
	transactionID, err := s.paymentProvider.ProcessPayment(req)
	if err != nil {
		// Update payment status to failed
		payment.Status = "failed"
		payment.UpdatedAt = time.Now()

		// Save payment to database
		if updateErr := s.paymentRepo.Update(payment); updateErr != nil {
			logrus.WithError(updateErr).Error("Failed to update payment status to failed")
		}

		// Publish payment failed event
		s.publishPaymentEvent("payment.failed", payment)

		return nil, fmt.Errorf("failed to process payment: %w", err)
	}

	// Update payment status to completed
	payment.Status = "completed"
	payment.TransactionID = transactionID
	payment.PaymentDate = time.Now()
	payment.UpdatedAt = time.Now()

	// Save payment to database
	if err := s.paymentRepo.Update(payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Publish payment completed event
	s.publishPaymentEvent("payment.completed", payment)

	// Return payment response
	paymentResponse := payment.ToResponse()
	return &paymentResponse, nil
}

// GetPaymentByID gets a payment by ID
func (s *paymentService) GetPaymentByID(id uuid.UUID) (*model.PaymentResponse, error) {
	// Find payment by ID
	payment, err := s.paymentRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	if payment == nil {
		return nil, fmt.Errorf("payment not found")
	}

	// Return payment response
	paymentResponse := payment.ToResponse()
	return &paymentResponse, nil
}

// GetPaymentByBookingID gets a payment by booking ID
func (s *paymentService) GetPaymentByBookingID(bookingID uuid.UUID) (*model.PaymentResponse, error) {
	// Find payment by booking ID
	payment, err := s.paymentRepo.FindByBookingID(bookingID)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	if payment == nil {
		return nil, fmt.Errorf("payment not found")
	}

	// Return payment response
	paymentResponse := payment.ToResponse()
	return &paymentResponse, nil
}

// GetUserPayments gets payments for a user
func (s *paymentService) GetUserPayments(userID uuid.UUID, page, pageSize int) ([]model.PaymentResponse, int64, error) {
	// Find payments by user ID
	payments, total, err := s.paymentRepo.FindByUserID(userID, page, pageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to find payments: %w", err)
	}

	// Convert payments to responses
	responses := make([]model.PaymentResponse, len(payments))
	for i, payment := range payments {
		responses[i] = payment.ToResponse()
	}

	return responses, total, nil
}

// UpdatePaymentStatus updates a payment status
func (s *paymentService) UpdatePaymentStatus(id uuid.UUID, req model.UpdatePaymentStatusRequest) (*model.PaymentResponse, error) {
	// Find payment by ID
	payment, err := s.paymentRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	if payment == nil {
		return nil, fmt.Errorf("payment not found")
	}

	// Update payment status
	payment.Status = req.Status
	payment.UpdatedAt = time.Now()

	// Update transaction ID if provided
	if req.TransactionID != "" {
		payment.TransactionID = req.TransactionID
	}

	// Update payment date if provided
	if !req.PaymentDate.IsZero() {
		payment.PaymentDate = req.PaymentDate
	}

	// Save payment to database
	if err := s.paymentRepo.Update(payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Publish payment updated event
	s.publishPaymentEvent("payment.updated", payment)

	// Return payment response
	paymentResponse := payment.ToResponse()
	return &paymentResponse, nil
}

// RefundPayment refunds a payment
func (s *paymentService) RefundPayment(id uuid.UUID, req model.RefundRequest) (*model.PaymentResponse, error) {
	// Find payment by ID
	payment, err := s.paymentRepo.FindByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to find payment: %w", err)
	}

	if payment == nil {
		return nil, fmt.Errorf("payment not found")
	}

	// Check if payment is completed
	if payment.Status != "completed" {
		return nil, fmt.Errorf("only completed payments can be refunded")
	}

	// Process refund with payment provider
	err = s.paymentProvider.RefundPayment(payment.TransactionID, req.Reason)
	if err != nil {
		return nil, fmt.Errorf("failed to process refund: %w", err)
	}

	// Update payment status to refunded
	payment.Status = "refunded"
	payment.UpdatedAt = time.Now()

	// Save payment to database
	if err := s.paymentRepo.Update(payment); err != nil {
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	// Publish payment refunded event
	s.publishPaymentEvent("payment.refunded", payment)

	// Return payment response
	paymentResponse := payment.ToResponse()
	return &paymentResponse, nil
}

// publishPaymentEvent publishes a payment event to RabbitMQ
func (s *paymentService) publishPaymentEvent(eventType string, payment *model.Payment) {
	// Create event payload
	payload := map[string]interface{}{
		"event_type":     eventType,
		"payment_id":     payment.ID.String(),
		"user_id":        payment.UserID.String(),
		"booking_id":     payment.BookingID.String(),
		"amount":         payment.Amount,
		"currency":       payment.Currency,
		"status":         payment.Status,
		"payment_method": payment.PaymentMethod,
		"timestamp":      time.Now(),
	}

	// Convert payload to JSON
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal payment event")
		return
	}

	// Publish event to RabbitMQ
	err = s.rmq.PublishMessage("payment_events", eventType, payloadJSON)
	if err != nil {
		logrus.WithError(err).Error("Failed to publish payment event")
	}
}