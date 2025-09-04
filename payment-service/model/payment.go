package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Payment represents a payment transaction
type Payment struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key" json:"id"`
	UserID        uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	BookingID     uuid.UUID `gorm:"type:uuid;index" json:"booking_id"`
	Amount        float64   `gorm:"type:decimal(10,2)" json:"amount"`
	Currency      string    `gorm:"type:varchar(3)" json:"currency"`
	Status        string    `gorm:"type:varchar(20)" json:"status"` // pending, completed, failed, refunded
	PaymentMethod string    `gorm:"type:varchar(50)" json:"payment_method"`
	TransactionID string    `gorm:"type:varchar(100)" json:"transaction_id"`
	PaymentDate   time.Time `json:"payment_date"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// BeforeCreate will set a UUID rather than numeric ID
func (p *Payment) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// PaymentResponse represents the response for a payment
type PaymentResponse struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	BookingID     uuid.UUID `json:"booking_id"`
	Amount        float64   `json:"amount"`
	Currency      string    `json:"currency"`
	Status        string    `json:"status"`
	PaymentMethod string    `json:"payment_method"`
	TransactionID string    `json:"transaction_id,omitempty"`
	PaymentDate   time.Time `json:"payment_date,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreatePaymentRequest represents the request to create a payment
type CreatePaymentRequest struct {
	BookingID     uuid.UUID `json:"booking_id" binding:"required"`
	Amount        float64   `json:"amount" binding:"required"`
	Currency      string    `json:"currency" binding:"required"`
	PaymentMethod string    `json:"payment_method" binding:"required"`
}

// UpdatePaymentStatusRequest represents the request to update a payment status
type UpdatePaymentStatusRequest struct {
	Status        string    `json:"status" binding:"required"`
	TransactionID string    `json:"transaction_id,omitempty"`
	PaymentDate   time.Time `json:"payment_date,omitempty"`
}

// ProcessPaymentRequest represents the request to process a payment
type ProcessPaymentRequest struct {
	BookingID     uuid.UUID `json:"booking_id" binding:"required"`
	Amount        float64   `json:"amount" binding:"required"`
	Currency      string    `json:"currency" binding:"required"`
	PaymentMethod string    `json:"payment_method" binding:"required"`
	CardNumber    string    `json:"card_number,omitempty"`
	CardExpiry    string    `json:"card_expiry,omitempty"`
	CardCVC       string    `json:"card_cvc,omitempty"`
	CardHolder    string    `json:"card_holder,omitempty"`
}

// RefundRequest represents the request to refund a payment
type RefundRequest struct {
	Reason string `json:"reason,omitempty"`
}

// ToResponse converts a Payment to a PaymentResponse
func (p *Payment) ToResponse() PaymentResponse {
	return PaymentResponse{
		ID:            p.ID,
		UserID:        p.UserID,
		BookingID:     p.BookingID,
		Amount:        p.Amount,
		Currency:      p.Currency,
		Status:        p.Status,
		PaymentMethod: p.PaymentMethod,
		TransactionID: p.TransactionID,
		PaymentDate:   p.PaymentDate,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}